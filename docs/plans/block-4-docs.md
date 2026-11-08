# Block 4 plan: Documentation and wrap-up

## Context

Blocks 0 to 3 are merged. main is at `3cfe823` (PR #6), and CI is green with the Backend, Frontend and Docker jobs. The application is complete. Block 4 finishes the documentation, adds the final coverage report, records the two pending merges, and checks the repository against the master prompt's definition of done. It adds no features and does not reopen any approved decision.

Branch: `docs/readme`, created from `3cfe823` in `C:\Projects\precizzle`.

The interview settled these decisions. D92 and D93 go into a new "Block 4: Documentation" table of the SPEC:

| # | Decision |
|---|---|
| D92 | The coverage report has two parts: a summary table in the README and the existing CI output. The table lists each Go package and the frontend statements, with the CI minimum for each. The CI output is the backend job's step summary and the frontend's Vitest text report. No new tooling or artifacts. |
| D93 | The README keeps its current sections. A short "Quick start" (Docker first, then the tests) replaces "Status: work in progress", and outdated lines are fixed. |
| — | After the Block 4 merge, the merged Block 2 branches and worktrees are deleted, locally and on origin: `feat/expression-api` (`C:\Projects\precizzle-expression-api`) and `feat/frontend` (`C:\Projects\precizzle-frontend`). This is repository housekeeping, so it goes in `PROMPTS.md`, not in the SPEC. |

## Milestones

### Milestone 1: README

Steps, all in `README.md`:

1. "Quick start" replaces "Status". It covers Docker build and run, the URL, `0.1 + 0.2` with `curl`, and how to run all the tests. The Development requirements also mention Docker.
2. A "Tests and coverage" summary table (D92), with the current numbers taken from the commands, not from memory:
   - Backend: `calc`, `api`, `expr` and `main`.
   - Frontend: statements and number of tests.
   - The CI minimum for each.
3. Fix outdated lines:
   - "Block 2 configures the proxy."
   - Any test count or wording that no longer matches.
   - Small inconsistencies between the README and the SPEC found in Milestone 2, such as the Escape key closing the history.
4. "Project documents" gets the master prompt and the Dockerfile.

Verification:
- Run every command in the README as written: Docker build and run, the 13 `curl` examples, the backend and frontend test, coverage and build commands, and `npm run dev` with `go run .` against the dev proxy.
- Every number in the coverage table is copied from that output.
- Every relative link in the README points to an existing file or anchor, checked with a small throwaway script.

### Milestone 2: SPEC review and definition of done

Steps:

1. Review `docs/SPEC.md` section by section against the code:
   - §1 operations and user interface.
   - §2 to §5 rules, limits and examples.
   - The errors table.
   - The out of scope list.
2. Mechanical checks:
   - Every error `code` in the SPEC table has a backend test and a frontend message (`frontend/src/messages.ts`).
   - Every limit constant in the code matches the SPEC.
   - Every example in §3 and §5 exists as a backend test or fixture.
3. Wording or documentation mismatches are fixed in the SPEC or the README. **Any difference in behavior is reported to the owner and not changed.**
4. `docs/SPEC.md`: the "Block 4: Documentation" table with D92 and D93.
5. Definition of done (master prompt §10): a short checklist with evidence for each item, kept for the checkpoint and the PR description, not as a new file.

Verification: the review findings, each with the code or test that confirms it, and the definition of done checklist, all at the checkpoint.

### Milestone 3: PROMPTS.md and final checks

Steps, all in `PROMPTS.md`:

1. §4.5: a "##### Merge" section for PR #5. It holds the owner's authorization of 2026-11-07, verbatim in Spanish with its English translation, taken from the earlier session's transcript. It also records merge commit `bebd62c` and the post-merge checks:
   - main in sync with origin.
   - The backend checks and the frontend lint, format, typecheck, coverage (408 tests, 99.01%) and build.
   - The 70 fixtures matched the real backend.
   - CI on main passed.
2. §5.4: a "#### Merge" section for PR #6. It holds the owner's authorization ("Autorizo") and records:
   - Merge commit `3cfe823`.
   - `chore/docker` deleted locally and on origin.
   - CI on main passed.
3. A new §6 "Block 4: Documentation and wrap-up":
   - The interview round, the plan link and the Block 4 prompts.
   - The starting prompt is the same session prompt logged in §5.1, so §6 links to it instead of copying it.

Verification:
- The full local suite on the branch: gofmt, vet, Go tests with coverage, and frontend lint, format check, typecheck, coverage and build.
- `docker build`.
- `git diff --check`.
- CLAUDE.md and AGENTS.md stay identical and under 80 lines.

## After the plan

The plan commit comes first (`docs: add the Block 4 plan`), then the milestones with their checkpoints. Then:
- Push and the PR, with your OK.
- The CodeRabbit triage.
- The merge, with your OK.
- Sync main and run the branch and worktree cleanup, which you already approved in the interview. The cleanup first checks that each worktree is clean and each branch is merged.

## Out of scope

- New features, API or UI changes, and any change to an approved decision.
- New tooling, coverage artifacts or badges.
- Rewriting README sections that were already reviewed, beyond fixing outdated lines.
