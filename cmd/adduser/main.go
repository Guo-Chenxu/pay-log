package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Guo-Chenxu/pay-log/config"
	"github.com/Guo-Chenxu/pay-log/dal"
	"github.com/Guo-Chenxu/pay-log/pkg/snowflake"
	"github.com/Guo-Chenxu/pay-log/service"
)

func main() {
	var username, password, configFile string

	cmd := &cobra.Command{
		Use:   "adduser",
		Short: "Add a user to pay-log",
		Run: func(cmd *cobra.Command, args []string) {
			config.Init(configFile)
			if err := snowflake.Init(2); err != nil {
				fmt.Fprintf(os.Stderr, "snowflake init failed: %v\n", err)
				os.Exit(1)
			}
			dal.Init()

			svc := service.NewAuthService()
			if err := svc.CreateUser(username, password); err != nil {
				fmt.Fprintf(os.Stderr, "failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("user %q created\n", username)
		},
	}

	cmd.Flags().StringVarP(&username, "username", "u", "", "username (required)")
	cmd.Flags().StringVarP(&password, "password", "p", "", "password (required)")
	cmd.Flags().StringVarP(&configFile, "config", "f", "config.yaml", "config file")
	cmd.MarkFlagRequired("username") //nolint:errcheck
	cmd.MarkFlagRequired("password") //nolint:errcheck

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
