package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autopilothub/zeroflight/internal/safety"
	"github.com/spf13/cobra"
)

func newPreflightCmd() *cobra.Command {
	var requirePass bool

	cmd := &cobra.Command{
		Use:   "preflight",
		Short: "Run autonomous flight preflight checklist",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			client, fileCfg, err := loadClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if err := client.WaitForConnection(ctx, 10*time.Second); err != nil {
				return err
			}

			state := client.State()
			limits := safety.LimitsFromConfig(fileCfg.Safety)
			results := safety.RunPreflight(state, limits, fileCfg.LinkTimeout())

			fmt.Println("ZeroFlight Preflight Checklist")
			fmt.Println("================================")
			for _, r := range results {
				status := "FAIL"
				if r.Passed {
					status = " OK "
				}
				fmt.Printf("[%s] %s — %s\n", status, r.Name, r.Message)
			}
			fmt.Println()

			if safety.AllPassed(results) {
				fmt.Println("Result: PASS — ready for autonomous commands")
				return nil
			}

			fmt.Println("Result: FAIL — resolve items above before goto/mission")
			if requirePass {
				return fmt.Errorf("preflight failed")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&requirePass, "require-pass", false, "exit with error if any check fails")
	return cmd
}
