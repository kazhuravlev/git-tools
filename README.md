# Git Tools

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

`go install github.com/kazhuravlev/git-tools/cmd/gt`

## Usage

| Command               | Alias     | Action                                                                   |
|-----------------------|-----------|--------------------------------------------------------------------------|
| `tag last`            | `t l`     | Show last semver tag in this repo                                        |
| `tag increment major` | `t i maj` | Find the last semver tag, increment major part and add tag to local repo |
| `tag increment minor` | `t i min` | Find the last semver tag, increment minor part and add tag to local repo |
| `tag increment patch` | `t i pat` | Find the last semver tag, increment patch part and add tag to local repo |
| `lint`                | `l`       | Run linter, that check the problems.                                     |

