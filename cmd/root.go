package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "os"
)

// RootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
    Use:   "devopsmate",
    Short: "DevOpsMate is a CLI tool",
    Long:  `A longer description of your DevOpsMate CLI tool.`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Hello from DevOpsMate CLI!")
    },
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
