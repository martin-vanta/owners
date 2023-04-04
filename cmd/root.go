package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "owners",
	Short: "Owners helps you find owners for files",
	RunE:  rootRun,
}

var (
	ownersFileName string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&ownersFileName, "owners_file_name", "", "OWNERS", "name of owners files")

	rootCmd.AddCommand(findCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(githubCmd)
}

func rootRun(cmd *cobra.Command, args []string) error {
	return nil
}
