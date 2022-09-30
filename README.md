# Git Tools

Set of helpful tools to do a routine job.

## Installation

`go install github.com/kazhuravlev/git-tools/cmd/gt`

## Usage

| Command               | Alias     | Action                                                                   |
|-----------------------|-----------|--------------------------------------------------------------------------|
| `tag last`            | `t l`     | Show last semver tag in this repo                                        |
| `tag increment major` | `t i maj` | Find the last semver tag, increment major part and add tag to local repo |
| `tag increment minor` | `t i min` | Find the last semver tag, increment minor part and add tag to local repo |
| `tag increment patch` | `t i pat` | Find the last semver tag, increment patch part and add tag to local repo |
