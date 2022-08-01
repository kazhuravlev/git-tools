package main

import (
	"fmt"
	repomanager "github.com/kazhuravlev/git-tools/pkg/repo-manager"
	"github.com/pkg/errors"
	"os"
)
import "github.com/urfave/cli"

const (
	flagRepoPath = "repo"
)

func main() {
	a := &cli.App{
		Version: "0.1.0",
		Name:    "gt",
		Usage:   "Git tools",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     flagRepoPath,
				Required: false,
				Value:    ".",
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
						Subcommands: []cli.Command{
							{
								Name:      "major",
								ShortName: "maj",
								Action:    buildTagIncrementor(repomanager.ComponentMajor),
							},
							{
								Name:      "minor",
								ShortName: "min",
								Action:    buildTagIncrementor(repomanager.ComponentMinor),
							},
							{
								Name:      "patch",
								ShortName: "pat",
								Action:    buildTagIncrementor(repomanager.ComponentPatch),
							},
						},
					},
					{
						Name:      "last",
						ShortName: "l",
						Action:    cmdTagGetLast,
					},
				},
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

func cmdTagGetLast(c *cli.Context) error {
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
