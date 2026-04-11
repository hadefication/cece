Run a code review and security review loop on recent changes. Keep iterating until all legitimate issues are resolved. Do NOT tag or release — that is a separate step.

## Step 1: Identify changes

Run `git diff` to find unstaged changes. If there are none, run `git diff HEAD~1` to review the last commit. Note which files changed and what the changes do.

## Step 2: Launch reviews in parallel

Launch both agents simultaneously using the Agent tool:

1. **Code reviewer** (`feature-dev:code-reviewer`): Review the changed files for bugs, logic errors, code quality issues, and adherence to project conventions in CLAUDE.md. Use confidence-based filtering — only report issues with genuine confidence.

2. **Security reviewer** (`pr-review-toolkit:silent-failure-hunter`): Review the changed files for silent failures, inadequate error handling, path traversal, symlink exploitation, TOCTOU races, and inappropriate fallback behavior.

For both agents, list the specific changed files and describe what the changes do so they have full context.

## Step 3: Triage findings

Review all findings from both agents. Categorize each as either **fix** or **dismiss**:

- **Fix**: Real bugs, actual security vulnerabilities, convention violations, misleading error messages, dead code that creates false confidence.
- **Dismiss**: Overengineering suggestions (streaming for small files, rollback for simple copies), pre-existing issues not part of the current change, theoretical risks that require unrealistic attack vectors, adding validation for things that can't happen.

Present the triage: what you're fixing and what you're dismissing (with one-line reasons for dismissals).

## Step 4: Apply fixes

Implement all fixes from the "fix" category. Then run:

```bash
go build ./...
go vet ./...
go test ./... -v
```

All three must pass clean.

## Step 5: Commit and push

If any fixes were applied, commit them with a descriptive message and push to origin. Do NOT tag or create a release.

## Step 6: Loop

If any fixes were applied in Step 4, go back to Step 1 — the fixes themselves may have introduced new issues. Re-identify the changed files (now including your fixes), re-run both reviewers, re-triage, and re-fix.

Repeat until a review pass returns **zero actionable findings** in the fix category. Only then report done.
