# Precizzle: master prompt

This is the master prompt for the whole project. It defines what we are building, how we work together and the plan in blocks. Every block gets its own interview, plan and execution, and I review the work at each checkpoint.

## 1. The product

Precizzle is a full-stack calculator built for financial and accounting use, so every result is an exact decimal.

- Frontend: React with TypeScript.
- Backend: a REST microservice in Go.
- The frontend sends every calculation to the backend API. The backend is the single source of truth for results.

Operations: addition, subtraction, multiplication, division, exponentiation, square root and percentage.

Frontend:

- An intuitive UI to enter input and read results.
- Input validation and clear error messages.
- A responsive layout that works well on mobile.

Backend:

- One endpoint per operation.
- Input validation and explicit handling of edge cases: division by zero, malformed or missing input, the square root of a negative number, powers that would produce huge results, and any other case we find.
- JSON responses with one consistent shape for results and another for errors.

Quality:

- Clean, readable and idiomatic code in both layers.
- Unit tests for the key functionality of both layers, with a coverage report.
- A README with setup instructions, how to run the frontend and the backend, API call examples, design decisions and assumptions.
- A Dockerfile that runs the frontend and the backend together.
- A prompt log (PROMPTS.md) with the prompts that drove the project.

When priorities conflict, the order is: correctness, clarity, maintainability, and only then extra features.

## 2. Core decision: exact decimal arithmetic

Binary floating point cannot represent most decimal fractions. That is why 0.1 + 0.2 returns 0.30000000000000004 in most languages, and that is unacceptable for money. In Precizzle, 0.1 + 0.2 returns 0.3.

- No binary floating point anywhere in the calculation path: not in the Go computation, not in JSON parsing and not in the frontend, where JavaScript numbers never hold operands or results.
- Compare shopspring/decimal, cockroachdb/apd and math/big. Ground the comparison in their documentation and in a small throwaway experiment (0.1 + 0.2, 1 / 3, 2 to the power of -2, the square root of 2). Recommend one, explain the trade-offs and let me decide.
- Some results cannot be exact (1 / 3, the square root of 2, some powers). For those we define explicit precision and rounding rules together, document them in docs/SPEC.md and in the README, and cover them with tests.

## 3. Me and my environment

- Windows with PowerShell. Node, npm, git and gh are installed, and gh is logged in as msckell.
- Go and Docker Desktop are installed on C:.
- Project folder: C:\Projects\precizzle. GitHub repository: msckell/precizzle (public).
- The CodeRabbit GitHub App is installed for all my repositories, so it reviews the pull requests of this one.
- You run every command: git, gh, go, npm and docker. I do not run console commands.
- I have never used Go and I want to learn it through this project. The first time a Go concept or idiom appears, explain it in the chat in a few lines. Those explanations never go into code comments.
- The first time a frontend or design term appears, explain it in plain words.
- Be concise and direct. No filler and no inflated language.
- Language: I may write to you in Spanish. Answer in Spanish in the chat. Everything that ends up in the repository or on GitHub is written in English: code, comments, commit messages, branch names, pull requests, replies to CodeRabbit, docs, plans and PROMPTS.md.

## 4. How we work

### Roles

I set the direction and make the product and design decisions. You do the mechanical work (commands, scaffolding, code, tests and git) and bring me options grounded in facts. Most of my time should go into answering strategic questions and reviewing results.

### Sessions and memory

Each block runs in a fresh session, so the context stays clean. The project memory lives in files: this prompt, CLAUDE.md, docs/SPEC.md, docs/plans/ and the git history. At the start of every session, read them before doing anything else.

### The block cycle

Block 0 starts with its own setup steps (section 9). After that, every block follows this cycle, and no application code is written before the block plan is approved:

1. Sync: switch to main, pull, confirm that the working tree is clean and create the block branch.
2. Recover context: read the files above and tell me in a few lines where the project stands.
3. Interview: ask me about the decisions listed for the block and anything else you find while exploring. Record every decision and its reason in the decisions log of docs/SPEC.md.
4. Plan: in plan mode, write the block plan split into 2 to 4 milestones. For each milestone, list the steps, the files, the tests and how you will verify it. We iterate it section by section until I approve it. If you cannot switch to plan mode yourself, ask me to do it.
5. Save the plan: the first commit on the branch is the approved plan, verbatim, in docs/plans/.
6. Execute milestone by milestone, with small commits.
7. Review: open the pull request and follow the CodeRabbit protocol (section 6).
8. Close: update PROMPTS.md, merge and sync main.

### How to ask me questions

- Use the AskUserQuestion tool, in rounds of up to 4 questions.
- Options must be concrete and mutually exclusive, each with its main trade-off in one line. Put your recommendation first and say why.
- If a question has no good discrete options, ask it in plain text.
- Skip obvious questions and anything already settled in docs/SPEC.md. Dig into the hard parts I may not have considered.
- Before offering options that depend on facts (library features, standards, API behavior), check those facts in the official documentation or with a quick experiment.
- If something is ambiguous and changes product behavior or the API, ask me. If it is an implementation detail with an idiomatic answer, decide it and mention it at the next checkpoint.

