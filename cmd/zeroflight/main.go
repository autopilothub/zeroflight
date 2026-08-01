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
	"github.com/autopilothub/zeroflight/internal/safety"
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
	root.PersistentFlags().StringVar(&connection, "connection", "", "override connection (serial:/dev/serial0:115200 or udp:host:port)")

	root.AddCommand(newStatusCmd())
	root.AddCommand(newGotoCmd())
	root.AddCommand(newHoverCmd())
	root.AddCommand(newMissionCmd())
	root.AddCommand(newPreflightCmd())
	root.AddCommand(newLogCmd())
	root.AddCommand(newServeCmd())

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

	fmt.Printf("\nHome\n")
	if state.Home.Valid {
		fmt.Printf("  lat=%.7f lon=%.7f alt=%.1fm\n", state.Home.Lat, state.Home.Lon, state.Home.AltM)
	} else {
		fmt.Printf("  (waiting for GPS_GLOBAL_ORIGIN)\n")
	}
}

type navCommandOptions struct {
	command   string
	lat       float64
	lon       float64
	alt       float32
	useYaw    bool
	yaw       float32
	wait      bool
	timeout   time.Duration
	force     bool
	logPath   string
	useGPSAlt bool
}

func runNavCommand(opts navCommandOptions) error {
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
	if safety.IsLinkStale(state, fileCfg.LinkTimeout()) {
		return fmt.Errorf("mavlink link stale; last telemetry %s ago", time.Since(state.Time).Round(time.Second))
	}

	preflight, err := inav.CheckGotoPreflight(state, inav.GotoPreflightOptions{Force: opts.force})
	if err != nil {
		return err
	}
	for _, warning := range preflight.Warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", warning)
	}

	lat := opts.lat
	lon := opts.lon
	alt := opts.alt

	if opts.useGPSAlt {
		if state.GPS.Lat == 0 && state.GPS.Lon == 0 {
			return fmt.Errorf("no valid GPS position for hover")
		}
		lat = state.GPS.Lat
		lon = state.GPS.Lon
		if alt == 0 {
			alt = state.GPS.RelAltM
		}
	}

	limits := safety.LimitsFromConfig(fileCfg.Safety)
	if err := safety.ValidateTarget(state.Home, lat, lon, alt, limits); err != nil {
		return err
	}

	req := inav.GotoRequest{
		LatDeg: lat,
		LonDeg: lon,
		AltM:   alt,
	}
	if opts.useYaw {
		req.YawDeg = &opts.yaw
	}

	if err := client.SendGoto(req); err != nil {
		return err
	}

	logger := inav.NewGotoAuditLogger(opts.logPath)
	if err := logger.Log(inav.GotoAuditEntry{
		Command:    opts.command,
		LatDeg:     lat,
		LonDeg:     lon,
		AltM:       alt,
		Mode:       state.Mode,
		Armed:      state.Armed,
		FixType:    state.GPS.FixType,
		Satellites: state.GPS.Satellites,
		HDOP:       state.GPS.HDOP,
		Forced:     opts.force,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write audit log: %v\n", err)
	}

	fmt.Printf("%s sent: lat=%.7f lon=%.7f alt=%.1fm\n", opts.command, lat, lon, alt)

	if !opts.wait {
		return nil
	}

	deadline := time.Now().Add(opts.timeout)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}

		cur := client.State()
		if safety.IsLinkStale(cur, fileCfg.LinkTimeout()) {
			return fmt.Errorf("mavlink link lost during %s", opts.command)
		}
		if !cur.GCSNavActive {
			return fmt.Errorf("GCS NAV deactivated during %s", opts.command)
		}

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
}

func newGotoCmd() *cobra.Command {
	opts := navCommandOptions{command: "goto"}

	cmd := &cobra.Command{
		Use:   "goto",
		Short: "Send MAV_CMD_DO_REPOSITION (requires GCS NAV mode)",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.lat = gotoLat
			opts.lon = gotoLon
			opts.alt = gotoAlt
			opts.useYaw = gotoUseYaw
			opts.yaw = gotoYaw
			opts.wait = gotoWait
			opts.timeout = gotoTimeout
			opts.force = gotoForce
			opts.logPath = gotoLogPath
			return runNavCommand(opts)
		},
	}

	cmd.Flags().Float64Var(&gotoLat, "lat", 0, "target latitude (degrees)")
	cmd.Flags().Float64Var(&gotoLon, "lon", 0, "target longitude (degrees)")
	cmd.Flags().Float32Var(&gotoAlt, "alt", 10, "target altitude (meters, relative)")
	cmd.Flags().Float32Var(&gotoYaw, "yaw", 0, "target yaw (degrees, 1-359)")
	cmd.Flags().BoolVar(&gotoUseYaw, "set-yaw", false, "set explicit yaw via param4")
	cmd.Flags().BoolVar(&gotoWait, "wait", false, "wait until arrival radius is reached")
	cmd.Flags().DurationVar(&gotoTimeout, "timeout", 3*time.Minute, "arrival wait timeout")
	cmd.Flags().BoolVar(&gotoForce, "force", false, "proceed despite high HDOP")
	cmd.Flags().StringVar(&gotoLogPath, "log", "", "goto audit log path (default logs/goto.jsonl)")
	_ = cmd.MarkFlagRequired("lat")
	_ = cmd.MarkFlagRequired("lon")
	return cmd
}

func newHoverCmd() *cobra.Command {
	opts := navCommandOptions{command: "hover", useGPSAlt: true}

	cmd := &cobra.Command{
		Use:   "hover",
		Short: "Hold current GPS position (optional altitude change)",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.alt = hoverAlt
			opts.wait = hoverWait
			opts.timeout = hoverTimeout
			opts.force = hoverForce
			opts.logPath = hoverLogPath
			return runNavCommand(opts)
		},
	}

	cmd.Flags().Float32Var(&hoverAlt, "alt", 0, "target altitude (m, relative); 0 keeps current")
	cmd.Flags().BoolVar(&hoverWait, "wait", false, "wait until arrival radius is reached")
	cmd.Flags().DurationVar(&hoverTimeout, "timeout", 2*time.Minute, "arrival wait timeout")
	cmd.Flags().BoolVar(&hoverForce, "force", false, "proceed despite high HDOP")
	cmd.Flags().StringVar(&hoverLogPath, "log", "", "goto audit log path (default logs/goto.jsonl)")
	return cmd
}

var (
	gotoLat, gotoLon       float64
	gotoAlt                float32
	gotoYaw                float32
	gotoUseYaw, gotoWait   bool
	gotoForce              bool
	gotoTimeout            time.Duration
	gotoLogPath            string
	hoverAlt               float32
	hoverWait, hoverForce  bool
	hoverTimeout           time.Duration
	hoverLogPath           string
)
