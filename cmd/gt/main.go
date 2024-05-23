package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	repomanager "github.com/kazhuravlev/git-tools/internal/repo-manager"
	"github.com/urfave/cli/v3"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
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
			{
				Name:    "hooks",
				Aliases: []string{"h"},
				Usage:   "install and run hooks",
				Commands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"l"},
						Action:  withManager(cmdHooksList),
						Usage:   "list available hooks",
					},
					{
						Name:  "exec",
						Usage: "execute hook",
						Commands: []*cli.Command{
							{
								Name:   "commit-msg",
								Action: withManager(cmdHookExecCommitMsg),
							},
						},
					},
					{
						Name:    "install",
						Aliases: []string{"i"},
						Usage:   "install git-tool CLI as git-hook program",
						Commands: []*cli.Command{
							{
								Name:    "all",
								Aliases: []string{"a"},
								Action:  withManager(cmdHooksInstallAll),
								Usage:   "install all supported git-hooks",
							},
						},
					},
				},
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

func cmdHooksList(ctx context.Context, c *cli.Command, m *repomanager.Manager) error {
	fmt.Println("commit-msg")

	return nil
}

func cmdHookExecCommitMsg(ctx context.Context, c *cli.Command, m *repomanager.Manager) error {
	branch, err := m.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("cannot get current branch: %w", err)
	}

	commitMsgFilename := c.Args().First()
	commitBuf, err := os.ReadFile(commitMsgFilename)
	if err != nil {
		return fmt.Errorf("cannot read commit-msg: %w", err)
	}

	// TODO(zhuravlev): implement a different policies that can be configured through config file for this project.
	commitMessage := fmt.Sprintf("%s %s", branch, string(commitBuf))
	if err := os.WriteFile(commitMsgFilename, []byte(commitMessage), 0644); err != nil {
		return fmt.Errorf("cannot write commit-msg: %w", err)
	}

	return nil
}

func cmdHooksInstallAll(ctx context.Context, c *cli.Command, m *repomanager.Manager) error {
	// backup current hooks
	// add notes about backed up hooks
	// install ourself as hook executor

	hooks := []string{
		"commit-msg",
	}
	for i := range hooks {
		hookFilename := filepath.Join(".git", "hooks", hooks[i])
		hookFilenameBackup := hookFilename + "__backup"

		wasBackedUp, err := backupFile(hookFilename, hookFilenameBackup)
		if err != nil {
			return fmt.Errorf("backup file: %w", err)
		}

		content := bytes.NewBuffer(nil)
		content.WriteString("#!/bin/sh\n\n")
		content.WriteString("## NOTE: Code-Generated\n")
		content.WriteString("## This file was created automatically by https://github.com/kazhuravlev/git-tools\n\n")

		if wasBackedUp {
			if err := os.Remove(hookFilename); err != nil {
				return fmt.Errorf("remove hook file: %w", err)
			}

			content.WriteString(fmt.Sprintf("# hook file (%s) is backed up into (%s) at (%s)\n", hookFilename, hookFilenameBackup, time.Now().Format(time.DateTime)))
		}

		content.WriteString(fmt.Sprintf(`
if command -v gt >/dev/null 2>&1; then
  gt hooks exec %s $1
else
  echo "Can't find git-tools (gt binary)"
  echo "Check the docs https://github.com/kazhuravlev/git-tools"
	exit 1
fi
`, hooks[i]))
		content.WriteString("\n")

		target, err := os.Create(hookFilename)
		if err != nil {
			return fmt.Errorf("open hook file: %w", err)
		}

		if _, err := target.Write(content.Bytes()); err != nil {
			return fmt.Errorf("write hook file: %w", err)
		}

		if err := target.Close(); err != nil {
			return fmt.Errorf("close hook file: %w", err)
		}

		if err := os.Chmod(hookFilename, 0755); err != nil {
			return fmt.Errorf("chmod hook file: %w", err)
		}

		fmt.Println("Hook was installed:", hookFilename)
	}

	return nil
}

func backupFile(source, dest string) (bool, error) {
	_, err := os.Stat(source)

	switch {
	default:
		return false, fmt.Errorf("unknown error: %w", err)
	case os.IsNotExist(err):
		return false, nil // nothing was backed up
	case err == nil:
		in, err := os.Open(source)
		if err != nil {
			return false, fmt.Errorf("cannot open source file: %w", err)
		}

		defer in.Close()

		out, err := os.Create(dest)
		if err != nil {
			return false, fmt.Errorf("cannot create destination file: %w", err)
		}

		defer out.Close()

		if _, err := io.Copy(out, in); err != nil {
			return false, fmt.Errorf("cannot copy source file to destination file: %w", err)
		}

		return true, nil
	}
}
