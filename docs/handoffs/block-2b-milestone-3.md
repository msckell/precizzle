# Block 2b frontend: Milestone 3 handoff

## Status and authorization

Paused at the owner's request to conserve usage on November 7, 2026. **Milestone 3 is incomplete and has not reached its review checkpoint.** Milestone 2 was approved. Continue the existing approved Milestone 3 directly when the owner resumes; no new interview or redesign is needed.

- Worktree: `C:\Projects\precizzle-frontend`; branch: `feat/frontend`.
- Primary instructions: root `AGENTS.md`, still byte-for-byte identical to `CLAUDE.md`.
- Read `docs/prompts/00-master-prompt.md`, `docs/SPEC.md`, the unchanged approved plan `docs/plans/block-2b-frontend.md`, and `PROMPTS.md` §4.5.
- `origin/main`, including backend merge `f753ad1`, was already merged without rebasing. Do not modify backend implementation.
- Preserve D71–D86 and the approved D81 percentage-denominator exception documented in SPEC. Preserve explicit user parentheses.
- No push, PR or merge is authorized. Stop for review at the end of Milestone 3.

## Implemented

- `frontend/src/styles.css`: approved color tokens, two-column desktop layout, mobile stacking below 900px, purple focus, reduced-motion caret, equation/keypad/result styling. Nunito Variable is bundled through `main.tsx`.
- `components/Equation.tsx`: recursive characters, operator circles, fraction bars and editable slots, roots, powers, exact-value chips, active box, caret, selected structures, click placement, whole-expression accessible label, horizontal scrolling and edge fades, mapped error highlighting and callout.
- `components/Keypad.tsx` and `useKeyboard.ts`: approved operator rows and keys; shared commands for mouse and physical keyboard; modifiers/composition ignored; Enter/Space remain with focused buttons. History-specific keyboard behavior is implemented but not yet exercised by a history UI.
- `components/ResultCard.tsx`: exact/repeating/approximate badges, ≈, progressive decimal expansion, scientific display, full-decimal copy with success/failure feedback, empty/pending state and dimmed edited results.
- `App.tsx` uses one `useReducer` with pure `state.ts`. It connects the Milestone 2 editor, API and history storage; aborts a superseded request; rejects stale completion IDs; announces results and errors. Editing preserves a dimmed result, while a new expression after a result clears the card. Successful results are persisted.
- Error integration tests cover division-by-zero highlighting, a missing operand at the expression end, network failure and malformed responses. The added serialized error fixtures were checked against the real API.

Recent implementation commits, in order: `3247a43` (tokens/font/layout), `5a18006` (equation), `67babde` (keypad/keyboard and state wiring, including the ResultCard component), `b6f1c28` (result styling and integration tests). The next commit records positioned-error integration tests. Read Git history for its hash.

## Remaining implementation

1. **HistoryPanel and Header wiring.** There is currently only an `<h1>Precizzle</h1>`, no Header component or History button. Implement the history count/button and `HistoryPanel` with miniature equations, Load, Use and inline Clear confirmation. It replaces the keypad column on desktop and fills the screen on mobile. Escape closes it and focus returns to History. `state.ts` already contains history/open/load/use/confirmation actions; reuse them. Ensure loading/using an entry cancels any pending evaluation and cannot accept a stale response. The existing Equation supports `compact` read-only rendering.
2. **Approved brand and decorations.** Add the ribbon P as inline SVG with accessible name P; mint bottom-left wave, coral/orange/amber top-right wave, faint peach bottom wave. Preserve the light mint equals key. Do not invent alternative branding or layouts.

The exact approved reference is outside the repository: `C:\Projects\precizzle-design\precizzle-equation-editor.html`. It contains the ribbon SVG paths, gradient definitions, decorations, desktop/mobile/history examples and styling. Reuse that reference; follow SPEC where older mockup annotations differ from finalized serialization. A frontend-design skill was consulted only for fidelity to the approved design, not to reopen design choices.

## Verification performed at pause

From `frontend/`, all passed:

- `npm run lint`
- `npm run format:check`
- `npm run typecheck`
- `npm run coverage`: **12 test files, 368 tests; 94.97% statements, 92.87% branches, 97.24% functions, 94.93% lines**. The 90% statement gate passes without changed thresholds or exclusions.
- `npm run build`
- `git diff --check`

All **70 fixtures** in `src/mocks/handlers.ts`, including request fixtures, matched the real backend's HTTP status, Content-Type and complete response body. Verification imported the fixture tables using Node's TypeScript support and posted them to `http://127.0.0.1:8080/api/v1/evaluate`. This is fixture/API verification, not a completed browser walkthrough.

The existing 347 Milestone 2 tests remain. App tests now cover keypad/keyboard equivalence, visible multiplication before roots, select-then-delete, ignored modified keys, button Enter/Space behavior, expansion through 100 decimals, exact chaining, new-expression versus editing behavior, approximate results, clipboard behavior and error cases. Equation component tests cover nested rendering, click positions, selected/error structures and compact read-only rendering.

## Verification still required

- RTL/user-event/MSW integration tests for History Load/Use/Clear, confirmation cancellation, ignored calculator keys while history is open, Escape/focus return and persistence across remounts.
- App-level tests for replacing/cancelling pending evaluations, stale results after edits/history actions, additional keyboard navigation and cursor placement after a result. The request mechanism exists but its UI-level cancellation/race behavior is not yet covered.
- Exercise scrolling, edge fades, error-callout alignment, resize handling, long numbers/expressions and scientific results. Current geometry is implemented, not visually accepted.
- Full test/coverage/lint/format/types/build pass after remaining changes. Current lower coverage in `state.ts` (66.66%) mainly includes not-yet-connected history actions and some navigation/stale-response branches; complete behavior tests rather than changing coverage exclusions.
- Browser walkthrough against the **real backend**, including keyboard-only calculation, error correction, history, copy, D81/D55 behavior, responsive layout and focus. No browser walkthrough or screenshots have been completed for Milestone 3.
- Capture the planned 1280×800 and 390×844 screenshots and check no page overflow at 360px. Chrome is installed at `C:\Program Files\Google\Chrome\Application\chrome.exe`; browser tooling is available, but no browser automation dependency was added. The plan excludes a new browser E2E test suite.

## Runtime and later scope

At pause the backend was listening on `127.0.0.1:8080` (PID 1408 at that moment); no frontend server was listening on 5173/5174. Recheck processes rather than trusting a saved PID. Run `npm run dev` from `frontend/` for normal verification through Vite's `/api` proxy. If necessary, run the existing backend with `go run .` without changing its files.

MSW remains for automated tests. `dev:mock`, the browser worker and its startup hook still exist, scheduled for removal before the frontend PR in Milestone 4. The owner's later instruction overrides the plan's temporary-mock manual verification: ordinary development now uses the real API.

CI additions, README/command documentation, removal of the temporary browser mock, and the PR workflow belong to Milestone 4 or later. Do not advance to them before the Milestone 3 checkpoint.
