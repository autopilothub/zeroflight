package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autopilothub/zeroflight/internal/config"
	"github.com/autopilothub/zeroflight/internal/inav"
	"github.com/autopilothub/zeroflight/pkg/geo"
	"github.com/spf13/cobra"
)

var (
	cfgPath    string
	connection string
)

func main() {
	root := &cobra.Command{
		Use:   "zeroflight",
		Short: "INAV autonomous drone companion for Raspberry Pi",
	}

	root.PersistentFlags().StringVar(&cfgPath, "config", "configs/inav.yaml", "path to config file")
	root.PersistentFlags().StringVar(&connection, "connection", "", "override connection (serial:/dev/ttyAMA0:115200 or udp:host:port)")

	root.AddCommand(newStatusCmd())
	root.AddCommand(newGotoCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadClient(ctx context.Context) (*inav.Client, config.File, error) {
	fileCfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fileCfg, err
	}

	clientCfg, err := fileCfg.INAVConfig(connection)
	if err != nil {
		return nil, fileCfg, err
	}

	client := inav.NewClient(clientCfg)
	if err := client.Connect(ctx); err != nil {
		return nil, fileCfg, fmt.Errorf("connect: %w", err)
	}
	return client, fileCfg, nil
}

func newStatusCmd() *cobra.Command {
	var interval time.Duration
	var once bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Stream INAV telemetry (attitude, GPS, mode)",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			client, _, err := loadClient(ctx)
			if err != nil {
				return err
			}
			defer client.Close()

			if err := client.WaitForConnection(ctx, 10*time.Second); err != nil {
				return err
			}

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			printStatus(client.State())
			if once {
				return nil
			}

			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					printStatus(client.State())
				}
			}
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 500*time.Millisecond, "refresh interval")
	cmd.Flags().BoolVar(&once, "once", false, "print one snapshot and exit")
	return cmd
}

func printStatus(state inav.VehicleState) {
	armed := "disarmed"
	if state.Armed {
		armed = "armed"
	}
	gcsNav := "off"
	if state.GCSNavActive {
		gcsNav = "on"
	}

	fmt.Printf("\033[H\033[J")
	fmt.Printf("ZeroFlight INAV Telemetry  %s\n", state.Time.Format("15:04:05"))
	fmt.Printf("Connected: %v  %s  Mode: %s  GCS NAV: %s\n\n",
		state.Connected, armed, state.Mode, gcsNav)

	fmt.Printf("Attitude (rad)\n")
	fmt.Printf("  roll=%.3f pitch=%.3f yaw=%.3f\n", state.Attitude.Roll, state.Attitude.Pitch, state.Attitude.Yaw)
	fmt.Printf("  rates: roll=%.2f pitch=%.2f yaw=%.2f rad/s\n",
		state.Attitude.RollSpeed, state.Attitude.PitchSpeed, state.Attitude.YawSpeed)

	fmt.Printf("\nGPS\n")
	fmt.Printf("  lat=%.7f lon=%.7f alt=%.1fm rel=%.1fm\n",
		state.GPS.Lat, state.GPS.Lon, state.GPS.AltM, state.GPS.RelAltM)
	fmt.Printf("  fix=%d sats=%d hdop=%.1f speed=%.1fm/s climb=%.1fm/s\n",
		state.GPS.FixType, state.GPS.Satellites, state.GPS.HDOP, state.GPS.GroundSpeed, state.GPS.ClimbRate)

	fmt.Printf("\nBattery\n")
	fmt.Printf("  %.2fV %.1fA %d%%\n", state.Battery.VoltageV, state.Battery.CurrentA, state.Battery.RemainingPct)

	fmt.Printf("\nSensors\n")
	fmt.Printf("  gyro=%v accel=%v mag=%v baro=%v gps=%v rc=%v\n",
		state.Sensors.Gyro, state.Sensors.Accel, state.Sensors.Mag,
		state.Sensors.Baro, state.Sensors.GPS, state.Sensors.RC)
}

func newGotoCmd() *cobra.Command {
	var lat, lon float64
	var alt float32
	var yaw float32
	var useYaw bool
	var wait bool
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "goto",
		Short: "Send MAV_CMD_DO_REPOSITION (requires GCS NAV mode)",
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
			if err := inav.PreflightGoto(state); err != nil {
				return err
			}
			if fileCfg.Safety.MaxAltitudeM > 0 && alt > fileCfg.Safety.MaxAltitudeM {
				return fmt.Errorf("altitude %.1fm exceeds max %.1fm", alt, fileCfg.Safety.MaxAltitudeM)
			}

			req := inav.GotoRequest{
				LatDeg: lat,
				LonDeg: lon,
				AltM:   alt,
			}
			if useYaw {
				req.YawDeg = &yaw
			}

			if err := client.SendGoto(req); err != nil {
				return err
			}
			fmt.Printf("goto sent: lat=%.7f lon=%.7f alt=%.1fm\n", lat, lon, alt)

			if !wait {
				return nil
			}

			deadline := time.Now().Add(timeout)
			for time.Now().Before(deadline) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(500 * time.Millisecond):
				}

				cur := client.State()
				horiz := geo.DistanceM(cur.GPS.Lat, cur.GPS.Lon, lat, lon)
				altErr := float32(math.Abs(float64(cur.GPS.RelAltM - alt)))
				fmt.Printf("  distance=%.1fm alt_err=%.1fm mode=%s\n", horiz, altErr, cur.Mode)

				if horiz <= fileCfg.Safety.ArrivalRadiusM &&
					altErr <= fileCfg.Safety.ArrivalAltitudeM {
					fmt.Println("arrived at target")
					return nil
				}
			}
			return fmt.Errorf("timed out waiting for arrival")
		},
	}

	cmd.Flags().Float64Var(&lat, "lat", 0, "target latitude (degrees)")
	cmd.Flags().Float64Var(&lon, "lon", 0, "target longitude (degrees)")
	cmd.Flags().Float32Var(&alt, "alt", 10, "target altitude (meters, relative)")
	cmd.Flags().Float32Var(&yaw, "yaw", 0, "target yaw (degrees, 1-359)")
	cmd.Flags().BoolVar(&useYaw, "set-yaw", false, "set explicit yaw via param4")
	cmd.Flags().BoolVar(&wait, "wait", false, "wait until arrival radius is reached")
	cmd.Flags().DurationVar(&timeout, "timeout", 3*time.Minute, "arrival wait timeout")
	_ = cmd.MarkFlagRequired("lat")
	_ = cmd.MarkFlagRequired("lon")
	return cmd
}
