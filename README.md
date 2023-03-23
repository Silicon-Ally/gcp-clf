# GCP Cloud Logging -> Combined Log Format

This repo contains utilities for exporting request logs from Cloud Logging to the Combined Log Format.

The Combined Log Format is just the [Common Log Format](https://en.wikipedia.org/wiki/Common_Log_Format) with two additional fields to record the [Referer](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referer) and [User-Agent](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent) of a request. This is the log format commonly used by [NGINX](https://nginx.org/en/docs/http/ngx_http_log_module.html) and [Apache](https://httpd.apache.org/docs/2.4/logs.html#combined) web servers, and so is supported by log processing tools like [GoAccess](https://goaccess.io/)

Currently, this tool supports extracting logs specifically from [Firebase Hosting](https://firebase.google.com/docs/hosting/)). To use this tool with Firebase hosting logs, make sure to [link Cloud Logging](https://firebase.google.com/docs/hosting/web-request-logs-and-metrics#link-and-monitor).

## Usage with GoAccess

To view logs from your server over a time period with [GoAccess](https://goaccess.io/), run:

```bash
go run ./cmd/exporter logs <project ID> \
  --start_time=<RFC3339 time> \
  --end_time=<RFC3339 time> > access.log

# For terminal UI
goaccess --log-format=combined access.log

# For web UI
goaccess --log-format=combined access.log -o report.html
# Then open `report.html` in your web browser of choice.
```

To view streaming logs, run:

```bash
# For terminal UI
go run ./cmd/exporter stream <project ID> | goaccess --log-format=COMBINED -

# For web UI
go run ./cmd/exporter stream <project ID> \
  | goaccess --log-format=COMBINED --real-time-html -o report.html -
# Then open `report.html` in your web browser of choice.
```
