# Block 1 plan: Backend in Go

## Context

Block 0 defined the product, the number model and the API contract in `docs/SPEC.md` (D1 to D27), and left a Go skeleton: `backend/main.go` serves an empty `ServeMux` on `127.0.0.1:8080`, and `backend/main_test.go` checks a 404. Block 1 builds the complete API with its tests, and a CI workflow that runs the Go checks on every pull request and on `main`.

The interview settled D28 to D35. They are written in the SPEC decisions log, but not committed yet:

| # | Decision |
|---|---|
| D28 | Only `math/big`: `big.Rat` for exact values, `big.Int.Sqrt` for roots, our own operand parser and our own round half even. |
| D29 | Standard `net/http` `ServeMux` with method patterns. Our own `not_found` and `method_not_allowed` in Problem Details. |
| D30 | `backend/main.go`, `backend/internal/calc` (number model) and `backend/internal/api` (HTTP). |
| D31 | Vite dev proxy for `/api`. No CORS in the backend. |
| D32 | `GET /healthz`, outside `/api/`, returns `200 {"status": "ok"}`. |
| D33 | `PRECIZZLE_ADDR`, `127.0.0.1:8080` by default. |
| D34 | CI fails below 95% statement coverage in `internal/calc` or 90% in `internal/api`. |
| D35 | Strict JSON with `encoding/json/v2`: case-sensitive names, duplicate member and trailing data are `invalid_json`. |

Branch: `feat/backend-api`, created from `main` at `a3dd534`. The working tree holds the uncommitted D28 to D35 in `docs/SPEC.md`.

Facts checked during the interview (Go 1.27.3, official docs and throwaway experiments, now deleted):

- `Rat.FloatString` rounds halves away from zero, so it cannot produce `decimal`.
- `Rat.SetString` accepts `+5` and `0x10`, so the operand syntax is validated before it.
- `Rat.FloatPrec` reports whether a rational terminates and how many fractional digits it needs.
- `Int.Sqrt` returns ⌊√x⌋.
- `jsontext.Value.IsValid` checks syntax, UTF-8 and duplicate names.
- The built-in `ServeMux` answers 404 and 405 in `text/plain`.
- Computing `q^1000` for `q` close to 10¹⁰⁰⁰ takes about 43 ms.

## Contract clarifications (SPEC changes for approval)

These fill gaps in the contract. None changes an agreed behavior.

1. **Health endpoint (D32).** Add `GET /healthz` to the API contract. Any other method on `/healthz` returns `405 method_not_allowed` with `Allow: GET, HEAD`. The Go pattern `GET` also matches `HEAD`.
2. **Unknown paths.** Every path without an endpoint returns `404 not_found` in Problem Details, not only paths under `/api/`. Block 3 can revisit this if Go serves the frontend.
3. **Media type.** `Content-Type` is compared after parsing its parameters, so `application/json; charset=utf-8` is accepted. A missing `Content-Type` is `unsupported_media_type`.
4. **Strict JSON (D35).**
   - Field names are case-sensitive: `"A"` is `unknown_field`.
   - Duplicate members, invalid UTF-8 and data after the object are `invalid_json`.
5. **Check order, made precise.**
   - The first unknown field in document order is reported.
   - Missing fields are checked in the order the endpoint table lists them.
   - Operand syntax is checked for every field before any operand limit.
   - In `percentage`, `operator` is checked together with operand syntax, in field order (`percent`, `base`, `operator`). A non-string `operator` is `invalid_operator`.
6. **`field` member** for calculation errors:

   | Error | `field` |
   |---|---|
   | `division_by_zero` in `divide` | `b` |
   | `division_by_zero` in `percentage` | `percent` |
   | `division_by_zero` in `power` (0 raised to a negative exponent) | `base` |
   | `negative_square_root` | `value` |
   | `exponent_not_integer`, `exponent_out_of_range` | `exponent` |
   | `result_too_large`, `result_too_precise` | none |

## Milestone 1: Plan, decisions and contract

Steps:

1. **First commit on the branch:** this plan, verbatim, in `docs/plans/block-1-backend-api.md`, as `docs(plans): add Block 1 plan`.
2. **`PROMPTS.md`:** add the section "3. Block 1: Backend in Go".
   - The starting prompt, verbatim in Spanish, with its English translation.
   - Two interview rounds as tables of selected answers, each with its final decision (D28 to D35).
   - The link to this plan.
