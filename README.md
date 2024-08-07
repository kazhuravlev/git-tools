# Git Tools

[![Go Reference](https://pkg.go.dev/badge/github.com/kazhuravlev/git-tools.svg)](https://pkg.go.dev/github.com/kazhuravlev/git-tools)
[![License](https://img.shields.io/github/license/kazhuravlev/git-tools?color=blue)](https://github.com/kazhuravlev/git-tools/blob/master/LICENSE)
[![Build Status](https://github.com/kazhuravlev/git-tools/actions/workflows/release.yml/badge.svg)](https://github.com/kazhuravlev/git-tools/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kazhuravlev/git-tools)](https://goreportcard.com/report/github.com/kazhuravlev/git-tools)

Set of helpful tools to do a routine job.

This tool can help you to work with a bunch of lib/services under your projects.
Especially when you start a new project - you need to increment a version of the
SDK repo too often. `gt` allows you to work
like `git commit -a -m 'some changes' && gt t i min && gt p a`. This set of
commands reads as "commit changes, increment minor version of last semver tag,
push commits to the origin with tags". Pretty simple, huh?

`gt` try to follow a chosen format of semver tags which you choose (`v1.2.3`
/`1.2.3`). If you want to follow a selected pattern, just add `gt` to your CI
system or `git-hook` and check that all tags have one concrete format.
See [Usage](#Usage).

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
