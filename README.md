gh-dependabot
=============

A [GitHub CLI](https://github.com/cli/cli) extension to quickly review and approve [Dependabot](https://github.blog/2020-06-01-keep-all-your-packages-up-to-date-with-dependabot/) PRs.

Installation
------------

This extension is developed and tested against a minimum version (2.2.0) of the [GitHub CLI](https://github.com/cli/cli).

1.	Install the `gh cli` - see the [installation instructions](https://github.com/cli/cli#installation)

2.	Install this extension:

	```sh
	gh extension install einride/gh-dependabot
	```

Usage
-----

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
