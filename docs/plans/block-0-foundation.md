# Block 0 plan: Foundation

## Context

Precizzle is a calculator for financial use where every result is exact or explicitly marked as approximate. Block 0 defines the product and the API contract in `docs/SPEC.md`. It also leaves minimal backend and frontend skeletons that build and pass one trivial test each, so that CodeRabbit can review the contract before any real code exists.

The interview settled decisions D1 to D20, which are already written in `docs/SPEC.md` but not committed yet. While planning, the owner added two requirements:

- Exact fractions must survive chaining: 1 ÷ 3 × 3 = 1.
- A non-zero result must never become 0.

Both change the number model. The decisions below replace parts of D9 to D16. The decisions log keeps the earlier entries, marks them as superseded and points to the new ones.

Branch: `chore/foundation`. Current state: one commit on `main` (`8c3f950 chore: initial project setup`). The working tree holds uncommitted `docs/SPEC.md` (D1 to D20) and `PROMPTS.md` (the restructured log).

## Decisions taken during planning (added to the SPEC decisions log in Milestone 1)

- **D21. Exact rationals.** Every result keeps its exact mathematical value, and the API is still stateless. A response is `200 {"value", "decimal", "kind"}`, which replaces the `{"result", "kind"}` shape of D16. `1 ÷ 3` returns `"value": "1/3"`. When the frontend chains with `"1/3"`, the result of `× 3` is exactly `"1"`. History entries also store `value`.
- **D22. Value syntax.** An operand or a `value` takes one of three forms:
  - A decimal: `-12.5` or `1e-3`, as in D12.
  - A reduced fraction `p/q`: integers, the sign on the numerator, `q > 1`. A `value` uses a fraction only when its decimal does not terminate (`1/4` is written `0.25`).
  - An approximate decimal with a `~` prefix: `~1.41421356237309504880`.

  If any operand is approximate, the result is approximate. Canonical forms follow D13: no trailing zeros, no `-0`, fractions reduced.
- **D23. Kinds.** `kind` is one of three values:
  - `exact`: a terminating decimal.
  - `repeating`: an exact rational whose decimal does not terminate.
  - `approximate`: an irrational result, or a result computed from an approximate operand.

  This replaces `irrational` in D10.
- **D24. The `decimal` field and the limits.**
  - `decimal` is display-only, and the frontend may abbreviate it.
    - Exact results include all their digits.
    - Repeating results show 100 fractional digits, rounded half even.
    - Approximate results show 20 fractional digits, rounded half even. The approximate `value` is `~` followed by this same decimal.
  - Scientific notation is used when 0 < |x| < 10⁻²⁰. The mantissa keeps the same number of fractional digits.
  - Limits:
    - |x| < 10⁴⁰, otherwise `result_too_large`.
    - The reduced denominator of an exact value is at most 10¹⁰⁰⁰. This is the same limit as 1000 fractional digits.
    - A non-zero approximate |x| is at least 10⁻¹⁰⁰⁰.
    - Breaking either of the last two limits returns `result_too_precise`.
  - A non-zero result never becomes 0. Example: 1 ÷ 3e30 returns `value "1/3000000000000000000000000000000"`, `decimal "3.3333…3333e-31"` (100 fractional digits) and `kind "repeating"`.

  Together these replace the precision rules of D9 and the limits of D12.
- **D25. Copy and display.** Copying puts `decimal` on the clipboard with all its digits, and chaining and history use `value`. This replaces the copy rule of D11. The display rules of D11 still apply to `decimal`, so repeating results can expand to 100 digits. Tiny values are shown in scientific notation.
- **D26. Frontend tooling.** ESLint with typescript-eslint and Prettier, with TypeScript pinned to `~6.0.3`. The current Vite React-TS template (create-vite 9.3.4) ships oxlint and TypeScript 7.0, and typescript-eslint 8.72 supports only TypeScript below 6.1.
- **Implementation choices, reported at the milestone checkpoints:**
  - The backend rejects unknown JSON fields (`400 unknown_field`).
  - A field-specific error includes a `field` extension member.
  - The Problem Details `type` is a URI that points to the error's entry in `docs/SPEC.md`.
  - The request body limit is 64 KiB (`413`).
- **Note for Block 1 (not decided here):** exact rational arithmetic is native to Go's `math/big.Rat`, and an integer square root avoids binary floating point. The library comparison in Block 1 should weigh this.

## Milestone 1: Specification

Steps:

1. **First commit on the branch:** this plan, verbatim, in `docs/plans/block-0-foundation.md`, as `docs(plans): add Block 0 plan`.
2. **Decisions log:** append D21 to D26 to the log in `docs/SPEC.md`. Add a "Superseded by" note to D9, D10, D11 (copy rule), D12 (limits), D13 (canonical forms) and D16 (response shape).
3. **`PROMPTS.md`:** add the planning round. It holds the tooling choice and the two Spanish prompts with their English translations: tiny values never become 0, and exact fractions survive chaining. It also holds the selected options for the approximation mark, copy and repeating digits, with status notes that point to D21 to D26.
4. **First commit after the plan:** steps 2 and 3 together, as `docs: record Block 0 decisions and prompt log`.
5. **`docs/SPEC.md`, product and number model.** Write these sections:
   - **Product:** the operations, and the UI behavior from D1, D2, D3 and D11.
   - **Numbers:** the operand grammar for decimals, scientific notation, `p/q` and `~`; the canonical forms; the limits.
   - **Exactness and precision:**
     - The rules for `kind` in each operation:
       - Division terminates only when the reduced denominator has no prime factors other than 2 and 5.
       - The square root of a rational is rational only when the reduced numerator and denominator are perfect squares.
       - Negative powers behave like a division.
       - Any approximate operand makes the result approximate.
     - The `decimal` rules and rounding half even.
     - An examples table:

       | Calculation | Result |
       |---|---|
       | `0.1 + 0.2` | `0.3` |
       | `1 ÷ 3` | `1/3` |
       | `1 ÷ 3 × 3` | `1` |
       | `2 ÷ 3` | decimal ends in `…667` |
       | `2^-2` | `0.25` |
       | `3^-1` | `1/3` |
       | `√2` | approximate |
       | `√2.25` | `1.5` |
       | `√(1/9)` | `1/3` |
       | `200 + 15%` | `230` |
       | `200 ÷ 15%` | `4000/3` |
       | `1 ÷ 3e30` | scientific notation |
       | `0^0` | `1` |

   Commit as `docs(spec): define the number model and precision rules`.
