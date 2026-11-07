# Block 2b plan: Frontend

## Context

Block 2 was split into two parallel parts. Part 2a builds `POST /api/v1/evaluate` in `backend/`. Part 2b, this plan, builds the React and TypeScript frontend in `frontend/` against the frozen contract in SPEC §5 and decisions D39 to D50.

The frontend is still the Block 0 skeleton:
- `App` renders `<h1>Precizzle</h1>`, with one smoke test.
- Vite 8, React 19.3, TypeScript `~6.0.3` (D26), ESLint 10 with typescript-eslint, Prettier, and Vitest 5.1.

Before the interview, the owner approved an updated Equation mockup. It is kept outside the repository, in `C:\Projects\precizzle-design\precizzle-equation-editor.html`. The interview then settled D71 to D86. They are written in the SPEC decisions log in Milestone 1:

| # | Decision |
|---|---|
| D71 | Editor visuals from the mockup:<br>• a purple box follows the cursor;<br>• operators sit in soft circles;<br>• the operator keys are two rows, `+ − × ÷` and `( ) xʸ √ %`;<br>• result chips are mint (exact, repeating) or orange with ≈ (approximate);<br>• empty slots are dashed;<br>• the history replaces the keypad column on desktop and is full screen on mobile, with an inline confirmation;<br>• "Showing N of 100 decimals" appears only once a result is expanded. |
| D72 | Brand and decorations:<br>• a ribbon "P" in the logo's inner colors, without purple;<br>• a mint wave at the bottom left, a coral to orange to amber wave at the top right and a faint peach wave at the bottom;<br>• the `=` key in light mint `#84EDCE`.<br>The reference logo stays a color reference only. |
| D73 | The editor is our own tree model with pure functions for inserting, deleting and moving the cursor. |
| D74 | ÷ (fraction) and xʸ take the operand right before the cursor as numerator or base. √ inserts an empty root, with the visible × of D46 when it applies. |
| D75 | Backspace or Delete next to a structure first selects all of it, and a second press deletes it. |
| D76 | Right after a result:<br>• an operator chains it (D1);<br>• a digit starts a new expression;<br>• clicking or moving the cursor edits the expression and dims the result card. |
| D77 | The physical keyboard works everywhere except in the open history. `r` inserts √, Enter or `=` evaluates, and Escape clears. |
| D78 | Accessibility:<br>• labelled keys and a visible purple focus ring;<br>• results and errors are announced in live regions;<br>• the expression is read as a whole, without cursor context. |
| D79 | Long expressions stay on one line and scroll sideways. The cursor is always kept visible, and an edge fade marks hidden content. |
| D80 | One `useReducer` in `App` holds the state, and a pure editor module holds the logic. The components are `Header`, `Equation`, `ResultCard`, `Keypad` and `HistoryPanel`. |
| D81 | Serialization always groups explicitly: a fraction is `((num)/(den))`, an exponent `^(exp)`, a root `sqrt(...)` and a chip `[value]`. A position map translates `position` back to an element. |
| D82 | A temporary MSW worker serves the SPEC §5 examples in `npm run dev:mock`. It is removed before the pull request. |
| D83 | Tests:<br>• table-driven unit tests for the editor, the serializer and the formatting;<br>• integration tests with React Testing Library, user-event and MSW;<br>• no browser end-to-end tests: the real API is checked by hand before the pull request. |
| D84 | Well-formed numbers (D47):<br>• `.` at the start of a number inserts `0.`;<br>• a second `.` in the same number is ignored;<br>• a trailing `.` is dropped when sending;<br>• `e` notation cannot be typed. |
| D85 | The history lives in `localStorage` under a versioned key. It stores the editor tree and the result, so "Load" restores structures. It is validated on read and holds only successful calculations. |
| D86 | Nunito is bundled with `@fontsource-variable/nunito`. There is no request to Google Fonts. |

Branch: `feat/frontend`, created from the shared base at `863e268`. The working tree is clean.

Facts checked:
- **MSW 3.0.4:** it exports `http` and `HttpResponse` from `msw/http`, `setupServer` from `msw/node` and `setupWorker` from `msw/browser`.
- **Vitest 5.1.0:** it accepts Vite 8 and pairs with `@vitest/coverage-v8` 5.1.0.
- **`@testing-library/jest-dom` 7.0.1:** it has a `/vitest` entry.
- **jsdom 31:** it needs Node `^24.15.0`.
- **`@fontsource-variable/nunito`:** version 5.3.0.
- **Leading zeros:** the backend accepts them (`007` is 7).

Open outside this block: PR #3 has to decide the `position` of `(2+`. That case gets its mock only after the decision.

## Milestone 1: Plan, decisions and tooling

Steps:

