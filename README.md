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

By default, `gt` works with the repo in the current directory. If you want to
specify another path to repo - add `--repo=/path/to/repo` flag.

```shell
gt --repo /path/to/repo tag last
```

| Command             | Action                                                       |
|---------------------|--------------------------------------------------------------|
| `t l`               | Show last semver tag in this repo                            |
| `t i major`         | Increment `major` part for semver                            |
| `t i minor`         | Increment `minor` part for semver                            |
| `t i patch`         | Increment `patch` part for semver                            |
| `lint`              | Run linter, that check the problems                          |
| `hooks install all` | Install commit-msg hook that adds branch name to each commit |

### Force add new semver tag

By default `gt` will throw an error when you try to increment a tag on commit, that already have another semver tag.

In order to skip this error - provide additional flag to increment command like
that: `gt t i min --ignore-exists-tag`.

### Examples

```shell
# Get last semver tag in this repo
$ gt tag last
v1.9.0 (c2e70ec90579ba18fd73078e98f677aec75ae002)

# Show only tag name (useful for ci/cd)
$ gt tag last -f tag
v1.9.0
```
