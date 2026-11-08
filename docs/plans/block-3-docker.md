# Block 3 plan: Docker

## Context

Blocks 0 to 2 are merged: main is at `bebd62c` (PR #5), CI is green, the backend tests pass and the frontend has 408 tests at 99.01% statement coverage. The master prompt still asks for "a Dockerfile that runs the frontend and the backend together", and the definition of done says that the documented Docker commands build and start the whole app.

Block 3 adds that, without changing the API or any approved behavior. Branch: `chore/docker`, created from `bebd62c` in `C:\Projects\precizzle`.

The interview settled these decisions. They go into a new "Block 3: Docker" table of the SPEC, from D88:

| # | Decision |
|---|---|
| D88 | One image. The Go server also serves the frontend build. When `PRECIZZLE_STATIC_DIR` is set, `GET` and `HEAD` requests outside `/api/` and `/healthz` are served from that directory with the standard file server, and a missing file is its plain `404`. When the variable is unset or empty, the server behaves exactly as today. The API, `/healthz` and every Problem Details response stay unchanged. |
| D89 | Multi-stage build: `node:24.15-alpine` builds the frontend, `golang:1.27.3-alpine` builds a static binary (`CGO_ENABLED=0`), and the final stage is `alpine:3.23` with only the binary and `dist/`. It runs as a non-root user. Inside the container `PRECIZZLE_ADDR` is `:8080` (D27), and the documented command publishes it only on the host loopback: `-p 127.0.0.1:8080:8080`. |
| D90 | A Dockerfile `HEALTHCHECK` calls `GET /healthz` (D32) with busybox `wget`, which `alpine` already has. |
| D91 | A `docker` CI job builds the image, starts it, waits until it is healthy, checks that `0.1 + 0.2` returns `0.3` through `evaluate`, and checks that `/` serves the app's HTML. |

## Design

### Backend: `backend/main.go`

- A new unexported `newHandler(staticDir string) http.Handler`:
  - Empty `staticDir`: returns `api.NewHandler()` unchanged.
  - Otherwise, a `ServeMux` that sends `/api/` and `/healthz` to the API handler, `GET /` to `http.FileServerFS(os.DirFS(staticDir))`, and `/` (other methods) to the API handler, so they keep its Problem Details `not_found`.
- `main` reads `PRECIZZLE_STATIC_DIR` with the same `getenv` style as `listenAddr` (D33). A short "why" comment points to D88.
- `internal/api` does not change.

### Docker: `Dockerfile` and `.dockerignore` at the root

- Stage `frontend`: copy `package.json` and `package-lock.json`, `npm ci`, copy `frontend/`, `npm run build`.
- Stage `backend`: copy `backend/`, `go build -trimpath -ldflags="-s -w"`.
- Final stage: create a non-root user, copy the binary and `dist/`, set `PRECIZZLE_ADDR=:8080` and `PRECIZZLE_STATIC_DIR`, `EXPOSE 8080`, `HEALTHCHECK` (D90), `USER`, `ENTRYPOINT`.
- `.dockerignore` keeps `node_modules`, `dist`, `coverage`, `.git` and the docs out of the build context.

### CI: `.github/workflows/ci.yml`

- A new `docker` job after `frontend`. It builds with `docker build`, runs the container on `127.0.0.1:8080`, polls `docker inspect` until the status is `healthy` (with a time limit), runs the two `curl` checks of D91 and prints `docker logs` if any step fails. The `backend` and `frontend` jobs do not change.

## Milestones

### Milestone 1: the backend serves the frontend build

Steps and files:
1. `docs/SPEC.md`: the "Block 3: Docker" table with D88 to D91.
2. `backend/main.go` and `backend/main_test.go`: `newHandler` and its table-driven tests, using `httptest` and a `t.TempDir()` holding `index.html`:
   - Without a static directory: `GET /` is `404 not_found` in Problem Details, as today.
   - With a static directory:
     - `GET /` and `HEAD /` serve `index.html`.
     - `GET /assets/x.js` serves the file.
     - A missing file is `404`.
     - `POST /api/v1/evaluate` with `0.1+0.2` returns `0.3`.
     - `GET /healthz` returns `ok`.
     - `GET /api/v1/add` is `405`, and `/api/v1/nope` is the API's `not_found`.
     - `POST /` is the API's `not_found`.

Verification: `gofmt -l .`, `go vet ./...`, `go test -race -cover ./...` (the thresholds of D34 and D51 hold). Also a manual run with `PRECIZZLE_STATIC_DIR=../frontend/dist go run .`, then the UI in the browser.

### Milestone 2: the image

Steps and files: `Dockerfile` and `.dockerignore`.

Verification:
- `docker build -t precizzle .` and the image size from `docker image ls`.
- `docker run --rm -d -p 127.0.0.1:8080:8080 --name precizzle precizzle`.
- `docker inspect` shows `healthy`.
- `docker exec precizzle id` shows a non-root user.
- The README `curl` examples work against the container, including `0.1+0.2` → `0.3`.
- In the browser, `http://127.0.0.1:8080` computes, expands and copies a result, and loads the font with no external requests.

### Milestone 3: CI and documentation

Steps and files:
1. `.github/workflows/ci.yml`: the `docker` job (D91).
2. `README.md`: a short "Docker" section with the build and run commands, the port and the health check.
3. `CLAUDE.md` and `AGENTS.md`, which are identical: the Docker commands and the CI job. They are at 79 lines and the master prompt keeps them under 80, so this needs room. **For the owner:** replace the "Block 2: parallel work" section, obsolete since PR #5, with the Docker commands. If you prefer, the removal waits for Block 4 and the Docker commands go in compactly.
4. `PROMPTS.md`: a new §5 "Block 3: Docker" with the starting prompt verbatim, the interview round and a link to this plan.

Verification:
- `docker build` from a clean checkout.
- The smoke test script of the job, run locally with Git Bash.
- After the push, the green `docker` job is the final proof.

## Out of scope

- Docker Compose, nginx, TLS, a reverse proxy and publishing the image to a registry.
- SPA fallback routing (the app has one page), cache headers and compression.
- Any change to the API, the frontend code or the approved decisions.

## After the plan

Plan commit (`docs: add the Block 3 plan`), then milestones with checkpoints. Then the push and PR with your OK, the CodeRabbit triage, and the merge with your OK. The PR #5 merge record in `PROMPTS.md` §4.5 stays for Block 4, as agreed.
