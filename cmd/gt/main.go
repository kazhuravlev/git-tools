package main

import (
	"fmt"
	repomanager "github.com/kazhuravlev/git-tools/internal/repo-manager"
	"github.com/pkg/errors"
	"os"
)
import "github.com/urfave/cli"

const (
	flagRepoPath = "repo"
)

var (
	version = "unknown-local-build"
)

func main() {
	a := &cli.App{
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
		Commands: []cli.Command{
			{
				Name:      "tag",
				ShortName: "t",
				Usage:     "manage tags",
				Subcommands: []cli.Command{
					{
						Name:      "increment",
						ShortName: "i",
						Usage:     "find the last semver tag and increment the concrete part",
						Subcommands: []cli.Command{
							{
								Name:      "major",
								ShortName: "maj",
								Action:    buildTagIncrementor(repomanager.ComponentMajor),
								Usage:     "increment major part of semver",
							},
							{
								Name:      "minor",
								ShortName: "min",
								Action:    buildTagIncrementor(repomanager.ComponentMinor),
								Usage:     "increment minor part of semver",
							},
							{
								Name:      "patch",
								ShortName: "pat",
								Action:    buildTagIncrementor(repomanager.ComponentPatch),
								Usage:     "increment patch part of semver",
							},
						},
					},
					{
						Name:      "last",
						ShortName: "l",
						Action:    cmdTagGetSemverLast,
						Usage:     "show last semver tag",
					},
				},
			},
			{
				Name:      "lint",
				ShortName: "l",
				Usage:     "run linter",
				Action:    cmdLint,
			},
		},
	}

	if err := a.Run(os.Args); err != nil {
		fmt.Println("Error: " + err.Error())
	}
}

func buildTagIncrementor(component repomanager.Component) func(ctx *cli.Context) error {
	return func(c *cli.Context) error {
		repoPath := c.GlobalString(flagRepoPath)
		if repoPath == "" {
			return errors.New("path to repo must be set by flag " + flagRepoPath)
		}

		m, err := repomanager.New(repoPath)
		if err != nil {
			return errors.Wrap(err, "cannot build repo manager")
		}

		oldTag, newTag, err := m.IncrementSemverTag(component)
		if err != nil {
			return errors.Wrap(err, "cannot increment minor")
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

func cmdTagGetSemverLast(c *cli.Context) error {
	repoPath := c.GlobalString(flagRepoPath)
	if repoPath == "" {
		return errors.New("path to repo must be set by flag " + flagRepoPath)
	}

	m, err := repomanager.New(repoPath)
	if err != nil {
		return errors.Wrap(err, "cannot build repo manager")
	}

	maxTag, err := m.GetTagsSemverMax()
	if err != nil {
		return errors.Wrap(err, "cannot get max tag")
	}

	fmt.Printf("%s (%s)\n", maxTag.TagName(), maxTag.Ref.Hash())
	return nil
}

func cmdLint(c *cli.Context) error {
	repoPath := c.GlobalString(flagRepoPath)
	if repoPath == "" {
		return errors.New("path to repo must be set by flag " + flagRepoPath)
	}

	m, err := repomanager.New(repoPath)
	if err != nil {
		return errors.Wrap(err, "cannot build repo manager")
	}

	tags, err := m.GetTagsSemverTopN(100)
	if err != nil {
		return errors.Wrap(err, "cannot last semver tags")
	}

	if len(tags) == 0 {
		return nil
	}

	hasPrefix := tags[0].HasPrefixV()
	var hasErrors bool
	for i := range tags {
		tag := &tags[i]
		if tag.HasPrefixV() != hasPrefix {
			fmt.Printf("Tag `%s` not in one style with others.\n", tag.TagName())
			hasErrors = true
		}
	}

	if hasErrors {
		os.Exit(1)
	}

	return nil
}
