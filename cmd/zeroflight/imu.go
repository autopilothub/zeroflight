package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/spf13/cobra"
)

func newImuCmd() *cobra.Command {
	var interval time.Duration
	var once bool

	cmd := &cobra.Command{
		Use:   "imu",
		Short: "Stream raw IMU from MSP (requires msp.enabled in config)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			sess, err := loadSession(ctx)
			if err != nil {
				return err
			}
			defer sess.Close()

			if !sess.Config().MSP.Enabled {
				return fmt.Errorf("msp is disabled; set msp.enabled: true in %s", cfgPath)
			}

			if err := sess.WaitReady(ctx); err != nil {
				return err
			}

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			printIMU(sess.State().RawIMU)
			if once {
				return nil
			}

			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					printIMU(sess.State().RawIMU)
				}
			}
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 100*time.Millisecond, "refresh interval")
	cmd.Flags().BoolVar(&once, "once", false, "print one sample and exit")
	return cmd
}

func printIMU(imu inav.RawIMU) {
	fmt.Printf("\033[H\033[J")
	fmt.Printf("ZeroFlight Raw IMU (MSP)  %s\n\n", time.Now().Format("15:04:05"))
	if !imu.Available {
		fmt.Println("waiting for MSP_RAW_IMU...")
		return
	}
	fmt.Printf("accel: %+v\n", imu.Accel)
	fmt.Printf("gyro:  %+v\n", imu.Gyro)
	fmt.Printf("mag:   %+v\n", imu.Mag)
	fmt.Printf("\nupdated %s\n", imu.Time.Format("15:04:05.000"))
}