3. **Commit:** steps 2 plus the D28 to D35 already in the working tree, as `docs: record Block 1 decisions and prompt log`.
4. **`docs/SPEC.md`:** apply the contract clarifications above in section 4.
   - A "Health" subsection.
   - The conventions about media type and strict JSON.
   - The `field` rules.
   - The refined check order.
   - The `not_found` row now covers every unknown path.

   Commit as `docs(spec): specify health endpoint, strict JSON and check order`.

Verification:

- Cross-read SPEC sections 2 to 4 against D1 to D35: no contradictions.
- In `PROMPTS.md`, the quotes are verbatim and the order is chronological.

## Milestone 2: `internal/calc`, the number model

Package `calc` holds parsing, exact arithmetic, rounding, formatting and limits. It has no HTTP code. Tests are written first, are table-driven, and build their expected long decimals independently of the code (for example `"0." + strings.Repeat("3", 100)`).

### Exported API

- `Number` holds a `*big.Rat` and an `approximate` flag.
- `Kind` is `Exact`, `Repeating` or `Approximate`, rendered as `exact`, `repeating` or `approximate`.
- `Result` holds `Value`, `Decimal` and `Kind`.
- `Parse(s string) (Number, error)` returns `ErrInvalidNumber` or `ErrOperandOutOfRange`.
- `Add`, `Subtract`, `Multiply`, `Divide`, `Power`, `Sqrt` and `Percentage` take `Number`s and return `(Result, error)`.
- Sentinel errors: `ErrDivisionByZero`, `ErrNegativeSquareRoot`, `ErrExponentNotInteger`, `ErrExponentOutOfRange`, `ErrResultTooLarge`, `ErrResultTooPrecise`. `PercentOperator` holds the four operators.

### Key rules

- **Parser.**
  - A hand-written scanner applies the SPEC grammar: decimal, `p/q` and `~decimal`.
  - For decimals, it computes the positions of the first and last significant digits from the mantissa and the exponent before building any `big.Rat`. `1e99999` is rejected as out of range without allocating, while `0e99999` is 0.
  - Exponents with too many digits to fit an `int` are out of range unless the mantissa is zero.
- **Operand and result limits.** One shared function checks, in this order:
  1. |x| < 10⁴⁰, otherwise too large.
  2. A terminating value needs at most 1000 fractional digits (`FloatPrec`).
  3. A repeating value has a reduced denominator of at most 10¹⁰⁰⁰.
  4. A non-zero approximate value does not round to 0 at 1000 fractional digits.
- **Formatting.**
  - **`value`:**
    - Exact terminating: plain digits.
    - Repeating: the reduced `p/q`.
    - Approximate: `~` followed by the `decimal`.
    - In every case, no `-0`, no leading zeros and no trailing fractional zeros.
  - **`decimal`:**
    - Exact: every digit.
    - Repeating: rounded half even to 100 fractional digits.
    - Approximate: rounded half even to 20 fractional digits.
  - **Tiny values.**
    - When 0 < |x| < 10⁻²⁰, the result uses scientific notation, and its magnitude e = ⌊log₁₀|x|⌋ is computed exactly with big integers.
    - Rounding happens at fractional digit −e+20 (approximate) or −e+100 (repeating). For approximate results the position is capped at 1000.
    - A mantissa that rounds up to 10 is renormalized.
- **Round half even:** n = |p|·10ᶠ divided by q. Round up when 2·remainder > q, or when it equals q and the quotient is odd.
- **Power.**
  - Exponent rules come first: the exponent must be exact, an integer and between −1000 and 1000.
  - Then 0 raised to a negative exponent is a division by zero, and 0⁰ = 1. A negative exponent inverts the base.
  - Two cheap bit-length bounds run before the full exponentiation. Each one rejects only results that would certainly be rejected:
    - A lower bound on |x|ᵉ above 10⁴⁰ returns `result_too_large`.
    - When |x|ᵉ is certainly below 10⁴⁰ and qᵉ is certainly above 10¹⁰⁰⁰, the result is `result_too_precise`.
  - Otherwise the power is computed exactly. That costs at most a few tens of milliseconds, measured in a benchmark.
- **Square root.**
  - Negative input is an error.
  - If the reduced p and q are perfect squares (`Int.Sqrt`, then squaring back), the root is exact, and it is still `approximate` if the operand was.
  - Otherwise the root is computed correctly rounded at the position its magnitude requires. With y = x·10²ᶠ and s = `Int.Sqrt`(⌊y⌋), round up when (2s+1)²·q < 4·p·10²ᶠ. A tie is impossible for an irrational root.
- **Percentage:** the D7 formulas. A percent of 0 with `divide` is a division by zero.
- **Approximation spreads:** a result is approximate when any operand is.

