package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/begonia-org/begonia"
	"github.com/begonia-org/begonia/config"
	"github.com/begonia-org/begonia/gateway"
	"github.com/begonia-org/begonia/internal"
	"github.com/begonia-org/go-sdk/client"
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
		Short: "Init Database",
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
		Short: "Start Gateway Server",

		Run: func(cmd *cobra.Command, args []string) {
			endpoint, _ := cmd.Flags().GetString("endpoint")
			env, _ := cmd.Flags().GetString("env")
			config := config.ReadConfig(env)
			worker := internal.New(config, gateway.Log, endpoint)
			hd, _ := os.UserHomeDir()
			_ = os.WriteFile(hd+"/.begonia/gateway.json", []byte(fmt.Sprintf(`{"addr":"http://%s"}`, endpoint)), 0666)
			worker.Start()

		},
	}
	cmd.Flags().StringP("endpoint", "", "127.0.0.1:12138", "Endpoint Of Your Service")
	// cmd.Flags().StringP("name", "", "begonia", "Name Of Your Gateway Server")

	return cmd
}
func NewEndpointDelCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "del",
		Short: "Delete Service From Gateway",

		Run: func(cmd *cobra.Command, args []string) {
			id, _ := cmd.Flags().GetString("id")

			DeleteEndpoint(id)
		},
	}
	cmd.Flags().StringP("id", "i", "", "ID Of Your Service")
	_ = cmd.MarkFlagRequired("id")
	return cmd

}
func NewEndpointAddCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add",
		Short: "Add Service To Gateway",

		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			desc, _ := cmd.Flags().GetString("desc")
			tags, _ := cmd.Flags().GetStringArray("tags")
			balance, _ := cmd.Flags().GetString("balance")
			endpoints, _ := cmd.Flags().GetStringArray("endpoint")

			RegisterEndpoint(name, endpoints, desc, client.WithBalance(strings.ToUpper(balance)), client.WithTags(tags))
		},
	}
	cmd.Flags().StringArrayP("endpoint", "p", []string{}, "Endpoint Of Your Service (example:127.0.0.1:1949)")
	cmd.Flags().StringP("name", "n", "", "Service Name")
	cmd.Flags().StringP("desc", "d", "", "Descriptions Set Of Your Service (example:./example/example.pb)")
	cmd.Flags().StringArrayP("tags", "t", []string{}, "Tags Of Your Service")
	cmd.Flags().StringP("balance", "b", "RR", "Balance Type Of Your Service (options: RR WRR LC WLC CH SED NQ)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("endpoint")
	_ = cmd.MarkFlagRequired("desc")
	return cmd
}
func NewEndpointCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "endpoint",
		Short: "Manage Service Of Gateway",
	}
	// cmd.Flags().StringP("addr", "a", "http://127.0.0.1:12138", "Address Of Begonia Gateway server")
	// _ = cmd.MarkFlagRequired("addr")

	cmd.AddCommand(NewEndpointAddCmd())
	cmd.AddCommand(NewEndpointDelCmd())
	return cmd
}
func NewBegoniaInfoCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Output version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Begonia Version: ", begonia.Version)
			fmt.Println("Build Time: ", begonia.BuildTime)
			fmt.Println("Commit: ", begonia.Commit)
			fmt.Println("Env: ", begonia.Env)
		},
	}
	return cmd
}

func main() {
	cmd := NewGatewayCmd()
	cmd = addCommonCommand(cmd)
	rootCmd := &cobra.Command{Use: "gateway", Short: "网关服务"}
	rootCmd.AddCommand(cmd)
	rootCmd.AddCommand(NewBegoniaInfoCmd())
	rootCmd.AddCommand(addCommonCommand(NewInitCmd()))
	rootCmd.AddCommand(NewEndpointCmd())
	if err := cmd.Execute(); err != nil {
		log.Fatalf("failed to start master: %v", err)
	}
}
