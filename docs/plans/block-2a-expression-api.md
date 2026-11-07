# Block 2a plan: Expression API

## Context

Block 2 adds free expressions with precedence. The frontend sends every calculation to a new endpoint, `POST /api/v1/evaluate` (D41). The shared base, `docs/block-2-contract` and PR #3, holds D39 to D50 and the contract in SPEC §5. The base also has two fixes from the CodeRabbit review of PR #3:

- A missing operand inside an unclosed group is reported first, so `(2+` gives position 3.
- D41 now says that `position` replaces `field` only in errors about the expression.

Block 2a builds the backend side: tokenizer, parser, evaluator, endpoint, tests, CI and documentation. It works only in `backend/` and in the shared files assigned to 2a in `CLAUDE.md`. The seven operation endpoints keep their exact behavior.

Branch: `feat/expression-api`, at `863e268`, in the worktree `C:\Projects\precizzle-expression-api`.

The 2a interview settled these decisions. They go into the "Block 2a" table of the SPEC:

| # | Decision |
|---|---|
| D51 | The tokenizer, the recursive descent parser and the evaluator live in a new package, `backend/internal/expr`. It calls the operations of `calc` (D44). CI fails below 95% statement coverage in `internal/expr`, the same as `internal/calc` (D34). |
| D52 | `BenchmarkEvaluateWorstCase` measures the costliest expression that fits in the 64 KiB body: 10 worst-case powers and 90 more operations. It runs by hand, like `BenchmarkPowerWorstCase`. `FuzzEvaluate` runs its seed corpus under `go test`; CI does no open-ended fuzzing. |
| D53 | `expr` returns an `*expr.Error` that holds a cause and a `position`. The cause is a sentinel error of `calc` or of `expr` (`ErrInvalidExpression`, `ErrExpressionTooComplex`). `api` maps the cause with `errors.Is`, as it does today, and writes `position` in place of `field`. Any other error is `internal_error` (D37). |
| D54 | Position rules that §5 leaves open. See below. |

### D54: position rules that §5 leaves open (for approval)

None of these changes a SPEC example. Each follows "the first character that cannot continue the expression" and the left-to-right order of check 8.

1. A token is validated only when the grammar accepts it at that point. `2 1.2.3` is `invalid_expression` at 2, because no number can follow `2`. `1.2.3` is `invalid_number` at 0.
2. `sqrt(` is matched character by character, and the error points to the first character that differs. `sqrx(4)` points to 3, `sqrt (4)` to 4 (the space), and `sqrt` at the end of the expression to 4, the end. A letter that cannot start any token points to itself.
3. With several unclosed groups, the innermost is reported, because it is met first. `((2+3` points to 1.
4. When both limits of D45 are broken, the earlier operator is reported: the 101st operation or the 11th power, whichever comes first in the text.

## Design

### `internal/calc` (one addition)

- `func (n Number) Result() Result` returns the canonical result of an operand. It is needed when an expression has no operation, such as `5.0`, `[1/7]` or `((2))`. A parsed `Number` is already within the limits, so the method cannot fail.
- No other change. Every operation still goes through `Add`, `Subtract`, `Multiply`, `Divide`, `Power`, `Sqrt`, `Percentage` and `ApplyPercentage`.

### `internal/expr`

- **`Evaluate(s string) (calc.Result, error)`** is the only entry point. It applies checks 8 to 11 of SPEC §5, in this order:
  1. Syntax and nesting.
  2. Operand limits.
  3. The operation count.
  4. Evaluation.
- **Scanner (lazy).** The parser asks for one token at a time, so the first problem from left to right is reported, even when a later character is invalid too.
  - Only spaces (U+0020) can separate tokens. Any other character outside the syntax (a tab, a non-ASCII byte) is `invalid_expression` at its index.
  - A number token starts with a digit or `.`. It runs over digits, `.`, `e` and `E`, and a `+` or `-` right after `e` or `E`. When the parser takes it as an operand, it reads it with `calc.Parse`:
    - `ErrInvalidNumber` gives `invalid_number` at the start of the token.
    - `ErrOperandOutOfRange` is recorded and reported in check 9.
  - A value token runs from `[` to the first `]` and is read with `calc.Parse`. With no `]`, the error is `invalid_expression` at the `[`.
  - The `sqrt(` token follows D54.2.
- **Parser.** It implements the grammar of §5 directly: `sum`, `term`, `signed`, `power`, `postfix` and `primary`.
  - Unary minus is read in a loop. `^` is right-associative through recursion; the 64 KiB body limit and Go's growing stacks keep that recursion bounded.
  - `(` and `sqrt(` raise a depth counter. Opening level 101 is `expression_too_complex` at the opening character.
  - The parser marks each term of a sum that is only a percentage: B% with or without minus signs in front, and nothing else (D43). It decides from the syntax, so `(15%)` or `15%*2` is never one.
  - Each operator records its position: `+ - * / ^ %`, unary minus, and the `s` of `sqrt(`. After the syntax and the limits pass, one walk in text order applies D45 (D54.4).