### Tests

- Every row of the SPEC examples and errors tables (§3), as one table-driven test that names the row.
- **Parser:**
  - Each accepted form.
  - Each rejected form: `+1`, `.5`, `5.`, `1,5`, `1 000`, `0x10`, `1/0`, `-1/-3`, `~1/3`, `~~1`, an empty string.
  - Limit boundaries: `1e39`, `9.99…e39`, `1e40`, 1000 and 1001 fractional digits, a denominator of 10¹⁰⁰⁰ and just above it.
  - `0e99999` and `1e99999`.
- **Canonical form:** `-0`, `1.10`, `4/2`, `-2/4` and `0.50e1`.
- **Rounding:**
  - Half even ties (`0.125` → `0.12`, `0.135` → `0.14`) at the 20th and the 100th digit.
  - The renormalized mantissa case.
  - The 1000-digit cap for approximate values.
- **Power:** `0^0`, `0^-1`, `2^1001`, `2^-1001`, `0.25^501`, `1.0001^-2000`, `1e-20^-2`, `(1/3)^-2`, an approximate base, and the exponents `2.0`, `4/2`, `1e2`, `2.5` and `~2`.
- **Square root:**
  - √2 at 20 digits, checked against an independently written digit string.
  - √2.25, √(1/9), √0, √1e-1000, √(1/3), √~4 and a negative input.
- **Percentage:** the five D7 rows, the divide-by-0% case and the approximate cases.
- **Round trip:** every `Result.Value` parses back to the same `Number`, so chaining always works. A fuzz target (`FuzzParse`) runs its seed corpus under `go test`.
- **Benchmark:** `BenchmarkPowerWorstCase` with the largest denominator raised to 1000.

### Commits

1. `feat(calc): parse operands and format canonical results`, with the limits, rounding and scientific notation.
2. `feat(calc): add addition, subtraction, multiplication and division`.
3. `feat(calc): add integer powers`.
4. `feat(calc): add correctly rounded square roots`.
5. `feat(calc): add percentages`.
6. `test(calc): cover every example in the specification`.

Each commit includes its own tests, written before the code.

Verification: from `backend/`, `gofmt -l .` prints nothing, and `go vet ./...` and `go test -cover ./internal/calc` pass with coverage ≥ 95%. The output of `go test -bench=PowerWorstCase -run=^$ ./internal/calc` stays well below 100 ms per operation.

## Milestone 3: `internal/api`, the server and configuration

### `internal/api`

- `NewHandler() http.Handler` builds the `ServeMux`:
  - `POST /api/v1/{add,subtract,multiply,divide,power,sqrt,percentage}`: seven explicit patterns.
  - The same paths without a method answer `405 method_not_allowed` with `Allow: POST`. A pattern with a method is more specific, so it wins for `POST`.
  - `GET /healthz`, and `/healthz` without a method answers `405` with `Allow: GET, HEAD`.
  - `/` answers `404 not_found` for every other path.
- **Problem Details:** one `writeProblem` helper writes `type`, `title`, `status`, `detail`, `code` and an optional `field` with `application/problem+json`. A single table maps each code to its status and title, and the SPEC is the source of that table. Successes are written as `application/json`.
- **Request pipeline:** checked in the SPEC order.
  1. Media type with `mime.ParseMediaType` (415).
  2. Body read through `http.MaxBytesReader` at 64 KiB (413, detected with `*http.MaxBytesError`).
  3. Syntax and duplicates with `jsontext.Value.IsValid`, and the top-level value must be an object (`invalid_json`).
  4. Members walked in document order with `jsontext.Decoder` (`unknown_field`).
  5. Required fields; `null` counts as missing (`missing_field`).
  6. Each field must be a JSON string, then `calc.Parse`. Every field's syntax is checked first (`invalid_number`, `invalid_operator`), then every field's limits (`operand_out_of_range`).
  7. The calculation. `calc` errors are mapped to codes and fields with `errors.Is`.
- One small handler per endpoint names its fields and calls one `calc` function. The shared pipeline lives in one function that receives the field list.

### `backend/main.go`