1. **First commit on the branch:** this plan, verbatim, in `docs/plans/block-2b-frontend.md`, as `docs(plans): add Block 2b plan`.
2. **`docs/SPEC.md`:**
   - D71 to D86 in the "Block 2b: Frontend" table.
   - New bullets in §1 "User interface" for the keyboard map (D77), D74 to D76, D79 and the visual rules of D71 and D72.
   - No change to §2 to §5 or D39 to D50.
3. **`PROMPTS.md` §4.5:**
   - The starting prompt, verbatim.
   - The mockup iterations in Spanish with their English translations: repeating results at 4 decimals, the decorations, the logo and the multicolor P, the three P variants, the choice of the ribbon and the softer `=`.
   - Three interview rounds as tables, each with its final decision.
   - The link to this plan.

   Commit steps 2 and 3 as `docs: record Block 2b decisions and prompt log`.
4. **Tooling.**
   - Add the dev dependencies `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom`, `jsdom`, `@vitest/coverage-v8` and `msw`, plus the dependency `@fontsource-variable/nunito`, at the versions above.
   - In `vite.config.ts`:
     - the `/api` proxy to `http://127.0.0.1:8080` (D31);
     - `test` options: `jsdom`, a setup file and v8 coverage of `src/**`.
   - Coverage counts application code only. It excludes `main.tsx` (D49), test files and the test-only mocks in `src/mocks/`, and fails below 90% of statements.
   - Scripts `coverage` (`vitest run --coverage`) and `dev:mock` (`vite --mode mock`).
   - `src/test/setup.ts` loads jest-dom and starts the MSW server, resets its handlers after each test and stops it at the end.

   Commit as `build(frontend): add the test stack, coverage and the API proxy`.
5. **`src/mocks/handlers.ts`.**
   - The `POST /api/v1/evaluate` handler answers from a fixture table, keyed by the exact `expression` string:
     - every SPEC §5 example (21 results and 11 errors with `position`);
     - the expressions the integration tests send in serialized form, such as `((1)/(7))` and `[1/7]*7`. Each of these was checked with exact rational arithmetic.
   - An unknown expression answers `500 internal_error`, so a missing fixture fails loudly.
   - The handler records every request (URL, method, `Content-Type` and body) for assertions.

   Commit as `test(frontend): mock the evaluate endpoint with MSW`.
6. **Temporary development mock (D82).**
   - `npx msw init public` adds `public/mockServiceWorker.js`.
   - `src/mocks/browser.ts` starts the worker with the same handlers.
   - `main.tsx` starts it only when `import.meta.env.MODE === "mock"`.

   Commit as `chore(frontend): add a temporary mocked API for development`.

Verification:
- `npm run lint`, `format:check`, `typecheck`, `test` and `build` pass.
- A test sends one SPEC success and one SPEC error through `fetch` and gets the fixture bodies back.
- `npm run dev:mock` starts, and the page's network tab shows the mocked response.

## Milestone 2: Editor core, formatting and API client

Everything here is plain TypeScript without React. Tests are written first and are table-driven with `it.each`.

### Editor model (`src/editor/`)

- **`model.ts`:**
  - A `Row` is a list of `Node`s.
  - A `Node` is one of:
    - `char`: a digit, `.`, `+`, `-`, `*`, `%`, `(` or `)`;
    - `chip`: `value`, `decimal` and `kind`;
    - `frac`: `num` and `den` rows;
    - `sqrt`: a `body` row;
    - `pow`: an `exp` row.
  - The base of a power is the node before it in the row.
  - A `Cursor` is a path of slots plus an index in the innermost row.
  - `EditorState` holds the root row, the cursor and an optional selected structure (D75).
- **`edit.ts`, pure functions from state to state:**
  - Inserting:
    - `insertDigit` and `insertDot` apply D84;
    - `insertOperator`;
    - `insertParen` applies D46;
    - `insertFraction` and `insertPower` take the preceding operand (D74). The operand is a number run, a chip, a structure or a parenthesized group, with its trailing `%` and powers.
    - `insertSqrt` applies D46;
    - `insertChip`.
  - Editing:
    - `toggleSign` adds or removes a unary `-` before the operand at the cursor;
    - `backspace` and `deleteForward` apply D75, and remove a structure whose slots are all empty.
  - Moving:
    - `moveLeft` and `moveRight` enter and leave slots;
    - `moveUp` and `moveDown` go between numerator and denominator, and into and out of an exponent;
    - `moveHome`, `moveEnd` and `placeCursor(path, index)`.
  - `clear`.
- **`serialize.ts`:**
  - `serialize(row)` returns the D81 text and a span list that maps each character range to its element:
    - the `/` of a fraction maps to its bar;
    - the `s` of `sqrt(` and the `^` map to their structure;
    - `[` maps to its chip;
    - the parentheses of an empty slot map to that slot.
  - A trailing `.` is dropped (D84). × and − are written as `*` and `-`.
  - `locate(spans, position)` returns the element to mark, or the end of the expression when `position` equals its length.
