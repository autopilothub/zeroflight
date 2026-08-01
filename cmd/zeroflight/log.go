package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	telemetrylog "github.com/autopilothub/zeroflight/internal/log"
	"github.com/spf13/cobra"
)

func newLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log",
		Short: "Record telemetry to files",
	}
	cmd.AddCommand(newLogTelemetryCmd())
	return cmd
}

func newLogTelemetryCmd() *cobra.Command {
	var output string
	var interval time.Duration
	var duration time.Duration

	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Write telemetry CSV until duration elapses or Ctrl+C",
		RunE: func(cmd *cobra.Command, args []string) error {
			if output == "" {
				output = "logs/telemetry.csv"
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			sess, err := loadSession(ctx)
			if err != nil {
				return err
			}
			defer sess.Close()

			if err := sess.WaitReady(ctx); err != nil {
				return err
			}

			logger := telemetrylog.NewTelemetryCSV(output)
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			deadline := time.Now().Add(duration)
			fmt.Printf("logging telemetry to %s (interval %s)\n", output, interval)

			for {
				state := sess.State()
				if err := logger.Write(state); err != nil {
					return err
				}

				if duration > 0 && time.Now().After(deadline) {
					fmt.Println("logging complete")
					return nil
				}

				select {
				case <-ctx.Done():
					fmt.Println("logging stopped")
					return nil
				case <-ticker.C:
				}
			}
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "logs/telemetry.csv", "CSV output path")
	cmd.Flags().DurationVar(&interval, "interval", time.Second, "sample interval")
	cmd.Flags().DurationVar(&duration, "duration", 0, "stop after duration (0 = until Ctrl+C)")
	return cmd
}
