# Git Tools

[![Go Reference](https://pkg.go.dev/badge/github.com/kazhuravlev/git-tools.svg)](https://pkg.go.dev/github.com/kazhuravlev/git-tools)
[![License](https://img.shields.io/github/license/kazhuravlev/git-tools?color=blue)](https://github.com/kazhuravlev/git-tools/blob/master/LICENSE)
[![Build Status](https://github.com/kazhuravlev/git-tools/actions/workflows/release.yml/badge.svg)](https://github.com/kazhuravlev/git-tools/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kazhuravlev/git-tools)](https://goreportcard.com/report/github.com/kazhuravlev/git-tools)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

A CLI tool for managing git repositories with semantic versioning tags. It helps with version bumps, tag validation,
author reporting, and lightweight git hook automation.

**Quick example:**

```shell
git commit -am "changes" && gt t i min && git push --follow-tags
# Commits changes, increments minor version (e.g., v1.2.3 → v1.3.0), ready to push
```

**Key features:**

- Automatic semver tag incrementing (major/minor/patch)
- Tag format consistency checking (`v1.2.3` vs `1.2.3`) and duplicate semver tags on the same commit
- Git hook installation for `commit-msg`
- Author statistics based on commit history

## Installation

**Golang**

```shell
go install github.com/kazhuravlev/git-tools/cmd/gt@latest
```

**Homebrew**

```shell
brew install kazhuravlev/git-tools/git-tools
```

**Docker (zsh)**

```shell
echo 'alias gt="docker run -it --rm -v $(pwd):/workdir -w /workdir kazhuravlev/gt:latest"' >> ~/.zshrc
 ```

## Usage

All commands work with the current directory by default. Use `--repo=/path/to/repo` to specify a different repository.

### Commands

**Tag Management:**

```shell
gt tag last              # Show the highest semver tag and commit hash (alias: gt t l)
gt tag last -f tag       # Show only the tag name (useful for CI/CD)

gt tag increment major   # Create the next major tag: v1.2.3 → v2.0.0 (alias: gt t i maj)
gt tag increment minor   # Create the next minor tag: v1.2.3 → v1.3.0 (alias: gt t i min)
gt tag increment patch   # Create the next patch tag: v1.2.3 → v1.2.4 (alias: gt t i pat)
```

**Other Commands:**

```shell
gt lint                  # Validate recent semver tags for style consistency and duplicate tags per commit
gt authors               # List commit authors with statistics (alias: gt a)
gt hooks install all     # Install supported git hooks into .git/hooks (alias: gt h i a)
gt hooks list            # List available hooks (alias: gt h l)
```

### Options

**Force tag creation:**

```shell
# Create a new tag even if current commit already has a semver tag
gt t i min --ignore-exists-tag
```

By default, `gt tag increment ...` refuses to create a new semver tag if the current commit already has one.

**Custom repository path:**

```shell
gt --repo=/path/to/repo tag last
```

### Examples

```shell
# Typical workflow: commit, bump version, push
git commit -am "Add new feature"
gt t i min  # Creates v1.3.0 tag
git push --follow-tags

# CI/CD: Get version for build artifact
VERSION=$(gt tag last -f tag)
echo "Building version: $VERSION"

# Check recent tags before release
gt lint

# Show commit authors sorted by commit count
gt authors

# Install the commit-msg hook that prefixes the current branch name
gt hooks install all
```