### Checkpoints

Stop and wait for my OK at each of these points:

1. End of the interview, with a summary of the decisions.
2. Plan approval.
3. End of each milestone.
4. Before pushing and opening the pull request, with its title and description.
5. After the CodeRabbit review, with the triage from section 6.
6. Before merging.

At every checkpoint, show evidence instead of claims: the files changed, the commands you ran with a short excerpt of their output (tests, linters, build, coverage), open questions and the next step.

Inside a milestone you work on your own and commit each small step once its checks pass.

### Actions that need my explicit OK

Creating the repository, pushing, opening pull requests, commenting on GitHub, merging and deleting branches. Never force push, never rewrite published history and never skip checks (for example with --no-verify). If something blocks you, stop and tell me.

## 5. Git and GitHub conventions

- One branch per block, created from an up-to-date main. Branch names are in section 9.
- main receives only the initial commit directly. Everything after it arrives through pull requests.
- Commit messages follow Conventional Commits (feat, fix, test, refactor, docs, build, ci, chore). Each commit holds one logical change and leaves the build and the tests passing.
- Pull request description: summary, motivation, how it was tested (with evidence), decisions taken and what is out of scope.
- Merge with a merge commit, so the step-by-step history and the boundaries of each block stay visible on main. Delete the branch after merging.

## 6. CodeRabbit review protocol

1. After opening the pull request, wait for the CodeRabbit review and fetch every comment with gh, nitpicks included. If no review shows up within a few minutes, tell me.
2. Verify each finding before reporting it: read the code involved and, when it applies, reproduce the problem with a test or a command. Never pass along an unverified comment.
3. Present the triage with one entry per comment:
   - CodeRabbit says: the diagnosis in one or two lines, with a link to the comment.
   - Verification: valid, partly valid or not valid, with the evidence.
   - Proposed fix: what you would change.
   - Needed: must fix, optional or skip, and why.
4. I decide every entry. Apply the accepted ones in separate commits.
5. Once I approve, reply on each thread with what was done or why it was declined, and resolve it.
6. Repeat until CodeRabbit approves and every check is green.

## 7. Code rules

- Scope: build what docs/SPEC.md asks for. No extra features, abstractions, configuration or files that nobody requested. The simplest solution that is correct and clear wins.
- Validation lives at the boundaries: HTTP input in the backend and user input in the frontend. Internal code trusts its inputs.
- Comments orient the reader, they do not narrate:
  - Go: a one-sentence doc comment on each exported identifier, starting with its name, as Go convention expects.
  - A short "why" comment where a decision is not obvious from the code (precision, rounding, a validation rule).
  - Nothing else: no comments that restate the code, no change history, no commented-out code and no leftover TODOs. The same applies to the frontend.
- Tests:
  - Write the arithmetic tests first, in the table-driven style that is idiomatic in Go.
  - Tests check real behavior and edge cases. Never hardcode values to make a test pass and never delete or weaken a test to get green. If a test looks wrong, tell me.
- Tooling: gofmt and go vet in Go. TypeScript in strict mode, ESLint and Prettier in the frontend.
- Never make claims about code you have not opened. Read it first.
- Delete temporary files and experiments once you are done with them.

## 8. Documentation and prompt log

- docs/prompts/00-master-prompt.md: this prompt, verbatim.
- docs/SPEC.md: the source of truth. Product definition, API contract (endpoints, request and response JSON, error format), precision and rounding rules, decisions log and out of scope list.
- docs/plans/: every approved block plan, verbatim, one file per block.
- PROMPTS.md: the chronological log of the prompts that drove the project. Its first entry links this master prompt in docs/prompts/. For each block it holds the starting prompt and any other key prompt verbatim, plus a link to the block plan. When a key prompt was written in Spanish, keep the original and add the English translation right below it.
- README.md: built incrementally. Each block adds its own section (setup, how to run, API examples, design decisions and assumptions) and Block 4 polishes it.
- CLAUDE.md: a short distillation of this prompt, under 80 lines: pointers to the files above, the checkpoints, the actions that need my OK, the language rule, the code and comment rules, and the commands to build, test and run. Put the code and comment rules in their own section, since that is the part CodeRabbit applies. Claude Code loads this file in every session and CodeRabbit uses it as review guidelines, so keep it short and accurate, and update the commands as they appear.

## 9. The plan in blocks

If you think the block structure should change, propose it at a checkpoint.

### Block 0: Foundation

Branch: chore/foundation

Goal: define the product and the API contract, and leave a base that builds, passes its tests and goes through review.

Steps:

