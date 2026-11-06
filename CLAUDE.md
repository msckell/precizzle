# Precizzle

Full-stack calculator with exact decimal results for financial use. React + TypeScript frontend, Go REST backend. The backend is the single source of truth for every result.

## Project memory

Read these at the start of every session, before doing anything else:

- `docs/prompts/00-master-prompt.md`: the master prompt (product, workflow, plan in blocks).
- `docs/SPEC.md`: source of truth. API contract, precision and rounding rules, decisions log, out of scope.
- `docs/plans/`: approved block plans, verbatim.
- `PROMPTS.md`: prompt log. Git history.

## Workflow

- One branch per block from an up-to-date main. Everything after the initial commit reaches main through a pull request, merged with a merge commit.
- No application code before the block plan is approved. The first commit on a block branch is the plan in `docs/plans/`.
- Conventional Commits. One logical change per commit, build and tests passing.
- Checkpoints, stop and wait for an OK: end of interview, plan approval, end of each milestone, before push and PR, after the CodeRabbit triage, before merge. Show evidence (files, commands, output excerpts).
- Needs explicit OK: creating the repository, pushing, opening PRs, commenting on GitHub, merging, deleting branches.
- Never force push, rewrite published history or skip checks (`--no-verify`).
- No AI attribution in commits or PRs (no `Co-Authored-By` for assistants, no "Generated with").

## Language

Chat with the owner in Spanish. Everything in the repository or on GitHub is in English: code, comments, commits, branches, PRs, review replies, docs.

## Code and comment rules

- Build only what `docs/SPEC.md` asks for. No extra features, abstractions, configuration or files. The simplest correct and clear solution wins.
- Priorities: correctness, clarity, maintainability, then extra features.
- No binary floating point in the calculation path: not in Go arithmetic, not in JSON parsing, not in the frontend. JavaScript numbers never hold operands or results.
- Validation lives at the boundaries (HTTP input in the backend, user input in the frontend). Internal code trusts its inputs.
- Comments orient, they do not narrate:
  - Go: a one-sentence doc comment on each exported identifier, starting with its name.
  - A short "why" comment where a decision is not obvious (precision, rounding, a validation rule).
  - Nothing else: no comments restating code, no change history, no commented-out code, no TODOs. Same for the frontend.
- Tests check real behavior and edge cases. Go tests are table-driven; arithmetic tests are written first. Never hardcode values to pass a test, never delete or weaken a test to get green.
- Tooling: gofmt and go vet in Go. TypeScript strict mode, ESLint and Prettier in the frontend.
- Delete temporary files and experiments when done.

## Commands

Backend, run from `backend/`:

- Format check: `gofmt -l .` (must print nothing)
- Vet: `go vet ./...`
- Test: `go test ./...`
- Build: `go build ./...`
- Run: `go run .` (listens on `:8080`)
