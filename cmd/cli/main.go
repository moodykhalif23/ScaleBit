package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sme-cli",
	Short: "SME Microservices Platform CLI",
}

var createServiceCmd = &cobra.Command{
	Use:   "create-service [name]",
	Short: "Create new microservice",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lang, _ := cmd.Flags().GetString("lang")
		fmt.Printf("Creating %s service in %s...\n", args[0], lang)
		// Service generation logic
	},
}

func init() {
	createServiceCmd.Flags().StringP("lang", "l", "go", "Language template")
	rootCmd.AddCommand(createServiceCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