6. **`docs/SPEC.md`, API contract:**
   - Conventions: `POST /api/v1/{operation}`, `Content-Type: application/json`, strings only, unknown fields rejected, 64 KiB body limit.
   - One subsection per endpoint (`add`, `subtract`, `multiply`, `divide`, `power`, `sqrt`, `percentage`), with the request, a success response and the operation's specific errors.
   - The RFC 9457 error format (`type`, `title`, `status`, `detail`, `code`, optional `field`).
   - A table of every code with its HTTP status: `invalid_json`, `missing_field`, `unknown_field`, `invalid_number`, `operand_out_of_range`, `exponent_not_integer`, `exponent_out_of_range`, `invalid_operator`, `division_by_zero`, `negative_square_root`, `result_too_large`, `result_too_precise`, `not_found`, `method_not_allowed`, `payload_too_large`, `unsupported_media_type`.

   Commit as `docs(spec): define the API contract`.
7. **`README.md`:** add a short "Design decisions" section that links the SPEC. Commit as `docs: summarize design decisions in the README`.

Verification:

- Check every number in the examples table with a throwaway exact-arithmetic script (Node `BigInt` rationals, no floating point) in the scratchpad, then delete the script.
- Cross-check the SPEC against D1 to D26 and the out of scope list: every decision appears, and no contradictions remain.
- Re-read `PROMPTS.md`: the quotes are verbatim and the order is chronological.

## Milestone 2: Backend skeleton

Steps:

1. Run `go mod init github.com/msckell/precizzle/backend` in `backend/`.
2. Write `backend/main.go`:
   - `package main`.
   - `newHandler()` returns an empty `*http.ServeMux`.
   - `main` serves it on `:8080`.

   Package layout, routes, the health endpoint and configuration are Block 1 decisions.
3. Write `backend/main_test.go`: a table-driven test that uses `httptest` to check that an unknown path returns 404.
4. Update the `CLAUDE.md` Commands section with the backend commands (`go build`, `go test`, `go vet` and the gofmt check, run from `backend/`).

Commits: `build(backend): add Go module skeleton`, then `docs: add backend commands to CLAUDE.md`.

Verification: from `backend/`, `gofmt -l .` prints nothing, and `go vet ./...`, `go test ./...` and `go build ./...` all pass.

## Milestone 3: Frontend skeleton

Steps:

1. Run `npm create vite@9.3.4 frontend -- --template react-ts` without interactive prompts. First confirm the non-interactive flag with `--help`.
2. Trim the template:
   - Remove the demo assets, CSS, the frontend `README.md` and `frontend/.gitignore`, since the root `.gitignore` covers them.
   - `App` renders `<h1>Precizzle</h1>`, and the `index.html` title is "Precizzle".
3. Tooling (D26):
   - Pin `typescript@~6.0.3`.
   - Replace oxlint with an ESLint 10 flat config: `@eslint/js`, `typescript-eslint`, `eslint-plugin-react-hooks`, `globals` and `eslint-config-prettier`.
   - Add Prettier with its default settings.
   - Set `"strict": true` explicitly in `tsconfig.app.json` and `tsconfig.node.json`.
4. Tests:
   - Add Vitest 5.
   - `src/App.test.tsx` renders `App` with `react-dom/server`'s `renderToString` and checks that the output contains "Precizzle".
   - React Testing Library, jsdom and API mocking are left to Block 2.
5. Scripts: `dev`, `build` (`tsc -b && vite build`), `typecheck` (`tsc -b`), `lint` (`eslint .`), `format` and `format:check` (Prettier), `test` (`vitest run`), `preview`.
6. Update the `CLAUDE.md` Commands section and add a README "Development" section with the backend and frontend commands.

Commits: `build(frontend): scaffold Vite React TypeScript app`, `build(frontend): set up ESLint and Prettier`, `test(frontend): add Vitest smoke test`, then `docs: add frontend commands`.

Verification: from `frontend/`, `npm run lint`, `npm run format:check`, `npm run typecheck`, `npm test` and `npm run build` all pass. Also confirm that `git ls-files --eol` shows LF everywhere.

## Pull request (checkpoints 4 to 6)

- Title: `chore: Block 0 foundation: spec, API contract and skeletons`.
- Description: summary, motivation, how it was tested (with command output), decisions (D1 to D26) and what is out of scope.
- Then the CodeRabbit protocol. At close, `PROMPTS.md` gets the link to this plan.

## End-to-end verification for the block

- Go: `gofmt -l .` is empty, and `go vet`, `go test` and `go build` pass.
- Frontend: lint, format check, typecheck, test and build pass.
- `docs/SPEC.md` covers every decision and every edge case listed in the master prompt: division by zero, malformed or missing input, the square root of a negative number, huge powers, plus the cases added (tiny results, exact chaining, approximate propagation).
