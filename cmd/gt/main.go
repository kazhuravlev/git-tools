package main

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"os"
	"strings"
)
import "github.com/urfave/cli"

func main() {
	a := cli.NewApp()
	a.Version = "1.0.1"
	a.Name = "gt"
	a.Usage = "Git tools"

	a.Commands = []cli.Command{
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
							Name:      "minor",
							ShortName: "min",
							Action:    cmdTagIncrementMinor,
						},
					},
				},
			},
		},
	}

	if err := a.Run(os.Args); err != nil {
		panic("cannot run command: " + err.Error())
	}
}

func cmdTagIncrementMinor(c *cli.Context) error {
	r, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	tagrefs, err := r.Tags()
	if err != nil {
		return err
	}

	maxVersion := semver.MustParse("v0.0.0")
	hasPrefixV := true
	err = tagrefs.ForEach(func(t *plumbing.Reference) error {
		tagName := t.Name().Short()
		hasPrefixV = strings.HasPrefix(tagName, "v")

		version, err := semver.NewVersion(tagName)
		if err != nil {
			return err
		}
		if version.GreaterThan(maxVersion) {
			maxVersion = version
		}

		return nil
	})
	if err != nil {
		return err
	}

	newMaxTag := maxVersion.IncMinor()
	oldTagStr := maxVersion.String()
	newTagStr := newMaxTag.String()
	if hasPrefixV {
		oldTagStr = "v" + oldTagStr
		newTagStr = "v" + newTagStr
	}

	head, err := r.Head()
	if err != nil {
		return err
	}

	if _, err := r.CreateTag(newTagStr, head.Hash(), nil); err != nil {
		return err
	}

	fmt.Printf("Increment tag minor from %s => %s (%s)", oldTagStr, newTagStr, head.Hash().String())
	return nil
}
