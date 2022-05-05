package main

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"os"
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

	maxVersion := semver.MustParse("v0.1.0")
	err = tagrefs.ForEach(func(t *plumbing.Reference) error {
		fmt.Println("1", t.Name().Short())
		version, err := semver.NewVersion(t.Name().Short())
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

	fmt.Println("Found max tag", maxVersion.String())
	return nil
}
