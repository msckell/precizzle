# Precizzle

A full-stack calculator for financial and accounting use. Every result is exact or explicitly marked as approximate: 0.1 + 0.2 returns 0.3, and 1 ÷ 3 × 3 returns 1.

- Frontend: React with TypeScript.
- Backend: a REST microservice in Go, the single source of truth for results.

Operations: addition, subtraction, multiplication, division, exponentiation, square root and percentage.

## Status

Work in progress. Setup and run instructions and API examples are added as each block lands.

## Development

Requirements: Go 1.27 and Node.js 24 with npm.

Backend, from `backend/`:

```sh
go test ./...   # run the tests
go run .        # start the API on http://127.0.0.1:8080
```

Frontend, from `frontend/`:

```sh
npm ci              # install dependencies
npm test            # run the tests
npm run lint        # ESLint
npm run typecheck   # TypeScript
npm run build       # production build in dist/
npm run dev         # start the dev server
```

The API has no endpoints yet; Block 1 adds them.

### Known issues

- `npm audit` in `frontend/` reports a moderate advisory in `minimatch` ([GHSA-g3hv-75wg-m249](https://github.com/advisories/GHSA-g3hv-75wg-m249), a denial of service through crafted glob patterns, fixed in 11.0.1).
  - It reaches the project only through ESLint and typescript-eslint, which are development tools: it is not part of the built app, and the only glob patterns it processes are the project's own lint configuration.
  - ESLint 10.14 and typescript-eslint 8.72 require `minimatch` 10, and no patched 10.x release exists.
  - Forcing version 11 with npm `overrides` works in a local test, but it goes outside the versions ESLint supports, so it is not applied.
  - The advisory goes away when ESLint moves to `minimatch` 11.

## Design decisions

The full list, with the reason for each decision, is in the [decisions log](docs/SPEC.md#decisions-log). The main ones:

- **No binary floating point.** Numbers travel as JSON strings, and neither the backend nor the frontend ever stores an operand or a result in a floating-point number.
- **Exact rationals.** A result keeps its exact value: `1 ÷ 3` returns the value `1/3`, so chaining `× 3` returns exactly `1`. The API stays stateless, because the client sends that value back.
- **Every result states its kind.** A result is `exact` (a terminating decimal), `repeating` (an exact fraction) or `approximate` (an irrational root, or a result computed from an approximate value). Approximate values carry a `~` prefix, so the approximation follows them through every later calculation.
- **Display is separate from value.** Every response also has a `decimal` for display: 100 fractional digits for repeating results and 20 for approximate ones, rounded half even. The UI truncates it and expands it on demand, up to 100 digits. Tiny non-zero results switch to scientific notation and never become 0.
- **Pocket-calculator percentages.** 200 + 15% = 230, 200 − 15% = 170, 200 × 15% = 30, 200 ÷ 15% = 1333.33…
- **Standard errors.** Errors use RFC 9457 Problem Details with a stable machine-readable `code`.

## Project documents

- [docs/SPEC.md](docs/SPEC.md): product definition, number model, API contract, decisions log and out of scope list.
- [docs/plans/](docs/plans/): the approved plan of each block.
- [PROMPTS.md](PROMPTS.md): the prompts that drove the project.
