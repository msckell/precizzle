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
- Coverage: `go test -cover ./...` (CI fails below 95% in `internal/calc` or `internal/expr`, or 90% in `internal/api`)
- Benchmarks: `go test -run='^$' -bench=PowerWorstCase ./internal/calc` and `go test -run='^$' -bench=EvaluateWorstCase ./internal/expr`
- Fuzz: `go test -run='^$' -fuzz=FuzzEvaluate -fuzztime=60s ./internal/expr` (CI runs only the seed corpus)
- Build: `go build ./...`
- Run: `go run .` (listens on `127.0.0.1:8080`; set `PRECIZZLE_ADDR` to change it)

CI (`.github/workflows/ci.yml`) runs the gofmt check, `go vet` and `go test -race -cover` on every pull request and on `main`.

Frontend, run from `frontend/` after `npm ci`:

- Lint: `npm run lint`
- Format check: `npm run format:check` (fix with `npm run format`)
- Typecheck: `npm run typecheck`
- Test: `npm test`
- Coverage: `npm run coverage` (fails below 90% of statements)
- Build: `npm run build`
- Run: `npm run dev` (serves on `localhost:5173` and proxies `/api` to the backend, which must be running on `127.0.0.1:8080`)

CI runs `npm ci`, lint, the format check, typecheck, coverage and build in the `frontend` job.

Docker, run from the repository root:

- Build: `docker build -t precizzle .` (multi-stage; the final `alpine` image runs as a non-root user)
- Run: `docker run --rm -d -p 127.0.0.1:8080:8080 --name precizzle precizzle` (app and API on `http://127.0.0.1:8080`)
- Stop: `docker stop precizzle`
- In the image the server listens on `:8080` and serves the frontend build from `PRECIZZLE_STATIC_DIR` (D88). Without it, the server answers only the API.

CI's `docker` job builds the image, waits until its health check passes, and checks that `0.1+0.2` returns `0.3` and that `/` serves the app.