1. Confirm you are working in C:\Projects\precizzle. Before anything else, save this prompt verbatim to docs/prompts/00-master-prompt.md.
2. Check the environment: versions of go, node, npm, git and docker, gh auth status, and that Docker Desktop is running. Report anything missing.
3. Prepare the initial commit for main: .gitignore, a .gitattributes that normalizes line endings to LF (I work on Windows and CI runs on Linux, so this keeps formatters and linters from failing on line endings), a README.md stub, CLAUDE.md, .coderabbit.yaml, PROMPTS.md (with this prompt as its first entry) and docs/prompts/00-master-prompt.md. Checkpoint: show me the files, then create the public repository with gh and push main.
4. Create the block branch and continue with the block cycle from step 3 (interview). The execution covers docs/SPEC.md, the folder structure and minimal backend and frontend skeletons that build and pass one trivial test each.
5. The pull request lets CodeRabbit review the API contract before any real code exists.

For .coderabbit.yaml, check every key against the official schema and docs: language en-US, the chill review profile, approval once all comments are resolved, path instructions for the backend, the frontend and the tests, and the docstring pre-merge check turned off (the comment policy lives in CLAUDE.md, which CodeRabbit already reads).

Decisions to cover:

- Interaction model: a pocket calculator keypad that chains operations, or a simple form (operand, operator, operand).
- How numbers travel in JSON: as strings or as JSON numbers, and what each option means for exactness in Go and in JavaScript.
- Precision and rounding for inexact results: decimal places or significant digits, rounding mode (half up or half even), and whether the response says that a result was rounded.
- Percentage semantics: "B percent of A", "A plus B percent" as many calculators do, or "A divided by 100".
- Exponentiation scope: integer exponents only, negative exponents, fractional exponents, and limits that prevent huge results.
- Accepted input: negative numbers, scientific notation, decimal separator, maximum number of digits, and how results are normalized (for example 2.50 or 2.5).
- Error format: a custom JSON shape or RFC 9457 Problem Details, machine-readable error codes and the HTTP status for each case.
- Endpoint shape: paths, versioning, HTTP method, and the request body for single-operand operations such as square root.
- How the API contract is written: Markdown in docs/SPEC.md or an OpenAPI file.
- Repository layout for the backend and the frontend.
- What is explicitly out of scope.

### Block 1: Backend in Go

Branch: feat/backend-api

Goal: the complete API with its tests, and a CI workflow that runs the Go checks on every pull request.

Decisions to cover:

- Decimal library (section 2).
- HTTP layer: the standard library net/http, which supports method and path routing since Go 1.22, or a router such as chi.
- Package layout (for example cmd/ and internal/).
- How the frontend reaches the API during development: CORS in the backend or a Vite dev proxy.
- Health endpoint, configuration through environment variables and request size limits.
- Coverage expectations for the arithmetic package and the HTTP handlers.

CI: a GitHub Actions workflow that runs a gofmt check, go vet and go test with coverage on every pull request and on main.

### Block 2: Frontend in React and TypeScript

Branch: feat/frontend

Goal: the UI design and its implementation with tests, and the frontend checks in CI.

Process: before writing components, propose 2 or 3 visual directions (described, or as quick static mockups) and let me choose. Aim for a look that feels intentional, not a generic template.

Decisions to cover:

- Visual direction and layout on desktop and mobile.
- Keyboard support and accessibility (labels for screen readers, visible focus).
- What the browser validates before calling the API, and how backend errors are shown.
- Keeping operands as strings end to end, so JavaScript numbers never touch them.
- Tooling: Vite, Vitest and React Testing Library, and how the API is mocked in tests.

CI: extend the workflow with lint, typecheck, tests and coverage for the frontend.

### Block 3: Docker

Branch: chore/docker

Goal: build and run the frontend and the backend together with documented Docker commands.

Decisions to cover:

- One image where the Go server also serves the built frontend, or two containers with docker compose.
- Base images, final image size and running as a non-root user.
- A CI step that builds the image.

### Block 4: Documentation and wrap-up

Branch: docs/readme

Goal: a README that lets a new reader run, test and understand the project in minutes, plus the final coverage report and PROMPTS.md.

Decisions to cover:

- README structure and API examples with curl.
- Where the coverage report lives: summarized in the README, generated by CI as an artifact, or both.
- A final review of docs/SPEC.md against what was built.

## 10. Definition of done

- Every operation works end to end, through the UI and through the API.
- 0.1 + 0.2 returns 0.3 in both.
- Every edge case in docs/SPEC.md has a test and a clear error message.
- Unit tests and coverage reports for both layers, and CI green on main.
- The documented Docker commands build and start the whole app.
- README.md, docs/SPEC.md and PROMPTS.md are complete and match the code.

## 11. Start

Start Block 0 now: save this prompt as described, check the environment and prepare the initial commit for my review.
