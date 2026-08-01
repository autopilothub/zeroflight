package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/autopilothub/zeroflight/internal/mission"
	"github.com/autopilothub/zeroflight/internal/safety"
	"github.com/spf13/cobra"
)

func newMissionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mission",
		Short: "Upload or clear INAV waypoint missions",
	}
	cmd.AddCommand(newMissionUploadCmd())
	cmd.AddCommand(newMissionClearCmd())
	return cmd
}

func newMissionUploadCmd() *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a YAML mission file to INAV (disarmed only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			plan, err := mission.Load(file)
			if err != nil {
				return err
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
			state := sess.State()
			if state.Armed {
				return fmt.Errorf("disarm the vehicle before uploading a mission")
			}

			limits := safety.LimitsFromConfig(sess.Config().Safety)
			if err := safety.ValidateMission(state.Home, plan.Waypoints, limits); err != nil {
				return err
			}

			fmt.Printf("uploading %d waypoints...\n", len(plan.Waypoints))
			if err := sess.Client().UploadMission(ctx, plan.Waypoints); err != nil {
				return err
			}
			fmt.Println("mission uploaded; switch to MISSION mode on the transmitter to fly")
			return nil
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "mission YAML file")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func newMissionClearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear the stored mission on INAV",
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
			if sess.State().Armed {
				return fmt.Errorf("disarm the vehicle before clearing a mission")
			}

			if err := sess.Client().ClearMission(ctx); err != nil {
				return err
			}
			fmt.Println("mission cleared")
			return nil
		},
	}
	return cmd
}
