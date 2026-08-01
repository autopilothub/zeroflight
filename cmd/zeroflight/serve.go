package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/autopilothub/zeroflight/internal/api"
	"github.com/autopilothub/zeroflight/internal/service"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var listen string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run HTTP API with persistent MAVLink connection",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			svc, err := service.New(ctx, cfgPath, connection)
			if err != nil {
				return err
			}
			defer svc.Close()

			if err := svc.WaitReady(ctx); err != nil {
				return err
			}

			addr := svc.Config().ListenAddr(listen)
			server := api.NewServer(svc)
			httpServer := &http.Server{
				Addr:              addr,
				Handler:           server.Handler(),
				ReadHeaderTimeout: 5 * time.Second,
			}

			go func() {
				<-ctx.Done()
				shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer shutdownCancel()
				_ = httpServer.Shutdown(shutdownCtx)
			}()

			fmt.Printf("ZeroFlight API listening on http://%s\n", addr)
			fmt.Printf("ZeroFlight GCS dashboard: http://%s/\n", addr)
			fmt.Println("  GET  /health")
			fmt.Println("  GET  /api/v1/status")
			fmt.Println("  GET  /api/v1/preflight")
			fmt.Println("  POST /api/v1/goto")
			fmt.Println("  POST /api/v1/hover")
			fmt.Println("  POST /api/v1/mission/upload")
			fmt.Println("  POST /api/v1/mission/clear")

			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&listen, "listen", "", "override api listen address (host:port)")
	return cmd
}