- **`speak.ts`:** the plain-words reading of an expression for screen readers (D78), for example "2 plus 3, squared, over 3, plus square root of 2.25".

### Display and services (`src/`)

- **`format.ts`:** the D11 and D25 display, using string operations only.
  - Truncation: 4 decimals for `repeating`, 5 for the rest, then 20, 40, 60, 80 and 100.
  - Whether more digits remain.
  - The mantissa and exponent of a scientific `decimal`.
  - The short text of a chip.
- **`api.ts`:**
  - `evaluate(expression, signal)` posts `{"expression"}` with `Content-Type: application/json`.
  - It checks the response shape at the boundary: string fields, a known `kind`, and a Problem Details body with a `code` and an optional integer `position`.
  - It returns a result, a problem, or a network failure.
- **`messages.ts`:** one English message per SPEC error `code` (D47), plus network failure and unexpected response. Each message says what happened and how to fix it.
- **`history.ts`:** `load` and `save` under `precizzle.history.v1` (D85).
  - At most 100 entries, newest first.
  - Each entry is validated on read, and invalid data reads as empty.
  - Storage errors are caught.

### Tests

- **Editor:**
  - Every insert rule, with D46 (`2` then `(`, `)` then √, `%` then `(`) and D84 (`.`, `0.`, a double dot).
  - What ÷ and xʸ take: a number, `2+3` takes only 3, a chip, a group, a structure and an empty row.
  - D75 select then delete.
  - Cursor moves through nested structures, and `placeCursor`.
  - `toggleSign` at the start, after an operator and after a `-`.
- **Serializer:**
  - Each structure and nesting, the mockup expressions, chips with fractions and `~` values, and a trailing dot.
  - `(1/2)^2` stays `((1)/(2))^(2)`.
  - `locate` for every §5 error position.
  - An empty slot, and the end of the expression.
- **Format:** each D11 boundary (4 and 5 decimals, ≤ 5 shown in full, every step, early stop), scientific `decimal`s, negative values, integers and approximate values.
- **API:** the request shape asserted through MSW, a success, a 400 with `position`, a 422, a 500, a malformed body and a network error.
- **History:** a round trip, the 101st entry, corrupt JSON, a wrong version and a storage that throws.

### Commits

1. `feat(editor): add the expression model and cursor movement`
2. `feat(editor): insert and delete at the cursor`
3. `feat(editor): serialize expressions with a position map`
4. `feat(frontend): format results for display`
5. `feat(frontend): call the evaluate endpoint`
6. `feat(frontend): keep the history in localStorage`

Verification: lint, typecheck and `npm run coverage` pass, with every module of this milestone near 100% statement coverage.

## Milestone 3: Interface

### Components (`src/components/`), following the mockup and D71/D72

- **`App`:**
  - The reducer joins the editor state, the phase (`editing`, `pending`, `result`, `error`), the shown result, the error, the history and whether it is open.
  - Only one request runs at a time: a new `=` aborts the previous one with `AbortController`.
- **`Header`:** the ribbon P as an inline SVG with `aria-label="P"`, the word and the History button with its count.
- **`Equation`:**
  - It renders the tree recursively: characters, operator circles, chips, fractions, roots, exponents, dashed empty slots, the active box and the caret.
  - Clicking a character places the cursor on the nearer side, and clicking a slot places it at the end of that slot.
  - It scrolls sideways to keep the caret visible, with edge fades (D79).
  - It marks the error element and shows the error callout next to it.
- **`ResultCard`:**
  - The kind badge, and the approximate mark (≈ plus the label).
  - The ellipsis expansion, the "Showing N of 100 decimals" caption once expanded, and Copy (`navigator.clipboard.writeText(decimal)`).
  - The dimmed state (D76) and the "Press = to calculate" empty state.
- **`Keypad`:**
  - The operator rows of D71, the digits, ±, `.`, C, backspace and `=`.
  - Every key has an `aria-label` and dispatches the same actions as the keyboard.
- **`HistoryPanel`:**
  - Mini equations and Load, Use and Clear history, with the inline confirmation.
  - It is the keypad column on desktop and full screen on mobile.
  - Escape closes it, and focus returns to the History button.
- **`useKeyboard`:**
  - One `keydown` listener on `window` with the D77 map.
  - It ignores keys with Ctrl, Meta or Alt, and every key while the history is open.
  - When a button has focus, Enter and Space stay with that button.

### Styles and accessibility

- `src/styles.css` holds the D39 tokens, Nunito Variable with tabular figures, the decorations as an absolutely positioned layer and the layout:
  - two columns from 900 px;
  - stacked below 900 px;
  - no horizontal page scroll at 360 px.
