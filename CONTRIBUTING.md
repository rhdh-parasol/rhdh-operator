# Contributing

This is the rhdh-parasol playground for the rhdh-operator project.
Contributions are welcome — here is how to get started.

## Proposing a Change

1. Fork the repository or create a branch off `main`.
2. Make your changes in the new branch.
3. Open a pull request against `main`.

## Pull Request Expectations

Every pull request should include:

- A clear description of **what** changed and **why**.
- A link to the related GitHub issue, if one exists (use
  `Fixes #<number>` or `Relates to #<number>` in the PR description).
- Passing CI checks before requesting review.

## Fullsend (AI-Assisted Workflows)

This repository uses [fullsend](https://github.com/fullsend-ai/fullsend)
to automate parts of the issue and pull request lifecycle. Fullsend
agents run as GitHub Actions triggered by slash commands in issue or PR
comments.

### Available Commands

| Command | Where | What it does |
|---------|-------|-------------|
| `/fs-triage` | Issue comment | Analyzes the issue, identifies root cause, and labels it `ready-to-code` when triage is complete. |
| `/fs-code` | Issue comment | Creates a pull request that implements the triaged issue. |
| `/fs-code <instruction>` | Issue comment | Same as `/fs-code`, but with additional guidance for the coding agent. |
| `/fs-review` | PR comment | Runs an automated code review on the pull request. |
| `/fs-fix` | PR comment | Attempts to fix failing CI checks or review feedback on the PR. |
| `/fs-fix-stop` | PR comment | Disables the fix agent for the current PR (adds the `fullsend-no-fix` label). |

### Configured Roles

The fullsend configuration (`.fullsend/config.yaml`) enables the
following roles:

- **triage** — analyzes issues to identify root cause, affected
  components, and proposed fixes.
- **coder** — implements triaged issues by creating pull requests with
  tests.
- **review** — reviews pull requests for correctness and style.
- **fix** — automatically attempts to resolve CI failures or review
  feedback on pull requests.
- **retro** — runs retrospectives on completed work.
- **prioritize** — helps prioritize the issue backlog.

## Reporting Issues

Use [GitHub Issues](../../issues) to report bugs or suggest
improvements for this repository.
