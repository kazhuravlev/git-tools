package main

import (
	"context"
	"errors"
	"fmt"
	repomanager "github.com/kazhuravlev/git-tools/internal/repo-manager"
	"github.com/urfave/cli/v3"
	"os"
	"strings"
)

const (
	flagRepoPath        = "repo"
	flagIgnoreExistsTag = "ignore-exists-tag"
)

var (
	version = "unknown-local-build"
)

var (
	cliFlagIgnoreExistsTag = &cli.BoolFlag{
		Name:  flagIgnoreExistsTag,
		Usage: "Use this option to force adding a new semver tag event when another one is exists",
		Value: false,
	}
)

func main() {
	a := &cli.Command{
		Version: version,
		Name:    "gt",
		Usage:   "Git tools",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     flagRepoPath,
				Required: false,
				Value:    ".",
				Usage:    "path to the repo which you want to manage",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "tag",
				Aliases: []string{"t"},
				Usage:   "manage tags",
				Commands: []*cli.Command{
					{
						Name:    "increment",
						Aliases: []string{"i"},
						Usage:   "find the last semver tag and increment the concrete part",
						Commands: []*cli.Command{
							{
								Name:    "major",
								Aliases: []string{"maj"},
								Flags:   []cli.Flag{cliFlagIgnoreExistsTag},
								Action:  withManager(buildTagIncrementor(repomanager.ComponentMajor)),
								Usage:   "increment major part of semver",
							},
							{
								Name:    "minor",
								Aliases: []string{"min"},
								Flags:   []cli.Flag{cliFlagIgnoreExistsTag},
								Action:  withManager(buildTagIncrementor(repomanager.ComponentMinor)),
								Usage:   "increment minor part of semver",
							},
							{
								Name:    "patch",
								Aliases: []string{"pat"},
								Flags:   []cli.Flag{cliFlagIgnoreExistsTag},
								Action:  withManager(buildTagIncrementor(repomanager.ComponentPatch)),
								Usage:   "increment patch part of semver",
							},
						},
					},
					{
						Name:    "last",
						Aliases: []string{"l"},
						Action:  withManager(cmdTagGetSemverLast),
						Usage:   "show last semver tag",
					},
				},
			},
			{
				Name:    "lint",
				Aliases: []string{"l"},
				Usage:   "run linter",
				Action:  withManager(cmdLint),
			},
		},
	}

	if err := a.Run(context.Background(), os.Args); err != nil {
		fmt.Println("Error: " + err.Error())
	}
}

func buildTagIncrementor(component repomanager.Component) func(context.Context, *cli.Command, *repomanager.Manager) error {
	return func(ctx context.Context, c *cli.Command, m *repomanager.Manager) error {
		ignoreExistsTag := c.Bool(flagIgnoreExistsTag)

		repoPath := c.String(flagRepoPath)
		if repoPath == "" {
			return errors.New("path to repo must be set by flag " + flagRepoPath)
		}

		m, err := repomanager.New(repoPath)
		if err != nil {
			return fmt.Errorf("cannot build repo manager: %w", err)
		}

		curTag, err := m.GetCurrentTagSemver()
		if err != nil {
			return fmt.Errorf("get current tag: %w", err)
		}

		if curTag.HasVal() && !ignoreExistsTag {
			return fmt.Errorf("semver tag is already exists: %s", curTag.Val().TagName())
		}

		oldTag, newTag, err := m.IncrementSemverTag(component)
		if err != nil {
			return fmt.Errorf("cannot increment minor: %w", err)
		}

		fmt.Printf(
			"Increment tag component [%s] from %s => %s (%s)\n",
			string(component),
			oldTag.TagName(),
			newTag.TagName(),
			newTag.Ref.Hash().String(),
		)
		return nil
	}
}

func cmdTagGetSemverLast(ctx context.Context, c *cli.Command, m *repomanager.Manager) error {
	maxTag, err := m.GetTagsSemverMax()
	if err != nil {
		return fmt.Errorf("cannot get max tag: %w", err)
	}

	fmt.Printf("%s (%s)\n", maxTag.TagName(), maxTag.Ref.Hash())
	return nil
}

func cmdLint(ctx context.Context, c *cli.Command, m *repomanager.Manager) error {
	tags, err := m.GetTagsSemverTopN(100)
	if err != nil {
		return fmt.Errorf("cannot last semver tags: %w", err)
	}

	if len(tags) == 0 {
		return nil
	}

	hasPrefix := tags[0].HasPrefixV()
	var hasErrors bool
	commit2tags := make(map[string][]string, len(tags))
	for i := range tags {
		tag := &tags[i]
		if tag.HasPrefixV() != hasPrefix {
			fmt.Printf("Tag `%s` not in one style with others.\n", tag.TagName())
			hasErrors = true
		}

		commit2tags[tag.CommitHash()] = append(commit2tags[tag.CommitHash()], tag.TagName())
	}

	for commitHash, commitTags := range commit2tags {
		if len(commitTags) == 1 {
			continue
		}

		fmt.Printf("Commit `%s` have a several semver tags: `%s`.\n", commitHash, strings.Join(commitTags, ", "))
		hasErrors = true
	}

	if hasErrors {
		os.Exit(1)
	}

	return nil
}
