Tag and release cece. Pass the version bump type as an argument: patch, minor, major, or an explicit version.

## Steps

1. Run `/release $ARGUMENTS` to tag, push, and create the GitHub release.
2. Wait for CI to complete: poll `gh run list --workflow=release.yml --limit=1 --json status,conclusion` until `conclusion` is `success`.
3. Run `cece update` to install the new version locally.
4. Run `cece version` to confirm the update.
