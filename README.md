# gh-dependabot

A [GitHub CLI][gh-cli] extension to quickly review and approve
[Dependabot][dependabot] PRs.

[gh-cli]: https://github.com/cli/cli
[dependabot]: https://github.blog/2020-06-01-keep-all-your-packages-up-to-date-with-dependabot/

## Installation

This extension is developed and tested against a minimum version (2.2.0) of the [GitHub CLI][gh-cli].

1. Install the `gh cli` - see the [installation instructions][cli-installation]

2. Install this extension:

```sh
gh extension install einride/gh-dependabot
```

[cli-installation]: https://github.com/cli/cli#installation

## Usage

```
 $ gh dependabot --help
Manage Dependabot PRs.

Usage:
  gh dependabot [flags]

Examples:
gh dependabot --org einride

Flags:
  -h, --help          help for gh
  -o, --org string    organization to query (e.g. einride)
  -t, --team string   team to query (e.g. einride/team-transport-execution)
```