- **Evaluator.** It walks the tree. Every operation calls one `calc` function, and its result goes back in through `calc.Parse(result.Value)`, so an approximate intermediate result enters as its `value` (D44).
  - Unary minus is `calc.Subtract(0, x)`, which keeps the approximate mark and the limits.
  - A plain `%` is `calc.Percentage`.
  - In a sum, each term is evaluated from left to right. When the terms that follow are only percentages, each one applies to the term before it with `calc.ApplyPercentage(T, PercentAdd or PercentSubtract, B)`. The percent B carries the minus signs written in front of it, so `200+-15%` = 170. Only then is the term added to or subtracted from the running total. This matches "an operation runs as soon as its operands are known", because the parser already knows which terms are percentages.
  - Error positions follow the §5 table: `/` for `division_by_zero`, `^` for power errors and for 0 raised to a negative exponent, the `s` for `negative_square_root`, and the operator, or the `%` of T ± B%, for limit errors.
  - A `calc.Parse` failure on a returned `value` would be a bug. It is returned as a plain error, so `api` answers `internal_error` and logs it.

### `internal/api`

- `"/api/v1/evaluate": evaluate()` joins the operations map, so the 405 with `Allow: POST` comes from the existing loop.
- `evaluate()` uses `operation` with the allowed fields `["expression"]` and `missing(f, "expression")`. A value that is not a JSON string is `invalid_expression` with `field: "expression"`. Then `expr.Evaluate` runs.
- `problem` gains a position, written as `"position": <int>` only when it is set, because 0 is a valid position. `field` and `position` never appear together.
- The new codes `invalid_expression` and `expression_too_complex` go into `problemTypes` with status 400 and the titles from the SPEC.
- Each `detail` names the position, for example "The expression divides by zero at position 1.", as in the §5 example.

## Milestone 1: Plan, decisions and prompt log

1. First commit: this plan, verbatim, in `docs/plans/block-2a-expression-api.md`, as `docs(plans): add Block 2a plan`.
2. `docs/SPEC.md`: rows D51 to D54 in the "Block 2a" table only.
3. `PROMPTS.md` §4.4, with nothing written outside it:
   - The starting prompt of this session, verbatim.
   - The PR #3 triage and its outcome.
   - The interview as a table of selected answers with the final decisions.
   - The link to this plan.

   Steps 2 and 3 go in one commit, `docs: record Block 2a decisions and prompt log`.
4. **Gate before any code:** PR #3 is merged with the owner's OK. Then `git merge main` into `feat/expression-api` (a merge, never a rebase), so the branch has the final contract, including `(2+` → 3.

Verification: SPEC §5 against D51 to D54 shows no contradictions, D51 to D54 touch only the 2a table, and `git diff main` after the merge shows only 2a files.

## Milestone 2: `internal/expr`

Tests come first in each commit and are table-driven. Expected values are written by hand from exact arithmetic, never copied from the code's output.

### Tests

- **SPEC §5:** every row of both example tables (21 values and 12 errors, `(2+` → 3 included), each case named after its row. Every example in the "Meaning" text: `50-200-15%` = −120, `200-15%*2` = 199.7, `2^50%` → `exponent_not_integer`, `-2^2`, `(-2)^2` = 4 and `8/4/2` = 1.
- **D7 equivalence:** for each D7 case written as an expression, the result equals `calc.ApplyPercentage` or `calc.Percentage` on the same operands.
- **Percentages (D43):**
  - Compounding and runs of three percentages.
  - A percentage as the first term (`15%+10%` = 0.165).
  - Minus signs in front (`200--15%` = 230).
  - A percentage of a `sqrt(...)` or a value (`200+[1/2]%`).
  - `(15%)` and `200+(15%)`, which are not pure percentages.
  - `200/0%` → `division_by_zero` at the `/`.
  - A T ± B% that breaks a limit points to the `%`.
- **Syntax and positions:**
  - Every error code with its position.
  - Spaces between tokens, and a space that splits a token.
  - Tab, non-ASCII, an empty expression, only spaces, `)`, `()`, `2)` and `2 3`.
  - Each D54 rule.
  - Unclosed `(`, `sqrt(` and `[`. `[ 1 ]`, `[]` and `[1/0]` are `invalid_number` at the `[`.
- **Limits:**
  - Exactly 100 operations and 10 powers pass. The 101st operation and the 11th power fail at their operator.
  - Unary minus and `sqrt` count as operations.
  - Nesting of exactly 100 levels passes. Level 101 fails at its opener, with `(` and with `sqrt(`.