- It reads `PRECIZZLE_ADDR` through a small `listenAddr(getenv func(string) string) string` with the default `127.0.0.1:8080`.
- It builds an `http.Server` with `Handler: api.NewHandler()` and timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`), then calls `log.Fatal(srv.ListenAndServe())`.
- The 404 test in `main_test.go` moves to `internal/api`. `main_test.go` keeps a table-driven test of `listenAddr`.

### Tests (`httptest`, table-driven)

- Each of the 16 error codes, with its status, `Content-Type`, `code`, `title`, `type` URI and `field`. The `405` responses also check `Allow`.
- Check order: requests with two problems at once return the first one in SPEC order. Cases include an unknown field together with a JSON number, `b` invalid together with `a` out of range, and the operator together with an operand.
- **Strict JSON:**
  - `"A"`, a duplicate `"a"`, `{}` followed by more data, a top-level array, `null`, an empty body and invalid UTF-8.
  - A JSON number, `true`, an object and `null` as field values.
- **Media type:** none, `text/plain`, and `application/json; charset=utf-8`.
- The body limit at exactly 64 KiB and at 64 KiB + 1.
- One success per endpoint, plus the D21 chain: `divide 1 3`, then `multiply` with the returned `value` and `3`, returns `"1"`.
- `GET` and `HEAD /healthz`, and `POST /healthz` returning 405.

### Commits

1. `feat(api): write Problem Details errors and decode strict JSON requests`.
2. `feat(api): add the calculation endpoints`.
3. `feat(api): add the health endpoint`.
4. `feat(backend): configure the listen address and server timeouts`.

Verification:

- `gofmt -l .` prints nothing, and `go vet ./...` and `go test -cover ./...` pass, with `internal/api` at 90% or more.
- Manual smoke test: `go run .` in the background, then `curl.exe` for 0.1 + 0.2, 1 ÷ 3, the chain to 1, √2, 200 + 15%, a 422, a 404, a 405 and `/healthz`. The checkpoint shows each response.
- Start with `PRECIZZLE_ADDR=127.0.0.1:8081` and confirm the server listens on the new port.

## Milestone 4: CI and documentation

1. **`.github/workflows/ci.yml`**, job `backend`:
   - Runs on `pull_request` and on `push` to `main`. `working-directory: backend`, `permissions: contents: read`.
   - `actions/checkout@v7`, then `actions/setup-go@v7` with `go-version-file: backend/go.mod` and `cache: false`. There is no `go.sum`, because the backend has no dependencies.
   - Steps: the gofmt check (fail if `gofmt -l .` prints anything), `go vet ./...`, and `go test -race -cover ./...`.
   - The D34 gate parses each package's coverage line and fails below 95% (`internal/calc`) or 90% (`internal/api`). The percentages also go to the job summary.

   Commit as `ci: run gofmt, go vet and tests with coverage for the backend`.
2. **`CLAUDE.md`, Commands:** add coverage (`go test -cover ./...`), the benchmark and `PRECIZZLE_ADDR`. Commit as `docs: update backend commands in CLAUDE.md`.
3. **`README.md`, a "Backend API" section:**
   - How to run the API and the `PRECIZZLE_ADDR` variable.
   - The endpoint list, with `curl` examples for 0.1 + 0.2, 1 ÷ 3, the chain, √2 and an error.
   - The health check.
   - A note that the frontend reaches the API through the Vite proxy (D31), configured in Block 2.
   - Remove the line "The API has no endpoints yet".

   Commit as `docs: document the backend API in the README`.

Verification:

- Run the workflow's shell steps locally in Git Bash: the gofmt check, the vet and the coverage gate.
- Prove that the gate fails by running it with a threshold of 100%.
- The workflow itself is verified when it runs on the pull request.

## Pull request (checkpoints 4 to 6)

- Title: `feat: Block 1 backend API with exact rational arithmetic`.
- Description: summary, motivation, how it was tested (test, coverage and smoke-test output), decisions D28 to D35 and the contract clarifications, and what is out of scope (the frontend, the Vite proxy and Docker).
- Then the CodeRabbit protocol. At close, `PROMPTS.md` gets the PR prompts and the CodeRabbit triage.

## Out of scope for this block

- The Vite proxy and any change in `frontend/` (Block 2).
- Docker and the container listen address (Block 3).
- Coverage artifacts and a coverage report in the README (Block 4).
- Logging, metrics, graceful shutdown, rate limiting and CORS.

## End-to-end verification for the block

- From `backend/`: `gofmt -l .` prints nothing, `go vet ./...` passes, `go test -race -cover ./...` passes on CI with `calc` at 95% or more and `api` at 90% or more, and `go build ./...` passes.
- Every row of the SPEC examples and errors tables has a test in `calc`, and every error code has a test in `api`.
- The manual `curl` session returns exactly the SPEC examples, including `0.1 + 0.2 = 0.3` and `1 ÷ 3 × 3 = 1`.
- The CI workflow is green on the pull request.
