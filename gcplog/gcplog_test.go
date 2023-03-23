package gcplog

import (
	"net/http"
	"testing"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/Silicon-Ally/gcp-clf/combinedlog"
	"github.com/google/go-cmp/cmp"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestToRequestEntry(t *testing.T) {
	ts := time.Date(2020, time.February, 23, 1, 2, 3, 4567, time.UTC)

	in := &loggingpb.LogEntry{
		// No other fields are currently used
		Timestamp: timestamppb.New(ts),
		HttpRequest: &ltype.HttpRequest{
			// Only used fields are currently populated
			RequestMethod: http.MethodPost,
			RequestUrl:    "https://www.example.com/favicon.ico",
			Status:        http.StatusOK,
			ResponseSize:  12345,
			UserAgent:     "GoogleBot",
			RemoteIp:      "1.2.3.4",
			Referer:       "sub.example.com",
			Protocol:      "https",
		},
	}

	got, err := toRequestEntry(in)
	if err != nil {
		t.Fatalf("toRequestEntry: %v", err)
	}

	want := combinedlog.RequestEntry{
		RemoteAddr:   "1.2.3.4",
		Identity:     "",
		User:         "",
		RequestedAt:  ts,
		Method:       http.MethodPost,
		Path:         "/favicon.ico",
		Protocol:     "https",
		StatusCode:   http.StatusOK,
		ResponseSize: ptr(int64(12345)),
		Referer:      "sub.example.com",
		UserAgent:    "GoogleBot",
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("unexpected request entry (-want +got)\n%s", diff)
	}
}

func ptr[T any](in T) *T {
	return &in
}
