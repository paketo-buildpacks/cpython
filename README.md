# Paketo Buildpack for CPython Cloud Native

A copy of the upstream README.md can be found here: [README-upstream.md](README-upstream.md)

# Plotly Branches

- The `main-upstream` branch of this repository is used to sync in the `main` branch of the upstream as its source. It does not contain any custom Plotly code changes
- The `main` branch of this repository is used to be the most up to date branch containing custom Plotly code changes
- The `XY-release` branches of this repository are made off of the `main` branch as part of our code freeze process: [DE Release Guide](https://plotly.atlassian.net/wiki/x/doAbH)

# Plotly Release Flow

- All PRs (other than an upstream sync detailed below) should first land on `main`
- If a PR is needed for a given release, a new PR should be made against that release branch
- When needing to pull in upstream changes (typically done as a post-release task to stay up to date as needed):
  - Navigate the the [main-upstream](https://github.com/plotly/paketo-buildpacks_cpython/tree/main-upstream) branch
  - Select **Sync fork**
  - A PR should then be opened to add these changes to our `main` branch, favouring our custom work whenever there are conflicts

