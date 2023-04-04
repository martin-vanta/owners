package main

import (
	"github.com/martin-vanta/owners"
	"github.com/spf13/cobra"
)

var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Notifies owners in a GitHub Action",
	RunE:  githubRun,
}

func githubRun(cmd *cobra.Command, args []string) error {
	actions, err := owners.GetGitHubActions()
	if err != nil {
		return err
	}

	if actions.Draft {
		return nil
	}

	if err := actions.Prepare(); err != nil {
		return err
	}

	differ := owners.NewGitDiffer(actions.BaseRef, actions.HeadRef)
	diffs, err := differ.Diff()
	if err != nil {
		return err
	}

	results, err := owners.FindOwners(actions.OwnersFileName, diffs)
	if err != nil {
		return err
	}

	return actions.WriteComment(results)
}
