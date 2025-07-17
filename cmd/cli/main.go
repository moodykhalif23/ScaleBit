package main

import (
	"fmt"
	"os"

	"io/ioutil"
	"path/filepath"
	"strings"

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
		serviceName := args[0]
		port := "8080" // default port, could be a flag
		templateDir := filepath.Join("internal", "pkg", "cli", "templates", lang+"-service")
		outputDir := serviceName
		os.MkdirAll(outputDir, 0755)

		// Copy and replace placeholders in main.go
		mainTmpl, _ := ioutil.ReadFile(filepath.Join(templateDir, "main.go"))
		mainStr := strings.ReplaceAll(string(mainTmpl), "{{SERVICE_NAME}}", serviceName)
		mainStr = strings.ReplaceAll(mainStr, "{{PORT}}", port)
		ioutil.WriteFile(filepath.Join(outputDir, "main.go"), []byte(mainStr), 0644)

		// Copy and replace placeholders in Dockerfile
		dockerTmpl, _ := ioutil.ReadFile(filepath.Join(templateDir, "Dockerfile"))
		dockerStr := strings.ReplaceAll(string(dockerTmpl), "{{SERVICE_NAME}}", serviceName)
		dockerStr = strings.ReplaceAll(dockerStr, "{{PORT}}", port)
		ioutil.WriteFile(filepath.Join(outputDir, "Dockerfile"), []byte(dockerStr), 0644)

		// Generate microservice.yaml CRD manifest
		crd := `apiVersion: sme.moodykhalif23.github.com/v1alpha1
kind: Microservice
metadata:
  name: ` + serviceName + `
spec:
  image: ` + serviceName + `:latest
  port: ` + port + `
  replicas: 1
`
		ioutil.WriteFile(filepath.Join(outputDir, "microservice.yaml"), []byte(crd), 0644)

		fmt.Printf("Service '%s' scaffolded in %s/\n", serviceName, outputDir)
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
