package main

import (
	"encoding/json"
	"fmt"

	"github.com/martin-vanta/owners"
	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find owners for files",
	RunE:  findRun,
}

var (
	changedFilesFilePath string
	gitSince             string
	outputFormat         string
)

func init() {
	findCmd.PersistentFlags().StringVarP(&changedFilesFilePath, "file", "f", "", "file with list of file names")
	findCmd.PersistentFlags().StringVarP(&gitSince, "since", "", "", "files changed since this git ref")
	findCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", `output format (one of "text", "json")`)
}

func findRun(cmd *cobra.Command, args []string) error {
	var differ owners.Differ
	if changedFilesFilePath != "" {
		differ = owners.NewFileDiffer(changedFilesFilePath)
	} else if gitSince != "" {
		differ = owners.NewGitDiffer(gitSince)
	} else {
		differ = owners.NewLiteralDiffer(nil)
	}

	diffs, err := differ.Diff()
	if err != nil {
		return err
	}

	results, err := owners.FindOwners(ownersFileName, diffs)
	if err != nil {
		return err
	}

	switch outputFormat {
	case "text":
		fmt.Println(results.String())
	case "json":
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown output format: %s", outputFormat)
	}

	return nil
}
