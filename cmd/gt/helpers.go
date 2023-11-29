package main

import (
	"errors"
	"fmt"

	repomanager "github.com/kazhuravlev/git-tools/internal/repo-manager"
	"github.com/urfave/cli/v3"
)

func withManager(action func(c *cli.Context, manager *repomanager.Manager) error) cli.ActionFunc {
	return func(c *cli.Context) error {
		repoPath := c.String(flagRepoPath)
		if repoPath == "" {
			return errors.New("path to repo must be set by flag " + flagRepoPath)
		}

		manager, err := repomanager.New(repoPath)
		if err != nil {
			return fmt.Errorf("cannot build repo manager: %w", err)
		}

		return action(c, manager)
	}
}
