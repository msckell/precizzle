# Precizzle specification

The source of truth for the product, the API contract, the precision and rounding rules, the decisions log and the out of scope list. Decision numbers such as (D7) point to the [decisions log](#decisions-log).

## 1. Product

Precizzle is a calculator for financial and accounting use. Every result is either exact or explicitly marked as approximate. The React frontend sends every calculation to the Go backend, which is the single source of truth for results.

### Operations

| Operation | Meaning |
|---|---|
| Addition | a + b |
| Subtraction | a − b |
| Multiplication | a × b |
| Division | a ÷ b |
| Exponentiation | base raised to an integer exponent between −1000 and 1000 (D8, D18) |
| Square root | √value |
| Percentage | percent ÷ 100, optionally applied to a base with the preceding operator (D7) |

Percentage rules, where B% means B ÷ 100 (D7):

| Input | Formula | Example |
|---|---|---|
| B% | B ÷ 100 | 15% = 0.15 |
| A + B% | A + A × B ÷ 100 | 200 + 15% = 230 |
| A − B% | A − A × B ÷ 100 | 200 − 15% = 170 |
| A × B% | A × B ÷ 100 | 200 × 15% = 30 |
| A ÷ B% | A ÷ (B ÷ 100) | 200 ÷ 15% = 4000/3 = 1333.33… |

### User interface

- A form-based calculator. Operators are clickable buttons: +, −, ×, ÷ (shown as a fraction a/b), xʸ, √ and %. Operands are entered with an on-screen keypad (digits, `.`, ±, backspace) or with the physical keyboard (D1, D2).
- After a result, pressing an operator reuses that result as the first operand (D1).
- A history icon in the display opens the last 100 calculations, saved in the browser. Clicking an entry loads its operands, operator and result. The 101st entry drops the oldest (D3).
- Chaining and history use the exact `value` of a result. Copying puts its `decimal` on the clipboard with all its digits (D21, D25).
- Results are shown truncated, never rounded, with at most 100 fractional digits (D11):
  - At first, `repeating` results show 4 decimals and an ellipsis. Every other result with more than 5 decimals shows 5 decimals and an ellipsis. Results with 5 or fewer decimals show in full.
  - Each click on the ellipsis expands the result to 20, 40, 60, 80 and then 100 fractional digits. Expansion stops early when the `decimal` has fewer digits, and the ellipsis disappears once every displayable digit is visible.
  - Expanding never changes the stored value. Tiny values are shown in scientific notation (D25). Long numbers never break the layout.
  - `approximate` results are visibly marked as approximate.
- The UI text is in English (D19).

## 2. Numbers

All numbers travel as JSON strings, never as JSON numbers (D4).

### Operand syntax

An operand is one of these forms, with no whitespace (D12, D22):

| Form | Grammar | Examples |
|---|---|---|
| Decimal | `[-]digits[.digits][(e\|E)[+\|-]digits]` | `12`, `-12.5`, `0.1`, `1e-3`, `2.5E10` |
| Fraction | `[-]digits/digits` | `1/3`, `-2/7`, `4/2` |
| Approximate | `~` followed by a decimal | `~1.4142135623730950488`, `~-3.2e-25` |

- Rejected: a leading `+`, a lone `.` (`.5`, `5.`), thousands separators, `,` as the decimal separator, a zero denominator (`1/0`), and JSON numbers. Every one of these is `invalid_number`.
- Fractions in requests do not need to be reduced. `4/2` is read as 2.
- An operand outside the limits below is `operand_out_of_range`.

### Canonical form

Every `value` and `decimal` in a response is canonical (D13, D22):

- No leading zeros, no trailing fractional zeros and no `-0`: `1.10 + 1.20` returns `2.3`.
- A rational whose decimal terminates is written as a decimal (`1/4` is `0.25`). A rational whose decimal does not terminate is written as a reduced fraction with the sign on the numerator (`-1/3`).
- An exact `value` never uses scientific notation. An approximate `value` is `~` followed by its `decimal`, which may use scientific notation.

### Limits

Operands and results share the same limits, so every result is a valid operand and chaining always works (D12, D24):

- \|x\| < 10⁴⁰. A result outside this limit is `result_too_large`.
- Written as a reduced fraction, an exact value has a denominator of at most 10¹⁰⁰⁰. For a terminating decimal, this means at most 1000 fractional digits. A result outside this limit is `result_too_precise`.
- An approximate value needs at most 1000 fractional digits when written in plain notation. A non-zero approximate result that would round to 0 at 1000 fractional digits is `result_too_precise`.

## 3. Exactness and precision

### Kinds

Every result has a `kind` (D23):

| Kind | Meaning | `value` | `decimal` |
|---|---|---|---|
| `exact` | A terminating decimal | The decimal, all digits | The same digits |
| `repeating` | An exact rational whose decimal does not terminate | The reduced fraction | Rounded to 100 fractional digits |
| `approximate` | An irrational result, or a result computed from an approximate operand | `~` and the `decimal` | Rounded to 20 fractional digits |

### How each operation decides the kind

- With exact operands, addition, subtraction, multiplication, division, power and percentage produce an exact rational. That rational is `exact` when its reduced denominator has no prime factors other than 2 and 5, and `repeating` otherwise.
- A negative power is the reciprocal of the positive power: 3^-1 = 1/3.
- The square root of an exact rational is rational only when its reduced numerator and denominator are both perfect squares: √2.25 = 1.5 and √(1/9) = 1/3. Every other square root is irrational, and therefore `approximate`.
- If any operand is approximate, the operation works on the operand's written digits as an exact rational, and the result is `approximate`.
- The exponent of a power must be an exact integer, such as `2`, `2.0`, `4/2` or `1e2`. An approximate exponent is `exponent_not_integer`, even if its digits look like an integer.

### The `decimal` field and rounding

- `decimal` is for display, and the frontend may abbreviate it (D24).
- `exact` results keep every digit. `repeating` results are rounded to at most 100 fractional digits, and `approximate` results to at most 20, both with round half even. The canonical form then removes trailing zeros, so √2 rounded to 20 digits, `1.41421356237309504880`, becomes `1.4142135623730950488`.
- Square roots are correctly rounded: the 20 digits are those of the true root, not of an intermediate approximation.
- Round half even: a tie goes to the even digit, so 0.125 rounded to 2 digits is 0.12, and 0.135 is 0.14.
- When 0 < \|x\| < 10⁻²⁰, `decimal` uses scientific notation: one non-zero digit before the point, the same number of fractional digits in the mantissa, then `e` and the exponent (`3.3333…3333e-31`). An approximate result never goes past the 1000th fractional digit of its plain form, so its mantissa can have fewer than 20 fractional digits.
- A non-zero result never becomes 0. It is either represented or rejected with `result_too_precise` (D24).

### Examples

Long decimals are abbreviated here with `…`. The note in brackets gives the number of fractional digits in the real response.

| Calculation | `value` | `decimal` | `kind` |
|---|---|---|---|
| 0.1 + 0.2 | `0.3` | `0.3` | `exact` |
| 1.10 + 1.20 | `2.3` | `2.3` | `exact` |
| 1 ÷ 3 | `1/3` | `0.3333…3333` (100) | `repeating` |
| `1/3` × 3 | `1` | `1` | `exact` |
| 2 ÷ 3 | `2/3` | `0.6666…6667` (100) | `repeating` |
| 1 ÷ 7 | `1/7` | `0.1428…1429` (100) | `repeating` |
| 2 ^ −2 | `0.25` | `0.25` | `exact` |
| 3 ^ −1 | `1/3` | `0.3333…3333` (100) | `repeating` |
| 0 ^ 0 | `1` | `1` | `exact` |
| √2 | `~1.4142135623730950488` | `1.4142135623730950488` | `approximate` |
| √2.25 | `1.5` | `1.5` | `exact` |
| √(1/9) | `1/3` | `0.3333…3333` (100) | `repeating` |
| `~1.4142135623730950488` × `~1.4142135623730950488` | `~2` | `2` | `approximate` |
| 15% | `0.15` | `0.15` | `exact` |
| 200 + 15% | `230` | `230` | `exact` |
| 200 − 15% | `170` | `170` | `exact` |
| 200 × 15% | `30` | `30` | `exact` |
| 200 ÷ 15% | `4000/3` | `1333.3333…3333` (100) | `repeating` |
| 1 ÷ 3e30 | `1/3000000000000000000000000000000` | `3.3333…3333e-31` (100 in the mantissa) | `repeating` |
| 1e-25 + 0 | `0.0000000000000000000000001` | `1e-25` | `exact` |

Every row was checked with exact rational arithmetic. Errors:

| Calculation | Error |
|---|---|
| 1 ÷ 0, 0 ^ −1, 200 ÷ 0% | `division_by_zero` |
| √−4 | `negative_square_root` |
| 10 ^ 40 | `result_too_large` |
| 1e-1000 × 1e-1000 | `result_too_precise` |
| 2 ^ 1001 | `exponent_out_of_range` |
| 2 ^ 0.5, 2 ^ `~2` | `exponent_not_integer` |

### Known consequences

- Approximation spreads. √2 × √2 returns `~2` and stays `approximate`, because it was computed from rounded digits.
- Exact chaining depends on the frontend sending back `value`, not `decimal`.

## 4. API contract

### Conventions

- Every operation is `POST /api/v1/{operation}` (D14).
- The request has `Content-Type: application/json` and a body of at most 64 KiB. The body is a JSON object whose fields are strings (D4). Unknown fields are rejected, and `null` counts as a missing field.
- A success is `200` with `Content-Type: application/json` and the body `{"value": "...", "decimal": "...", "kind": "..."}` (D21).
- An error uses RFC 9457 Problem Details with `Content-Type: application/problem+json` (D5). See [Errors](#errors).
- The API is stateless and has no authentication.

### Endpoints

| Endpoint | Request body | Computes | Specific errors |
|---|---|---|---|
| `POST /api/v1/add` | `{"a", "b"}` | a + b | — |
| `POST /api/v1/subtract` | `{"a", "b"}` | a − b | — |
| `POST /api/v1/multiply` | `{"a", "b"}` | a × b | — |
| `POST /api/v1/divide` | `{"a", "b"}` | a ÷ b | `division_by_zero` |
| `POST /api/v1/power` | `{"base", "exponent"}` | base ^ exponent | `exponent_not_integer`, `exponent_out_of_range`, `division_by_zero` (0 raised to a negative exponent) |
| `POST /api/v1/sqrt` | `{"value"}` | √value | `negative_square_root` |
| `POST /api/v1/percentage` | `{"percent", "base"?, "operator"?}` | the rules in [Operations](#operations) | `missing_field` (`base` without `operator` or the reverse), `invalid_operator`, `division_by_zero` (`divide` with a percent of 0) |

- In `percentage`, `operator` is one of `add`, `subtract`, `multiply` or `divide`.
- Every endpoint can also return `invalid_json`, `missing_field`, `unknown_field`, `invalid_number`, `operand_out_of_range`, `result_too_large` and `result_too_precise`.

### Examples

In the examples below, a `decimal` shortened with `…` has 100 fractional digits in the real response.

Addition:

```http
POST /api/v1/add
Content-Type: application/json

{"a": "0.1", "b": "0.2"}
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"value": "0.3", "decimal": "0.3", "kind": "exact"}
```

Division, then chaining its `value`:

```text
POST /api/v1/divide    {"a": "1", "b": "3"}
200                    {"value": "1/3", "decimal": "0.3333…3333", "kind": "repeating"}

POST /api/v1/multiply  {"a": "1/3", "b": "3"}
200                    {"value": "1", "decimal": "1", "kind": "exact"}

POST /api/v1/subtract  {"a": "1", "b": "1/3"}
200                    {"value": "2/3", "decimal": "0.6666…6667", "kind": "repeating"}
```

Power, square root and percentage:

```text
POST /api/v1/power       {"base": "2", "exponent": "-2"}
200                      {"value": "0.25", "decimal": "0.25", "kind": "exact"}

POST /api/v1/sqrt        {"value": "2"}
200                      {"value": "~1.4142135623730950488", "decimal": "1.4142135623730950488", "kind": "approximate"}

POST /api/v1/percentage  {"percent": "15", "base": "200", "operator": "add"}
200                      {"value": "230", "decimal": "230", "kind": "exact"}

POST /api/v1/percentage  {"percent": "15"}
200                      {"value": "0.15", "decimal": "0.15", "kind": "exact"}
```

A tiny result:

```text
POST /api/v1/divide  {"a": "1", "b": "3e30"}
200                  {"value": "1/3000000000000000000000000000000", "decimal": "3.3333…3333e-31", "kind": "repeating"}
```

### Errors

Every error is a Problem Details object (RFC 9457) with these members:

| Member | Content |
|---|---|
| `type` | A URI that identifies the problem type and points to its row in the table below, e.g. `https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#division_by_zero` |
| `title` | A short, fixed summary of the problem type |
| `status` | The HTTP status code |
| `detail` | A human-readable explanation of this occurrence |
| `code` | A stable, machine-readable code (extension member) |
| `field` | The request field at fault, when there is one (extension member) |

Example:

```http
HTTP/1.1 422 Unprocessable Content
Content-Type: application/problem+json

{
  "type": "https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#division_by_zero",
  "title": "Division by zero",
  "status": 422,
  "detail": "The divisor b is zero.",
  "code": "division_by_zero",
  "field": "b"
}
```

`400` means the request is wrong and the client must fix it. `422` means the request is well formed, but the result does not exist or cannot be represented (D17).

| Code | Status | Title | When |
|---|---|---|---|
| <a id="invalid_json"></a>`invalid_json` | 400 | Invalid JSON | The body is not a single valid JSON object. |
| <a id="missing_field"></a>`missing_field` | 400 | Missing field | A required field is absent or `null`, or `percentage` has `base` without `operator` or the reverse. |
| <a id="unknown_field"></a>`unknown_field` | 400 | Unknown field | The body has a field that the endpoint does not define. |
| <a id="invalid_number"></a>`invalid_number` | 400 | Invalid number | A number field is not a JSON string, or the string does not match the [operand syntax](#operand-syntax). |
| <a id="operand_out_of_range"></a>`operand_out_of_range` | 400 | Operand out of range | An operand is outside the [limits](#limits). |
| <a id="exponent_not_integer"></a>`exponent_not_integer` | 400 | Exponent not integer | The exponent is not an exact integer. |
| <a id="exponent_out_of_range"></a>`exponent_out_of_range` | 400 | Exponent out of range | The exponent is below −1000 or above 1000. |
| <a id="invalid_operator"></a>`invalid_operator` | 400 | Invalid operator | `operator` is not `add`, `subtract`, `multiply` or `divide`. |
| <a id="division_by_zero"></a>`division_by_zero` | 422 | Division by zero | The divisor is zero: `b` in `divide`, a percent of 0 with `divide`, or 0 raised to a negative exponent. |
| <a id="negative_square_root"></a>`negative_square_root` | 422 | Negative square root | `sqrt` received a negative value. |
| <a id="result_too_large"></a>`result_too_large` | 422 | Result too large | \|result\| ≥ 10⁴⁰. |
| <a id="result_too_precise"></a>`result_too_precise` | 422 | Result too precise | The exact result needs a denominator above 10¹⁰⁰⁰, or a non-zero approximate result would round to 0 at 1000 fractional digits. |
| <a id="not_found"></a>`not_found` | 404 | Not found | No endpoint exists at the path under `/api/`. |
| <a id="method_not_allowed"></a>`method_not_allowed` | 405 | Method not allowed | The path exists, but the method is not `POST`. The response includes `Allow: POST`. |
| <a id="payload_too_large"></a>`payload_too_large` | 413 | Payload too large | The body is larger than 64 KiB. |
| <a id="unsupported_media_type"></a>`unsupported_media_type` | 415 | Unsupported media type | `Content-Type` is not `application/json`. |

When a request has several problems, the API reports only the first one, checked in this order: path and method, media type, body size, JSON syntax, unknown fields, missing fields, operand syntax, operand limits, exponent rules, and then the calculation itself.

## Decisions log

### Block 0: Foundation

| # | Decision | Reason |
|---|---|---|
| D1 | Form-based calculator with chaining. Operators (+, −, ×, ÷, xʸ, √, %) are clickable buttons. Operands are entered through an on-screen keypad (digits, `.`, ±, backspace) or the physical keyboard. After a result, pressing an operator reuses that result as the first operand. | Reusing results and entering calculations comfortably is the top UI priority, and users must not need to know keyboard shortcuts. One API call per step keeps the backend the source of truth. Some mobile number keyboards lack a minus key, which the ± button covers. |
| D2 | Division is presented in the UI as a fraction (a/b). | The owner asked for a clickable fraction element. It needs no API change. |
| D3 | A history icon in the display opens the last 100 calculations, saved in the browser (localStorage). Clicking an entry loads its operands, operator and result. The 101st entry drops the oldest. Frontend only. | The owner wants to go back to earlier calculations and results, and the history must survive a reload. |
| D4 | Numbers travel as JSON strings in requests and responses. JSON numbers are rejected as invalid input. | An experiment showed that JavaScript `JSON.parse` turns `12345678901234567890.123` into `12345678901234567000`, and Go's default decoder turns it into a `float64`. Strings are exact in both languages without special parsing. |
| D5 | Errors use RFC 9457 Problem Details (`application/problem+json`) with a machine-readable `code` extension member. | It is a standard, self-describing format, and `code` gives clients a stable value to branch on. |
| D6 | The API contract is written in Markdown in this file. | One source of truth, readable on GitHub and by CodeRabbit, with no extra tooling. Nothing in scope consumes an OpenAPI file. |
| D7 | Percentage divides its number by 100 and follows the preceding operator: A + B% = A + A·B/100, A − B% = A − A·B/100, A × B% = A·B/100, A ÷ B% = A ÷ (B/100), and B% alone = B/100. | These are the pocket calculator semantics the owner described: 200 + 15% = 230, 200 − 15% = 170, 200 × 15% = 30, 200 ÷ 15% = 1333.33…, 15% = 0.15. |
| D8 | Exponentiation accepts integer exponents, negative ones included. | Positive exponents stay exact, and integer exponents never produce complex-number cases. |
| D9 | **Superseded by D21 and D24.** Exact results are returned exactly, with up to 1000 fractional digits. Results that need approximation (repeating decimals, irrational roots) are rounded to 20 decimal places with round half even. An exact result that exceeds the supported limits is an error and is never silently rounded. | This keeps calculation precision separate from display precision. Half even avoids the upward bias of half up across many operations. |
| D10 | **Superseded by D23.** Every successful response states the kind of result: `exact`, `repeating` or `irrational`. | The display rule (D11) depends on it, and the backend can prove it. A reduced fraction repeats unless its denominator has no prime factors other than 2 and 5, and the square root of a rational number is irrational unless the numerator and the denominator are perfect squares. |
| D11 | **The stored value and the copy rule are superseded by D21 and D25.** Stored precision and display precision are separate. The stored value is the complete result the API returned (D9): up to 1000 fractional digits for exact results and 20 for approximated ones. The display is frontend only and shows at most 100 fractional digits, truncated and never rounded. At first it shows 4 decimals and an ellipsis for `repeating` results, and 5 decimals and an ellipsis for every other result with more than 5 decimals. Results with 5 or fewer decimals show in full with no ellipsis. Each click on the ellipsis expands the result to 20, 40, 60, 80 and then 100 fractional digits. Expansion stops early when the result has fewer digits, and stops at 100 even if more are stored. The ellipsis disappears once every displayable digit is visible. Expanding never recalculates or changes the stored value. Copying and chaining always use the complete stored value, whatever is visible. Long numbers never break the layout. | The owner prioritizes both comfortable reading and mathematical precision: the display stays short and readable, and no stored digit is ever lost. |
| D12 | **The limits are superseded by D24, and D22 adds formats.** Operands are plain decimals with an optional leading `-` and an optional `.`, and scientific notation (`1e-3`, `2.5E10`) is also accepted. A leading `+`, thousands separators and `,` as the decimal separator are rejected. Operands and results share the same limits, 40 integer digits and 1000 fractional digits, measured after expanding scientific notation. | One unambiguous syntax plus scientific notation for API users. Shared limits guarantee that every result is a valid operand, so chaining always works. |
| D13 | **Extended by D22.** Results are canonical: no trailing zeros, no leading zeros, no `-0` and no scientific notation (`1.10 + 1.20` returns `2.3`). | Each value has one representation, which keeps the API and the tests simple. The UI can still format results. |
| D14 | One endpoint per operation: `POST /api/v1/{operation}` with a JSON body. Percentage takes the percent plus an optional base and preceding operator. | Versioned paths and JSON bodies of strings follow from D4. Percentage needs the operator for D7, and keeping it in one call puts every percent rule in Go, where it is tested. |
| D15 | Repository layout: the Go module in `backend/`, the Vite app in `frontend/`, and documentation at the root. | Each layer is self-contained, and the layout matches the path instructions in `.coderabbit.yaml`. |
| D16 | **The response shape is superseded by D21.** Request bodies: `{"a","b"}` for add, subtract, multiply and divide; `{"base","exponent"}` for power; `{"value"}` for square root; `{"percent"}` for percentage, with optional `"base"` and `"operator"` (`add`, `subtract`, `multiply` or `divide`), which must be given together. A successful response is `200 {"result": "...", "kind": "..."}`. | Field names state each operand's role, and the success shape is the same for every operation. |
| D17 | **D24 restates the limits behind `result_too_large` and `result_too_precise`.** HTTP status codes: `400` when the request itself is wrong (malformed JSON, missing field, invalid or out-of-range number, non-integer or out-of-range exponent, invalid operator); `422` when the request is well formed but the result does not exist or cannot be represented (division by zero, square root of a negative number, `result_too_large` for more than 40 integer digits, `result_too_precise` for more than 1000 fractional digits); `404`, `405`, `413` for an oversized body, and `415` for a body that is not JSON. | Clients can tell a request they must fix from a calculation that has no representable answer. |
| D18 | Power edge cases: `0^0 = 1`, `0` raised to a negative exponent is a division by zero, and the exponent must be between −1000 and 1000. | `0^0 = 1` is the usual convention in calculators and programming languages. The exponent range bounds the cost of a single request. Most larger exponents would exceed the result limits anyway, but a few representable results are rejected too, such as `1.0001^-2000` (about 0.82). |
| D19 | The UI text is in English. | Everything in the repository is in English. |
| D20 | The out of scope list below is approved. | It keeps the calculator focused on the operations and behaviors in this document. |
| D21 | Every result keeps its exact mathematical value, and the API stays stateless. A successful response is `200 {"value": "...", "decimal": "...", "kind": "..."}`. `value` is used for chaining and stored in the history. `1 ÷ 3` returns `"value": "1/3"`, and chaining `× 3` on it returns exactly `"1"`. | The owner requires exact fractions to survive chaining and history reuse: 1 ÷ 3 × 3 must be 1, not 0.99999999999999999999. |
| D22 | An operand or a `value` takes one of three forms: a decimal (`-12.5`, `1e-3`), a reduced fraction `p/q` (integers, sign on the numerator, `q > 1`), or an approximate decimal with a `~` prefix (`~1.4142135623730950488`). A `value` is a fraction only when its decimal does not terminate (`1/4` is written `0.25`). If any operand is approximate, the result is approximate. | Each value string describes itself, so the frontend can send back exactly what it received, request bodies stay flat, and the API needs no state to know whether a value is exact. |
| D23 | `kind` is `exact` (a terminating decimal), `repeating` (an exact rational whose decimal does not terminate) or `approximate` (an irrational result, or one computed from an approximate operand). | With exact fractions, only approximation needs a warning. Approximation also spreads to every later result, not just to irrational ones. |
| D24 | `decimal` is for display only. Exact results include every digit. Repeating results are rounded to at most 100 fractional digits and approximate results to at most 20, both with round half even and without trailing zeros (D13). The `value` of an approximate result is `~` followed by its `decimal`. When 0 < \|x\| < 10⁻²⁰, `decimal` uses scientific notation, and the mantissa keeps the same number of fractional digits. Limits: \|x\| < 10⁴⁰ (otherwise `result_too_large`); the reduced denominator of an exact value is at most 10¹⁰⁰⁰, the same as 1000 fractional digits; and an approximate value needs at most 1000 fractional digits in plain notation, so a non-zero approximate result must not round to 0 at that precision. Breaking either of the last two limits returns `result_too_precise`. A non-zero result never becomes 0. | The owner requires that tiny non-zero results never become 0 and keep their magnitude, for example 1 ÷ 3e30 ≈ 3.3333… × 10⁻³¹. The limits are equivalent to the earlier 40 integer and 1000 fractional digits, so chaining always works. |
| D25 | Copying puts `decimal` on the clipboard with all its digits. Chaining and history use `value`. The display rules of D11 apply to `decimal`, so repeating results can expand to 100 digits, and tiny values are shown in scientific notation. | A decimal pastes correctly into other apps, while a fraction such as `1/3` can be misread (a spreadsheet turns it into a date). |
| D26 | Frontend tooling: ESLint with typescript-eslint, and Prettier, with TypeScript pinned to `~6.0.3`. | The master prompt asks for ESLint and Prettier. The current Vite React-TS template (create-vite 9.3.4) ships oxlint and TypeScript 7.0, but typescript-eslint 8.72 supports only TypeScript below 6.1. |
| D27 | During local development the backend listens only on `127.0.0.1:8080`. The Docker setup (Block 3) sets the address its environment needs. | The owner does not want the development server reachable from other machines on the network. |

## Out of scope

- User accounts and authentication.
- Server-side storage of the history (the history lives in the browser, D3).
- Expressions with several operators and operator precedence in one request.
- Fractional exponents, roots other than the square root, and complex numbers.
- Scientific functions such as trigonometry and logarithms.
- Memory keys (M+, M−, MR).
- Thousands separators, `,` as the decimal separator, currency formatting and internationalization.
