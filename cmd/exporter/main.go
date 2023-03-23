package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Silicon-Ally/gcp-clf/combinedlog"
	"github.com/Silicon-Ally/gcp-clf/gcplog"
	"github.com/namsral/flag"
)

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("args cannot be empty")
	}
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		startTime = fs.String("start_time", "", time.Now().Add(-7*24*time.Hour).Format(time.RFC3339))
		endTime   = fs.String("end_time", "", "")
	)
	// Allows for passing in configuration via a -config path/to/env-file.conf
	// flag, see https://pkg.go.dev/github.com/namsral/flag#readme-usage
	fs.String(flag.DefaultConfigFlagname, "", "path to config file")
	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %v", err)
	}

	ctx := context.Background()
	switch fs.Arg(0) {
	case "logs":
		var opts []gcplog.Option
		if *startTime != "" {
			t, err := time.Parse(*startTime, time.RFC3339)
			if err != nil {
				return fmt.Errorf("failed to parse --start_time: %w", err)
			}
			opts = append(opts, gcplog.WithStartTime(t))
		}
		if *endTime != "" {
			t, err := time.Parse(*endTime, time.RFC3339)
			if err != nil {
				return fmt.Errorf("failed to parse --end_time: %w", err)
			}
			opts = append(opts, gcplog.WithEndTime(t))
		}
		logs, err := gcplog.GetLogs(ctx, fs.Arg(1), opts...)
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}

		for _, log := range logs {
			fmt.Println(log.String())
		}
	case "stream":
		err := gcplog.StreamLogs(ctx, fs.Arg(1), func(rl combinedlog.RequestEntry) {
			fmt.Println(rl.String())
		})
		if err != nil {
			return fmt.Errorf("error while streaming logs: %w", err)
		}
	}

	return nil
}
