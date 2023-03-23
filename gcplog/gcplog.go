package gcplog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/Silicon-Ally/gcp-clf/combinedlog"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
)

type Option func(*options)

type options struct {
	startTime *time.Time
	endTime   *time.Time
}

func WithStartTime(t time.Time) Option {
	return func(o *options) {
		o.startTime = &t
	}
}

func WithEndTime(t time.Time) Option {
	return func(o *options) {
		o.endTime = &t
	}
}

func StreamLogs(ctx context.Context, projID string, fn func(combinedlog.RequestEntry)) error {
	c, err := logging.NewClient(ctx, option.WithQuotaProject((projID)))
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer c.Close()

	stream, err := c.TailLogEntries(ctx)
	if err != nil {
		return fmt.Errorf("failed to tail log entries: %w", err)
	}

	filters := []string{
		`resource.type="firebase_domain"`,
		fmt.Sprintf(`resource.labels.site_name=%q`, projID),
	}

	errChan := make(chan error)
	go func() {
		err = stream.Send(&loggingpb.TailLogEntriesRequest{
			ResourceNames: []string{
				"projects/" + projID,
			},
			Filter: strings.Join(filters, " AND "),
		})
		if err != nil {
			errChan <- fmt.Errorf("failed to send tail request: %w", err)
		}
	}()

	logChan := make(chan []*loggingpb.LogEntry)
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				errChan <- err
				break
			}

			logChan <- resp.Entries
		}
	}()

	for {
		select {
		case logs := <-logChan:
			for _, l := range logs {
				rl, err := toRequestEntry(l)
				if err != nil {
					return fmt.Errorf("failed to convert log entry: %w", err)
				}
				fn(rl)
			}
		case err := <-errChan:
			if errors.Is(err, io.EOF) {
				return nil
			} else if err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
			}
		}
	}
}

func GetLogs(ctx context.Context, projID string, logOpts ...Option) ([]combinedlog.RequestEntry, error) {
	o := &options{}
	for _, opt := range logOpts {
		opt(o)
	}

	c, err := logging.NewClient(ctx, option.WithQuotaProject((projID)))
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}
	defer c.Close()

	callOpts := []gax.CallOption{
		gax.WithRetry(func() gax.Retryer {
			return gax.OnCodes([]codes.Code{
				codes.ResourceExhausted,
			}, gax.Backoff{})
		}),
	}

	filters := []string{
		`resource.type="firebase_domain"`,
		fmt.Sprintf(`resource.labels.site_name=%q`, projID),
	}
	if o.startTime != nil {
		filters = append(filters, fmt.Sprintf(`timestamp>=%q`, o.startTime.Format(time.RFC3339)))
	}
	if o.endTime != nil {
		filters = append(filters, fmt.Sprintf(`timestamp<=%q`, o.endTime.Format(time.RFC3339)))
	}

	iter := c.ListLogEntries(ctx, &loggingpb.ListLogEntriesRequest{
		ResourceNames: []string{
			"projects/" + projID,
		},
		Filter: strings.Join(filters, " AND "),
		// The idea here is that quotas for Cloud Logging are fairly low by default (60
		// read reqs/minute), so if you have a decent amount of logs, you'll hit quota
		// fairly quickly.
		PageSize: 1000,
	}, callOpts...)

	var out []combinedlog.RequestEntry
	for {
		entry, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to load log entry: %w", err)
		}

		rEntry, err := toRequestEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("failed to convert log entry: %w", err)
		}
		out = append(out, rEntry)
	}

	return out, nil
}

func toRequestEntry(entry *loggingpb.LogEntry) (combinedlog.RequestEntry, error) {
	// We don't currently use anything in the payload.
	// payload, ok := entry.Payload.(*loggingpb.LogEntry_JsonPayload)
	// if !ok {
	// 	return nil, fmt.Errorf("log entry payload was %T, expected JSON", entry.Payload)
	// }

	httpReq := entry.HttpRequest
	reqUrl, err := url.Parse(httpReq.RequestUrl)
	if err != nil {
		return combinedlog.RequestEntry{}, fmt.Errorf("failed to parse request url: %w", err)
	}
	if reqUrl == nil {
		return combinedlog.RequestEntry{}, errors.New("no URL in log request")
	}
	return combinedlog.RequestEntry{
		RemoteAddr:   httpReq.RemoteIp,
		Identity:     "",
		User:         "",
		RequestedAt:  entry.Timestamp.AsTime(),
		Method:       httpReq.RequestMethod,
		Path:         reqUrl.Path,
		Protocol:     httpReq.Protocol,
		StatusCode:   int(httpReq.Status),
		ResponseSize: &httpReq.ResponseSize,
		Referer:      httpReq.Referer,
		UserAgent:    httpReq.UserAgent,
	}, nil
}