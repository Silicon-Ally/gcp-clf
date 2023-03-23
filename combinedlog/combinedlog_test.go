package combinedlog

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestRequestEntryString(t *testing.T) {
	tests := []struct {
		desc string
		in   RequestEntry
		want string
	}{
		{
			desc: "standard log",
			in: RequestEntry{
				RemoteAddr:  "1.2.3.4",
				Identity:    "",
				User:        "",
				RequestedAt: time.Date(2020, time.February, 23, 10, 11, 12, 1234325, time.UTC),
				Method:      "GET",
				Path:        "/test",
				// Normally, one would expect to see HTTP/1.0, HTTP/1.1, HTTP/2 etc, but I've
				// only ever seen 'https' in the logs, despite the documentation:
				// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
				Protocol:     "https",
				StatusCode:   200,
				ResponseSize: ptr(int64(12345)),
				Referer:      "example.com",
				UserAgent:    "Firefox 123.3.4",
			},
			want: `1.2.3.4 - - [23/Feb/2020:10:11:12 +0000] "GET /test https" 200 12345 example.com "Firefox 123.3.4"`,
		},
		{
			desc: "with user",
			in: RequestEntry{
				RemoteAddr: "1.2.3.4",
				Identity:   "",
				// All the examples seem to use frank for some reason that I have not dug in to.
				User:        "frank",
				RequestedAt: time.Date(2020, time.February, 23, 10, 11, 12, 1234325, time.UTC),
				Method:      "GET",
				Path:        "/test",
				// Normally, one would expect to see HTTP/1.0, HTTP/1.1, HTTP/2 etc, but I've
				// only ever seen 'https' in the logs, despite the documentation:
				// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
				Protocol:     "https",
				StatusCode:   200,
				ResponseSize: ptr(int64(12345)),
				Referer:      "example.com",
				UserAgent:    "Firefox 123.3.4",
			},
			want: `1.2.3.4 - frank [23/Feb/2020:10:11:12 +0000] "GET /test https" 200 12345 example.com "Firefox 123.3.4"`,
		},
		{
			desc: "no response size",
			in: RequestEntry{
				RemoteAddr: "1.2.3.4",
				Identity:   "",
				// All the examples seem to use frank for some reason that I have not dug in to.
				User:        "frank",
				RequestedAt: time.Date(2020, time.February, 23, 10, 11, 12, 1234325, time.UTC),
				Method:      "GET",
				Path:        "/test",
				// Normally, one would expect to see HTTP/1.0, HTTP/1.1, HTTP/2 etc, but I've
				// only ever seen 'https' in the logs, despite the documentation:
				// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
				Protocol:     "https",
				StatusCode:   200,
				ResponseSize: nil,
				Referer:      "example.com",
				UserAgent:    "Firefox 123.3.4",
			},
			want: `1.2.3.4 - frank [23/Feb/2020:10:11:12 +0000] "GET /test https" 200 - example.com "Firefox 123.3.4"`,
		},
		{
			desc: "no referer",
			in: RequestEntry{
				RemoteAddr:  "1.2.3.4",
				Identity:    "",
				User:        "",
				RequestedAt: time.Date(2020, time.February, 23, 10, 11, 12, 1234325, time.UTC),
				Method:      "GET",
				Path:        "/test",
				// Normally, one would expect to see HTTP/1.0, HTTP/1.1, HTTP/2 etc, but I've
				// only ever seen 'https' in the logs, despite the documentation:
				// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
				Protocol:     "https",
				StatusCode:   200,
				ResponseSize: ptr(int64(12345)),
				Referer:      "",
				UserAgent:    "Firefox 123.3.4",
			},
			want: `1.2.3.4 - - [23/Feb/2020:10:11:12 +0000] "GET /test https" 200 12345 - "Firefox 123.3.4"`,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			got := test.in.String()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected log output (-want +got)\n%s", diff)
			}
		})
	}
}

func ptr[T any](in T) *T {
	return &in
}