- Accessibility (D78):
  - A `role="status"` live region announces results, and `role="alert"` announces errors.
  - The equation has a label built with `speak.ts`.
  - Focus shows with `:focus-visible` in purple.
  - Under `prefers-reduced-motion`, the caret does not blink.

### Integration tests (RTL, user-event, MSW)

- `0.1 + 0.2 =` shows `0.3`, `Exact`.
- `1 ÷ 3 =` shows `0.3333…`. Expanding shows 20, 40, 60, 80 and 100 digits, and the ellipsis disappears at the end.
- Chaining sends `[1/3]*3` and shows `1` (D21).
- After a result, a digit starts a new expression and an arrow key dims the card (D76).
- An error marks the element at fault and shows its message. For example, `250 + 1 ÷ (3 − 3)` marks the fraction bar.
- Typing the same calculation with the keypad and with the keyboard sends the same request body.
- √ after a number inserts a visible ×.
- Backspace marks a structure and then deletes it.
- History:
  - Load restores the structures and the result.
  - Use inserts a chip at the cursor.
  - Clear asks for confirmation and empties both the list and the storage.
  - The panel ignores calculator keys while it is open.
- Copy writes the full `decimal`.
- An approximate result shows ≈ and its label.
- The live regions announce the result and the error.

### Commits

1. `feat(ui): add design tokens, the bundled font and the page layout`
2. `feat(ui): render the expression editor`
3. `feat(ui): add the keypad and keyboard input`
4. `feat(ui): show results with progressive expansion and copy`
5. `feat(ui): show errors next to the element at fault`
6. `feat(ui): add the calculation history`
7. `feat(ui): add the brand and background decorations`

Verification:
- `npm run coverage` passes at 90% or more.
- Check by hand with `npm run dev:mock`, and take screenshots with headless Chrome at 1280×800 and 390×844.
- Walk through a calculation with the keyboard only.

## Milestone 4: CI, documentation and integration

1. **`.github/workflows/ci.yml`:** add a `frontend` job after `backend`.
   - It runs with `working-directory: frontend` and `actions/setup-node` on Node 24.15 or newer, with the npm cache keyed on `frontend/package-lock.json`.
   - It runs `npm ci`, lint, `format:check`, typecheck, `npm run coverage` (where the threshold fails the job) and `build`.

   Commit as `ci: lint, typecheck, test and build the frontend`.
2. **`CLAUDE.md` frontend commands:**
   - coverage, `dev:mock` (until it is removed in step 5);
   - a note that `npm run dev` needs the backend on `127.0.0.1:8080`.

   Commit as `docs: update frontend commands in CLAUDE.md`.
3. **`README.md`:** a "Frontend" section right before "Backend API".
   - How to run it with the backend.
   - The keyboard map.
   - The tests and coverage.
   - The design notes: Equation, exact values in chips and the display rules.

   Commit as `docs: document the frontend in the README`.
4. **Integration:** after 2a is merged, merge `main` into `feat/frontend` (never a rebase).
   - Run the backend with `go run .` and the frontend with `npm run dev`.
   - Send every fixture expression to the real API with `curl.exe`, and confirm each response equals its fixture.
   - Repeat the Milestone 3 manual checks against the real API.
5. **Removal of the development mock:** remove `dev:mock`, `public/mockServiceWorker.js`, `src/mocks/browser.ts` and the hook in `main.tsx`, and update `CLAUDE.md`.

   Commit as `chore(frontend): remove the temporary mocked API`.

Verification:
- Run the job's commands locally.
- The fixture comparison against the real API shows no differences, and the checkpoint shows the output.
- The workflow runs on the pull request.

## Pull request (checkpoints 4 to 6)

- **Title:** `feat: Block 2b frontend with the Equation expression editor`.
- **Description:**
  - summary and motivation;
  - how it was tested: coverage, the comparison against the real API, and screenshots of desktop and mobile;
  - decisions D71 to D86;
  - out of scope.
- **Then:** the CodeRabbit protocol. `PROMPTS.md` §4.5 gets the execution and review prompts.

## Out of scope for this block

- Any change in `backend/`, SPEC §2 to §5 or D39 to D50.
- Docker (Block 3) and the coverage report in the README (Block 4).
- Browser end-to-end tests, internationalization, dark mode, cursor arrow keys on screen and keyboard shortcuts beyond D77.

## End-to-end verification for the block

- From `frontend/`, every check passes in CI: lint, format check, typecheck, coverage at 90% or more, and build.
- Against the real API:
  - `0.1 + 0.2 = 0.3`;
  - `1 ÷ 3`, then `× 3`, gives `1`;
  - `√2` is approximate with ≈;
  - `50 + 200 + 15%` gives 280;
  - a division by zero marks the fraction bar.
- The keypad and the keyboard produce identical requests.
- The layout works at 360 px and at 1280 px.
