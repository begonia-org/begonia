package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/wetrycode/begonia/config"
	"github.com/wetrycode/begonia/internal/pkg/logger"
)

// var ProviderSet = wire.NewSet(NewMasterCmd)
func addCommonCommand(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringP("env", "e", "dev", "Runtime Environment")
	return cmd
}
func NewGatewayCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "gateway start",
		Short: "启动网关服务",

		Run: func(cmd *cobra.Command, args []string) {
			endpoint, _ := cmd.Flags().GetString("endpoint")
			env, _ := cmd.Flags().GetString("env")
			config := config.ReadConfig(env)
			server := initApp(config, logger.Logger, endpoint)
			err := server.Start()
			if err != nil {
				log.Fatalf("failed to start master: %v", err)
			}

		},
	}
	cmd.Flags().StringP("endpoint", "", "0.0.0.0:12138", "Endpoint Of Your Service")
	return cmd
}

func main() {
	cmd := NewGatewayCmd()
	cmd = addCommonCommand(cmd)
	if err := cmd.Execute(); err != nil {
		log.Fatalf("failed to start master: %v", err)
	}
}
