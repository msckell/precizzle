# Block 5 plan: Product images

## Context

Blocks 0 to 4 are merged, and the owner's last UI fixes are on main (`3bf0a30`). The README describes the product only in text. The owner wants images that show the whole product, not only calculations: three calculations, the "…" expansion of a result, and the history with "Load", "Use" and its limit of 100 entries.

Agreed in the chat before this plan:

- Every screen is a real capture of the running app. The "mockup" is only the frame around it: a browser window or a phone on a soft background in the brand colors (D72). No old design mockups and no invented numbers.
- Three calculations, titled only by what the image shows. The titles never rank them (no "simple" or "complex") and use no slogans.
- Two animated GIFs, made automatically: the "…" expansion, and the history (Load, then Use).
- The owner sees every image and GIF before anything is committed.

Branch: `docs/screenshots`, created from an up-to-date main. The first commit is this plan, verbatim, in `docs/plans/block-5-screenshots.md`.

## Decisions to record

A new "Block 5: Product images" table in `docs/SPEC.md`:

| # | Decision |
|---|---|
| D94 | The README shows real captures of the app (the Docker image) inside a browser or phone frame on a brand-colored background: five PNG images and two GIFs in `docs/images/`. The titles are captions in the README, not text inside the images, and each one only describes what the image shows. The capture and compose scripts are throwaway tools and are not committed; the plan file records how to remake the images. |

## The images

Desktop captures use a 1280×800 viewport (the desktop layout starts at 900 px), mobile captures 390×844. Stills are taken at a device scale factor of 2.

| File | Frame | What is typed (physical keyboard) | Expected result | README caption |
|---|---|---|---|---|
| `loan-payment.png` | Browser | `200000*0.005/` then `1-(1+0.005)^-360`, Enter | `1199.10105…`, repeating (mint) | Loan payment formula with a fraction and an exponent |
| `vat-discount.png` | Phone | `1250+21%-10%`, Enter | `1361.25`, exact | Price with 21% VAT and a 10% discount |
| `square-root.png` | Browser | `10*r2`, Enter | `≈ 14.14213…`, approximate (orange) | Square root with an approximate result |
| `history.png` | Browser | Opens the history after the calculations above | Panel with the entries and "Saved on this device, up to 100." | History of the last 100 calculations |
| `expand-decimals.gif` | Browser | `1/7`, Enter, then five clicks on "…" | `0.1428…` to 20, 40, 60, 80 and 100 decimals, ending on "Showing 100 of 100 decimals" | Result expanded from 4 to 100 decimals |
| `history-load-use.gif` | Browser | Open history → Load the loan formula → open history → Use its result → `*360`, Enter | The formula comes back editable, then the chip `1199.1010…` × 360 gives the total paid | Loading a saved calculation and using its result in a new one |
| `banner.png` | Browser + phone | The loan capture and the VAT capture side by side | — | Top of the README; also usable as the GitHub social preview (1280×640) |

The history entries come from the calculations themselves, typed in the same browser context, so nothing is seeded by hand.

## Milestones

### Milestone 1: Capture and compose (scratchpad only)

All tools live in the session scratchpad (`…/scratchpad/shots/`), never in the repository.

1. `docker build -t precizzle .` and `docker run --rm -d -p 127.0.0.1:8080:8080 --name precizzle precizzle`; wait for `(healthy)`.
2. Check each expected result first with `curl` against `POST /api/v1/evaluate`, using the expression the editor sends (D81).
3. `npm i playwright-core` in the scratchpad and drive the installed Chrome (`channel: "chrome"`), so no browser download is needed.
4. `capture.mjs`: one desktop context and one mobile context. It types every calculation with `page.keyboard` (about 120 ms per key), waits for the result card, and saves the raw PNGs. Selectors use the accessible names that already exist: "Show more decimals", "History, N calculations", "Close history", the "Load" and "Use" buttons.
5. GIF frames: during the two animated flows, a Chrome DevTools screencast saves PNG frames with their timestamps (sharper than Playwright's VP8 video). A small cursor dot injected only into the capture page shows where each click happens; the app is not changed.
6. `frame.html` + `compose.mjs`: draws the browser frame (three dots and an address bar) or the phone frame, places the capture in it on the brand background, and exports PNGs with Playwright. Brand colors from D39 and D72; Nunito from `frontend/node_modules/@fontsource-variable/nunito`.
7. GIFs: ffmpeg overlays the frames into the same browser frame, then `palettegen` (stats_mode=diff) + `paletteuse`, about 12 fps, 960 px wide, infinite loop, holding the final state for 2 seconds.
8. Size targets: each PNG under 600 KB (Pillow `optimize`), each GIF under 3 MB (lower fps or width if needed).

Verification: every image and GIF opened and checked by eye; every number on screen matches the `curl` output of step 2; file sizes listed.

**Checkpoint: the owner reviews all seven files before anything is copied into the repository.**

### Milestone 2: README and records

1. Copy the approved files to `docs/images/` (commit: `docs: add product images`).
2. `README.md`:
   - `banner.png` right below the first paragraph.
   - A new "A quick tour" section before "Quick start": the three calculations, the expansion GIF, the history image and the history GIF, each with its caption as bold text and as alt text.
3. `docs/SPEC.md`: the "Block 5: Product images" table with D94.
4. `PROMPTS.md`: a "7. Block 5: Product images" section with the key prompts of this block in Spanish with an English translation, a link to this plan, and the execution summary.

Verification: every relative link and image path in the README checked with a throwaway script; the README rendered on GitHub once the branch is pushed.

**Checkpoint before push and PR.**

### Milestone 3: Pull request and review

1. With the owner's OK: push `docs/screenshots` and open the PR (summary, motivation, evidence, D94, out of scope). No AI attribution.
2. CodeRabbit triage recorded in `PROMPTS.md`; **checkpoint after the triage**.
3. With the owner's OK: merge with a merge commit, delete the branch, sync main.
4. Optional, by the owner: upload `banner.png` in GitHub Settings → Social preview (there is no API for it).

## Out of scope

- Any change to the app, its tests or CI.
- Committing the capture tools or adding them to the build.
- Dark mode, other languages or other device frames.

## Cleanup

`docker stop precizzle`, and the scratchpad tools and raw frames are deleted once the images are approved.
