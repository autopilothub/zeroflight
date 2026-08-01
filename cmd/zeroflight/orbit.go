package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autopilothub/zeroflight/internal/safety"
	"github.com/autopilothub/zeroflight/pkg/geo"
	"github.com/spf13/cobra"
)

func newOrbitCmd() *cobra.Command {
	var (
		lat, lon   float64
		useGPS     bool
		radius     float64
		points     int
		alt        float32
		wait       bool
		timeout    time.Duration
		force      bool
		logPath    string
	)

	cmd := &cobra.Command{
		Use:   "orbit",
		Short: "Fly a circular path using sequential goto commands",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			state := sess.State()
			fileCfg := sess.Config()

			centerLat, centerLon := lat, lon
			if useGPS || (centerLat == 0 && centerLon == 0) {
				if state.GPS.Lat == 0 && state.GPS.Lon == 0 {
					return fmt.Errorf("no valid GPS position for orbit center")
				}
				centerLat = state.GPS.Lat
				centerLon = state.GPS.Lon
			}

			waypoints := geo.CirclePoints(centerLat, centerLon, radius, points)
			fmt.Printf("orbit: center=%.7f,%.7f radius=%.0fm points=%d alt=%.1fm\n",
				centerLat, centerLon, radius, len(waypoints), alt)

			for i, wp := range waypoints {
				fmt.Printf("waypoint %d/%d: lat=%.7f lon=%.7f\n", i+1, len(waypoints), wp[0], wp[1])
				if err := runNavCommand(sess, navCommandOptions{
					command: fmt.Sprintf("orbit-%d", i+1),
					lat:     wp[0],
					lon:     wp[1],
					alt:     alt,
					wait:    wait,
					timeout: timeout,
					force:   force,
					logPath: logPath,
				}); err != nil {
					return fmt.Errorf("waypoint %d: %w", i+1, err)
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				if safety.IsLinkStale(sess.State(), fileCfg.LinkTimeout()) {
					return fmt.Errorf("mavlink link lost during orbit")
				}
			}

			fmt.Println("orbit complete")
			return nil
		},
	}

	cmd.Flags().Float64Var(&lat, "lat", 0, "orbit center latitude (default: current GPS)")
	cmd.Flags().Float64Var(&lon, "lon", 0, "orbit center longitude (default: current GPS)")
	cmd.Flags().BoolVar(&useGPS, "center-here", true, "use current GPS as orbit center")
	cmd.Flags().Float64Var(&radius, "radius", 50, "orbit radius in meters")
	cmd.Flags().IntVar(&points, "points", 8, "number of waypoints around the circle")
	cmd.Flags().Float32Var(&alt, "alt", 10, "target altitude (meters, relative)")
	cmd.Flags().BoolVar(&wait, "wait", true, "wait for arrival at each waypoint")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Minute, "arrival wait timeout per waypoint")
	cmd.Flags().BoolVar(&force, "force", false, "proceed despite high HDOP")
	cmd.Flags().StringVar(&logPath, "log", "", "goto audit log path (default logs/goto.jsonl)")
	return cmd
}
