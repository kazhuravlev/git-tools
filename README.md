# Git Tools

[![Go Reference](https://pkg.go.dev/badge/github.com/kazhuravlev/git-tools.svg)](https://pkg.go.dev/github.com/kazhuravlev/git-tools)
[![License](https://img.shields.io/github/license/kazhuravlev/git-tools?color=blue)](https://github.com/kazhuravlev/git-tools/blob/master/LICENSE)
[![Build Status](https://github.com/kazhuravlev/git-tools/actions/workflows/release.yml/badge.svg)](https://github.com/kazhuravlev/git-tools/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kazhuravlev/git-tools)](https://goreportcard.com/report/github.com/kazhuravlev/git-tools)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

A CLI tool for managing git repositories with semantic versioning tags. Simplifies version bumping, tag validation, and
git hook management.

**Quick example:**

```shell
git commit -am "changes" && gt t i min && git push --follow-tags
# Commits changes, increments minor version (e.g., v1.2.3 → v1.3.0), ready to push
```

**Key features:**

- Automatic semver tag incrementing (major/minor/patch)
- Tag format consistency checking (`v1.2.3` vs `1.2.3`)
- Git hook installation (commit-msg with branch name)
- Author statistics

## Installation

**Golang**

```shell
go install github.com/kazhuravlev/git-tools/cmd/gt@latest
```

**Homebrew**

```shell
brew install kazhuravlev/git-tools/git-tools
```

**Docker (zsh)** (will work only in current directory)

```shell
echo 'alias gt="docker run -it --rm -v `pwd`:/workdir kazhuravlev/gt:latest"' >> ~/.zshrc
 ```

## Usage

All commands work with the current directory by default. Use `--repo=/path/to/repo` to specify a different repository.

### Commands

**Tag Management:**

```shell
gt tag last              # Show last semver tag (alias: gt t l)
gt tag last -f tag       # Show only tag name (useful for CI/CD)

gt tag increment major   # Bump major version: v1.2.3 → v2.0.0 (alias: gt t i maj)
gt tag increment minor   # Bump minor version: v1.2.3 → v1.3.0 (alias: gt t i min)
gt tag increment patch   # Bump patch version: v1.2.3 → v1.2.4 (alias: gt t i pat)
```

**Other Commands:**

```shell
gt lint                  # Validate tag format consistency
gt authors               # List commit authors with statistics (alias: gt a)
gt hooks install all     # Install git hooks (alias: gt h i a)
gt hooks list            # List available hooks (alias: gt h l)
```

### Options

**Force tag creation:**

```shell
# Create a new tag even if current commit already has a semver tag
gt t i min --ignore-exists-tag
```

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

# Check tag consistency before release
gt lint
```