- **Check order:**
  - A syntax error after an out-of-range operand reports the syntax error.
  - An out-of-range operand comes before `expression_too_complex`, and that code comes before any evaluation error.
  - Evaluation stops at the first failing operation.
- **D44:** `10^50/10^20` → `result_too_large` at 2. `sqrt(2)*sqrt(2)` = `~2`. An approximate value keeps spreading. `[1/7]*7` = 1.
- **`calc`:** `Number.Result` on an exact value, a repeating value, an approximate value and `-0`.

### Fuzz and benchmark

- **`FuzzEvaluate`.** Its seeds are the SPEC examples and the D54 cases. For any input:
  - It does not panic.
  - Every error is an `*expr.Error` with 0 ≤ position ≤ len(s) and a known cause, so it never maps to `internal_error`.
  - Every success has a `Value` that `calc.Parse` accepts.
- **`BenchmarkEvaluateWorstCase`.** 10 full powers of `[~1.000…0001]` (1000 fractional digits) raised to 1000, the costliest path of `BenchmarkPowerWorstCase`. Then 90 more operations on the largest operands that fit. Before timing, it checks that the body fits in 64 KiB and checks the expected result.

### Commits

1. `feat(calc): write the canonical result of an operand`
2. `feat(expr): parse expressions with positions and limits`
3. `feat(expr): evaluate expressions with precedence and percentages`
4. `test(expr): add a fuzz target and a worst-case benchmark`

Verification, from `backend/`:

- `gofmt -l .` prints nothing.
- `go vet ./...` passes.
- `go test -cover ./...` shows `internal/expr` at 95% or more, and `calc` and `api` keep 100%.
- `go test -run='^$' -bench=EvaluateWorstCase ./internal/expr` reports about 1 s, well below the 10 s `WriteTimeout`.

## Milestone 3: Endpoint, CI and documentation

1. **`feat(api): add the evaluate endpoint`.** `evaluate.go`, the position in `problem.go`, the route in `api.go`, and `evaluate_test.go`, written first. Tests:
   - Each §5 success and error through HTTP: status, `Content-Type`, `type`, `title`, `code`, and `position` present with `field` absent.
   - `missing_field` when `expression` is absent or null, and `unknown_field`.
   - `invalid_expression` with `field: "expression"` for a number, `true`, an object and an array.
   - `GET /api/v1/evaluate` → 405 with `Allow: POST`.
   - The D21 chain: `1/3`, then `[value]*3` returns `1`.
   - The §5 error example, byte for byte in its members.
   - Every existing api test stays unchanged and passing.
2. **`ci: require 95% coverage in internal/expr`.** Add `internal/expr 95` to the threshold list of the `backend` job, and change nothing else.
3. **`docs: update backend commands in CLAUDE.md`.** In the backend commands only: the coverage line names `internal/expr` at 95%, and the benchmark line adds `go test -run='^$' -bench=EvaluateWorstCase ./internal/expr`.
4. **`docs: document the evaluate endpoint in the README`.** In "Backend API" only:
   - The endpoint and its syntax in short, with a link to SPEC §5.
   - `curl` examples: `50+200+15%`, `1/3*3`, chaining with `[1/7]*7`, `sqrt(2)*sqrt(2)`, and one error with `position`.

Verification:

- The CI steps run locally in Git Bash: the gofmt check, vet, `go test -race -cover` and the threshold gate. The gate fails when the threshold is set to 100%.
- Manual smoke test: `go run .`, then `curl.exe` for every README example, a 400 with `position`, a 405 and `/healthz`. The checkpoint shows each response.

## Pull request (checkpoints 4 to 6)

- Title: `feat: Block 2a expression API`.
- Description: summary, motivation, how it was tested (test, coverage, benchmark and smoke test output), decisions D51 to D54, and what is out of scope.
- Then the CodeRabbit protocol. At close, `PROMPTS.md` §4.4 gets the PR prompts and the triage.
- Merge order: PR #3, then this PR, then 2b.

## Out of scope

- Any change in `frontend/`, outside the 2a parts of the shared files, or to SPEC §2 to §5 and D39 to D50.
- New operations, implicit multiplication, functions other than `sqrt`, and variables.
- Caching, timeouts per request beyond the server's, rate limiting and CORS.
- Docker (Block 3) and the coverage report (Block 4).

## End-to-end verification for the block

- From `backend/`: `gofmt -l .` prints nothing, `go vet ./...` passes, `go test -race -cover ./...` passes, and the coverage is `calc` ≥ 95%, `api` ≥ 90% and `expr` ≥ 95%.
- Every row of the §5 tables has a named test in `expr` and an HTTP test in `api`.
- `curl` returns `280` for `50+200+15%`, `-180` for `50-200+15%`, `253` for `200+15%+10%` and `287.5` for `(50+200)+15%`.
- The seven operation endpoints pass their unchanged tests.
- CI is green on the pull request.
