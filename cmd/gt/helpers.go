package main

import (
	"context"
	"errors"
	"fmt"

	repomanager "github.com/kazhuravlev/git-tools/internal/repo-manager"
	"github.com/urfave/cli/v3"
)

func withManager(action func(context.Context, *cli.Command, *repomanager.Manager) error) cli.ActionFunc {
	return func(ctx context.Context, c *cli.Command) error {
		repoPath := c.String(flagRepoPath)
		if repoPath == "" {
			return errors.New("path to repo must be set by flag " + flagRepoPath)
		}

		manager, err := repomanager.New(repoPath)
		if err != nil {
			return fmt.Errorf("cannot build repo manager: %w", err)
		}

		return action(ctx, c, manager)
	}
}
