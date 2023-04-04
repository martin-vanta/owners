package main

import (
	"github.com/martin-vanta/owners"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate or edit a CODEOWNERS file",
	RunE:  generateRun,
}

var (
	codeOwnersFilePath string
)

func init() {
	generateCmd.PersistentFlags().StringVarP(&codeOwnersFilePath, "file", "f", "CODEOWNERS", "CODEOWNERS file path")
}

func generateRun(cmd *cobra.Command, args []string) error {
	return owners.GenerateCodeOwners(ownersFileName, codeOwnersFilePath)
}
