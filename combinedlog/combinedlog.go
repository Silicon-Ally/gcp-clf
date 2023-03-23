package combinedlog

import (
	"fmt"
	"strconv"
	"time"
)

type RequestEntry struct {
	RemoteAddr   string
	Identity     string
	User         string
	RequestedAt  time.Time
	Method       string
	Path         string
	Protocol     string
	StatusCode   int
	ResponseSize *int64
	Referer      string
	UserAgent    string
}

func (r RequestEntry) String() string {
	return fmt.Sprintf("%s %s %s [%s] %q %d %s %s %q",
		r.RemoteAddr,
		emptyAsHyphen(r.Identity),
		emptyAsHyphen(r.User),
		r.formattedTime(),
		r.requestInfo(),
		r.StatusCode,
		r.responseSize(),
		emptyAsHyphen(r.Referer),
		r.UserAgent)
}

func (r RequestEntry) formattedTime() string {
	return r.RequestedAt.Format("02/Jan/2006:15:04:05 -0700")
}

func (r RequestEntry) requestInfo() string {
	return fmt.Sprintf("%s %s %s", r.Method, r.Path, r.Protocol)
}

func (r RequestEntry) responseSize() string {
	if r.ResponseSize == nil {
		return "-"
	}
	return strconv.FormatInt(*r.ResponseSize, 10)
}

func emptyAsHyphen(in string) string {
	if in == "" {
		return "-"
	}
	return in
}