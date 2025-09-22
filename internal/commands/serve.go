/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package commands

import (
	"context"

	"github.com/weeb-vip/list-service/http"
	"github.com/weeb-vip/list-service/internal/logger"
	"github.com/weeb-vip/list-service/tracing"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger
		logger.Logger(
			logger.WithServerName("list-service"),
			logger.WithVersion("1.0.0"),
		)

		// Initialize tracing
		ctx := context.Background()
		tracedCtx, err := tracing.InitTracing(ctx)
		if err != nil {
			log := logger.FromCtx(ctx)
			log.Error().Err(err).Msg("Failed to initialize tracing")
			// Continue without tracing if initialization fails
			tracedCtx = ctx
		} else {
			defer func() {
				if err := tracing.Shutdown(context.Background()); err != nil {
					log := logger.FromCtx(tracedCtx)
					log.Error().Err(err).Msg("Error shutting down tracing")
				}
			}()
			log := logger.FromCtx(tracedCtx)
			log.Info().Msg("Tracing initialized successfully")
		}

		// Start the server with traced context
		return http.StartServerWithContext(tracedCtx)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
