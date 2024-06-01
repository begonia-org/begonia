package main

import (
	"log"

	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal"
	"github.com/spf13/cobra"
)

// var ProviderSet = wire.NewSet(NewMasterCmd)
func addCommonCommand(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringP("env", "e", "dev", "Runtime Environment")
	return cmd
}
func NewInitCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "init",
		Short: "初始化数据库",
		Run: func(cmd *cobra.Command, args []string) {
			env, _ := cmd.Flags().GetString("env")
			config := config.ReadConfig(env)
			operator := internal.InitOperatorApp(config)
			log.Printf("init database")
			err := operator.Init()
			if err != nil {
				log.Fatalf("failed to init database: %v", err)
			}
		},
	}
	return cmd

}
func NewGatewayCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "启动网关服务",

		Run: func(cmd *cobra.Command, args []string) {
			endpoint, _ := cmd.Flags().GetString("endpoint")
			// name, _ := cmd.Flags().GetString("name")
			env, _ := cmd.Flags().GetString("env")
			config := config.ReadConfig(env)
			worker := internal.New(config, gateway.Log, endpoint)
			worker.Start()

		},
	}
	cmd.Flags().StringP("endpoint", "", "127.0.0.1:12138", "Endpoint Of Your Service")
	// cmd.Flags().StringP("name", "", "begonia", "Name Of Your Gateway Server")

	return cmd
}

func main() {
	cmd := NewGatewayCmd()
	cmd = addCommonCommand(cmd)
	rootCmd := &cobra.Command{Use: "gateway", Short: "网关服务"}
	rootCmd.AddCommand(cmd)
	rootCmd.AddCommand(addCommonCommand(NewInitCmd()))

	if err := cmd.Execute(); err != nil {
		log.Fatalf("failed to start master: %v", err)
	}
}
