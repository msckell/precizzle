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

Expressions combine these operations with parentheses and the usual precedence (D41, D42). Inside an expression, a percentage after + or − applies to the term right before it, so 50 + 200 + 15% = 50 + 230 = 280 (D43). The rules are in [Expressions](#5-expressions).

### User interface

- The visual direction is "Equation": a white page, the calculation as an editable equation, and the result in its own mint card (D39).
- The calculation is an expression in an editor with a cursor, a thin vertical bar (D40). The user places the cursor by clicking or tapping between characters, or moves it with the arrow keys, and inserts or deletes digits, operators and parentheses anywhere. On-screen buttons and the physical keyboard behave the same way.
- Buttons: digits, `.`, ± and backspace; the operators +, −, ×, ÷ (drawn as a fraction a/b), xʸ, √ and %; and two parenthesis keys, `(` and `)` (D1, D2, D46). Each one inserts at the cursor. Fractions, square roots and exponents are drawn as structures whose parts can be edited (D40). Pressing √ or `(` right after a number, a `)` or a `%` first inserts a visible × (D46).
- Every calculation is computed by `POST /api/v1/evaluate` (D41).
- After a result, pressing an operator starts a new expression with that result as its first operand (D1).
- A history icon opens the last 100 calculations, saved in the browser. The 101st entry drops the oldest (D3). "Load" restores an entry's expression and result, "Use" inserts its result at the cursor, and "Clear history" deletes every entry after a confirmation (D48).
- Chaining, "Use" and the history use the exact `value` of a result. Copying puts its `decimal` on the clipboard with all its digits (D21, D25).
- An error is shown next to the part of the expression at fault, with a message written for each error `code` (D47).
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
- An exact value whose decimal terminates has at most 1000 fractional digits. An exact value whose decimal does not terminate has a reduced denominator of at most 10¹⁰⁰⁰. A result outside these limits is `result_too_precise`. The two rules differ on purpose: `1/2^1001` has a denominator below 10¹⁰⁰⁰, but its decimal needs 1001 fractional digits, so it is rejected.
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
- The notation depends on the value the API returns, after rounding, not on the value before rounding (D36). An approximate result just below 10⁻²⁰ that rounds to 10⁻²⁰ is written `0.00000000000000000001`, the same as when that value is sent back.
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
| 1e-1000 × 1e-1000, 0.25 ^ 501 | `result_too_precise` |
| 2 ^ 1001 | `exponent_out_of_range` |
| 2 ^ 0.5, 2 ^ `~2` | `exponent_not_integer` |

### Known consequences

- Approximation spreads. √2 × √2 returns `~2` and stays `approximate`, because it was computed from rounded digits.
- Exact chaining depends on the frontend sending back `value`, not `decimal`.

## 4. API contract

### Conventions

- Every operation is `POST /api/v1/{operation}` (D14). Expressions use `POST /api/v1/evaluate` (D41), described in [Expressions](#5-expressions).
- The request has `Content-Type: application/json` and a body of at most 64 KiB. Media type parameters are allowed, so `application/json; charset=utf-8` is accepted. The body is a JSON object whose fields are strings (D4). Unknown fields are rejected, and `null` counts as a missing field.
- The body is decoded strictly (D35). Member names are case-sensitive, so `"A"` is an unknown field. A duplicate member, invalid UTF-8 or any data after the object is `invalid_json`.
- A success is `200` with `Content-Type: application/json` and the body `{"value": "...", "decimal": "...", "kind": "..."}` (D21).
- An error uses RFC 9457 Problem Details with `Content-Type: application/problem+json` (D5). See [Errors](#errors).
- The API is stateless and has no authentication.
- A path that is not in clean form, with an empty segment (`//`) or a `.` or `..` segment, gets `307 Temporary Redirect` with a `Location` header that holds the clean path, for example `/api/v1//add` to `/api/v1/add` (D38). It is the standard behavior of Go's `ServeMux`. A `307` keeps the method and the body, so a client that follows it sends the same request to the clean path. The redirect is not a Problem Details response.

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
| `POST /api/v1/evaluate` | `{"expression"}` | the expression, see [Expressions](#5-expressions) | `invalid_expression`, `expression_too_complex`, and every error of the operations it contains |

- In `percentage`, `operator` is one of `add`, `subtract`, `multiply` or `divide`.
- Every endpoint can also return `invalid_json`, `missing_field`, `unknown_field`, `invalid_number`, `operand_out_of_range`, `result_too_large`, `result_too_precise` and, if the server has a bug, `internal_error`.

### Health

`GET /healthz` returns `200` with `Content-Type: application/json` and the body `{"status": "ok"}` (D32). It sits outside `/api/`, so the rule that every path under `/api/` accepts only `POST` has no exception. `HEAD /healthz` returns the same status and headers without a body. Any other method returns `method_not_allowed` with `Allow: GET, HEAD`.

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
| `position` | In `evaluate`, the index of the character at fault in the expression (extension member, see [Expression errors](#expression-errors)) |

Request errors name the field they are about. Calculation errors use these fields:

| Error | `field` |
|---|---|
| `division_by_zero` in `divide` | `b` |
| `division_by_zero` in `percentage` | `percent` |
| `division_by_zero` in `power` (0 raised to a negative exponent) | `base` |
| `negative_square_root` | `value` |
| `exponent_not_integer`, `exponent_out_of_range` | `exponent` |
| `result_too_large`, `result_too_precise` | none |

Example:

```http
HTTP/1.1 422 Unprocessable Entity
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

The text after the status code in the status line is informational, and clients rely only on the code. Go writes `422 Unprocessable Entity`, the name from RFC 4918; RFC 9110 renamed it `Unprocessable Content` with the same meaning.

`400` means the request is wrong and the client must fix it. `422` means the request is well formed, but the result does not exist or cannot be represented (D17). `500` means the server failed, whatever the request (D37).

| Code | Status | Title | When |
|---|---|---|---|
| <a id="invalid_json"></a>`invalid_json` | 400 | Invalid JSON | The body is not a single valid JSON object: a syntax error, invalid UTF-8, a duplicate member, data after the object, or a top-level value that is not an object. |
| <a id="missing_field"></a>`missing_field` | 400 | Missing field | A required field is absent or `null`, or `percentage` has `base` without `operator` or the reverse. |
| <a id="unknown_field"></a>`unknown_field` | 400 | Unknown field | The body has a field that the endpoint does not define. Names are case-sensitive. |
| <a id="invalid_number"></a>`invalid_number` | 400 | Invalid number | A number field is not a JSON string, or the string does not match the [operand syntax](#operand-syntax). |
| <a id="operand_out_of_range"></a>`operand_out_of_range` | 400 | Operand out of range | An operand is outside the [limits](#limits). |
| <a id="exponent_not_integer"></a>`exponent_not_integer` | 400 | Exponent not integer | The exponent is not an exact integer. |
| <a id="exponent_out_of_range"></a>`exponent_out_of_range` | 400 | Exponent out of range | The exponent is below −1000 or above 1000. |
| <a id="invalid_operator"></a>`invalid_operator` | 400 | Invalid operator | `operator` is not one of the JSON strings `add`, `subtract`, `multiply` or `divide`. |
| <a id="invalid_expression"></a>`invalid_expression` | 400 | Invalid expression | `expression` is not a JSON string, or it does not follow the [expression syntax](#syntax): a character outside the syntax, a missing or unexpected token, or an unclosed group. |
| <a id="expression_too_complex"></a>`expression_too_complex` | 400 | Expression too complex | The expression has more than 100 operations, more than 10 powers, or groups nested more than 100 levels deep. |
| <a id="division_by_zero"></a>`division_by_zero` | 422 | Division by zero | The divisor is zero: `b` in `divide`, a percent of 0 with `divide`, or 0 raised to a negative exponent. |
| <a id="negative_square_root"></a>`negative_square_root` | 422 | Negative square root | `sqrt` received a negative value. |
| <a id="result_too_large"></a>`result_too_large` | 422 | Result too large | \|result\| ≥ 10⁴⁰. |
| <a id="result_too_precise"></a>`result_too_precise` | 422 | Result too precise | An exact terminating result needs more than 1000 fractional digits, an exact repeating result has a reduced denominator above 10¹⁰⁰⁰, or a non-zero approximate result would round to 0 at 1000 fractional digits. |
| <a id="not_found"></a>`not_found` | 404 | Not found | No endpoint exists at the path. |
| <a id="method_not_allowed"></a>`method_not_allowed` | 405 | Method not allowed | The path exists, but not for this method: the operations accept only `POST`, and `/healthz` accepts `GET` and `HEAD`. The `Allow` header lists the accepted methods. |
| <a id="payload_too_large"></a>`payload_too_large` | 413 | Payload too large | The body is larger than 64 KiB. |
| <a id="unsupported_media_type"></a>`unsupported_media_type` | 415 | Unsupported media type | `Content-Type` is missing, or its media type is not `application/json`. |
| <a id="internal_error"></a>`internal_error` | 500 | Internal error | The calculation failed in a way the API does not expect, which is a bug in the server. The response gives no cause; the server records it in its log (D37). |

When a request has several problems, the API reports only the first one, checked in this order:

1. Path and method. A path that is not in clean form is redirected before this check (see [Conventions](#conventions)).
2. Media type.
3. Body size.
4. JSON syntax.
5. Unknown fields, the first one in document order.
6. Missing fields, in the order the [endpoint table](#endpoints) lists them.
7. Operand syntax, field by field in that same order. In `percentage`, `operator` is checked here as well, after `percent` and `base`. A field that is not a JSON string is `invalid_number`, or `invalid_operator` for `operator`.
8. Operand limits, field by field.
9. Exponent rules.
10. The calculation itself, including `internal_error`.

`evaluate` follows steps 1 to 6, and then its own order, described in [Expression errors](#expression-errors).

## 5. Expressions

`POST /api/v1/evaluate` computes an expression with several operations (D41). The frontend sends every calculation through it.

### Request and response

- The body is `{"expression": "..."}`, with the [conventions](#conventions) of the other endpoints: `application/json`, at most 64 KiB, strict decoding and unknown fields rejected.
- `expression` is required. A value that is not a JSON string is `invalid_expression` with `field: "expression"`.
- A success is `200` with `{"value", "decimal", "kind"}`, the same shape as the other endpoints, so its `value` can be chained (D21).

### Syntax

An expression is ASCII text made of these tokens. Spaces (U+0020) may separate tokens, but not split one (D50).

| Token | Form | Examples |
|---|---|---|
| Number | A decimal of the [operand syntax](#operand-syntax), without a sign | `12`, `0.5`, `1e-3`, `2.5E10` |
| Value | `[`, any operand of the [operand syntax](#operand-syntax), `]` | `[1/7]`, `[-2.5]`, `[~1.4142135623730950488]` |
| Operator | `+`, `-`, `*`, `/`, `^` or `%` | |
| Square root | `sqrt(`, then an expression, then `)` | `sqrt(2)`, `sqrt(9/4)` |
| Group | `(`, then an expression, then `)` | `(2+3)*4` |

The grammar, from the lowest to the highest precedence (D42):

```text
sum      = term { ("+" | "-") term }
term     = signed { ("*" | "/") signed }
signed   = "-" signed | power
power    = postfix [ "^" signed ]
postfix  = primary [ "%" ]
primary  = number | value | "(" sum ")" | "sqrt(" sum ")"
```

- `^` groups from right to left: `2^3^2` = 2^(3^2) = 512. Its exponent may be negated: `2^-2` = 0.25.
- Unary minus binds below `^`: `-2^2` = −4, and `(-2)^2` = 4.
- `*` and `/` bind tighter than `+` and `-`, and both levels group from left to right: `2+3*4` = 14, and `8/4/2` = 1.
- `%` follows a number, a value, a group or a square root, once: `15%%` is `invalid_expression`.
- There is no implicit multiplication: `2(3)` and `2sqrt(9)` are `invalid_expression` (D46).
- A number starts with a digit or `.` and runs over digits, `.`, `e` and `E`, plus a `+` or `-` right after `e` or `E`. A run that does not match the decimal operand syntax is `invalid_number`, for example `1.2.3`, `.5` or `2e`.
- A result goes back into an expression as a value in brackets, which keeps exact fractions and the `~` mark (D21, D22, D50). A fraction cannot travel as a division: as a number on its own, its denominator can exceed the operand limit of 10⁴⁰.

### Meaning

- Every operation applies the rules of [Numbers](#2-numbers) and [Exactness and precision](#3-exactness-and-precision) as if it were computed alone (D44). Every intermediate result must stay within the [limits](#limits). An approximate intermediate result enters the next operation as its `value`, so `sqrt(2)*sqrt(2)` returns `~2`, as when the two steps are chained.
- `-x` is the negation of x. `sqrt(x)` is the square root. `x^y` follows the power rules: the exponent must be an exact integer between −1000 and 1000, and 0 raised to a negative exponent is a division by zero (D8, D18).
- Percentages (D43):
  - A term is what `+` and `-` separate: a number, a value, a group or a square root, or a product, quotient, power or negation of them.
  - A term that is only a percentage, B% with or without minus signs in front, applies to the term T right before it when it follows `+` or `-`: T + B% = T + T × B ÷ 100, and T − B% = T − T × B ÷ 100. The pair then acts as one term and keeps the sign in front of T.
  - So `50+200+15%` = 50 + 230 = 280, `50-200+15%` = 50 − 230 = −180, and `50-200-15%` = 50 − 170 = −120.
  - Percentages in a row compound: `200+15%+10%` = 230 + 10% = 253.
  - A group is a term, so parentheses give the percentage of a subtotal: `(50+200)+15%` = 287.5.
  - Anywhere else, B% = B ÷ 100: `200*15%` = 30, `200/15%` = 4000/3, `15%+200` = 200.15, `200-15%*2` = 199.7, and `2^50%` is `exponent_not_integer`.
  - Every case of D7 returns the same result as the `percentage` endpoint.
- Evaluation runs from left to right. An operation runs as soon as its operands are known, and in each sum the percentages are applied to their terms before the terms are added and subtracted. Evaluation stops at the first operation that fails.

### Size limits

- An expression has at most 100 operations, and at most 10 of them are powers (D45). Every `+`, `-`, `*`, `/`, `^`, `%` and `sqrt` counts as one operation, unary minus included.
- Groups and square roots nest at most 100 levels deep (D50).
- Breaking either limit is `expression_too_complex`.

### Expression errors

Errors are Problem Details, as in [Errors](#errors). An error about the text of the expression or about one of its operations has a `position` member in place of `field`. `position` is a JSON number: the 0-based index of the character at fault. An expression is ASCII, so the index is the same in characters, bytes and UTF-16 code units.

| Code | `position` points to |
|---|---|
| `invalid_expression` | The first character that cannot continue the expression: an unexpected token, a character outside the syntax, or the end of the expression (its length) when an operand is missing. An unclosed `(`, `sqrt(` or `[` points to its opening character. A missing operand inside an unclosed group is reported first, so `(2+` points to the end. |
| `invalid_number`, `operand_out_of_range` | The first character of the number, or the `[` of the value. |
| `expression_too_complex` | The first operator beyond the limit, or the opening character that nests too deep. |
| `exponent_not_integer`, `exponent_out_of_range` | The `^`. |
| `division_by_zero` | The `/`, or the `^` when 0 is raised to a negative exponent. |
| `negative_square_root` | The `s` of `sqrt(`. |
| `result_too_large`, `result_too_precise` | The operator whose result breaks a limit: `+`, `-`, `*`, `/`, `^`, `%` or the `s` of `sqrt(`. For T + B% and T − B%, it is the `%`. |

The status of each code is the one in [Errors](#errors). After steps 1 to 6 of the general check order, `evaluate` checks:

7. `expression` is a JSON string (`invalid_expression` with `field`).
8. The syntax, read from left to right. The first problem met is reported: `invalid_expression`, `invalid_number`, or `expression_too_complex` for nesting.
9. The operand limits of numbers and values, from left to right (`operand_out_of_range`).
10. The number of operations and powers (`expression_too_complex`).
11. The evaluation, including `internal_error`.

### Expression examples

Every row was checked with exact rational arithmetic.

| Expression | `value` | `kind` |
|---|---|---|
| `0.1+0.2` | `0.3` | `exact` |
| `2+3*4` | `14` | `exact` |
| `(2+3)*4` | `20` | `exact` |
| `2^3^2` | `512` | `exact` |
| `-2^2` | `-4` | `exact` |
| `2^-2` | `0.25` | `exact` |
| `1/3*3` | `1` | `exact` |
| `[1/7]*7` | `1` | `exact` |
| `sqrt(9/4)` | `1.5` | `exact` |
| `sqrt(2)*sqrt(2)` | `~2` | `approximate` |
| `200+15%` | `230` | `exact` |
| `200-15%` | `170` | `exact` |
| `200*15%` | `30` | `exact` |
| `200/15%` | `4000/3` | `repeating` |
| `15%` | `0.15` | `exact` |
| `50+200+15%` | `280` | `exact` |
| `50-200+15%` | `-180` | `exact` |
| `200+15%+10%` | `253` | `exact` |
| `(50+200)+15%` | `287.5` | `exact` |
| `15%+200` | `200.15` | `exact` |
| `200+-15%` | `170` | `exact` |

| Expression | Error | `position` |
|---|---|---|
| `1/0` | `division_by_zero` | 1 |
| `0^-1` | `division_by_zero` | 1 |
| `sqrt(-4)` | `negative_square_root` | 0 |
| `2^0.5` | `exponent_not_integer` | 1 |
| `2^1001` | `exponent_out_of_range` | 1 |
| `10^50/10^20` | `result_too_large` | 2 |
| `2+` | `invalid_expression` | 2 |
| `(2+3` | `invalid_expression` | 0 |
| `(2+` | `invalid_expression` | 3 |
| `2(3)` | `invalid_expression` | 1 |
| `15%%` | `invalid_expression` | 3 |
| `1.2.3` | `invalid_number` | 0 |

Request example:

```http
POST /api/v1/evaluate
Content-Type: application/json

{"expression": "50+200+15%"}
```

```http
HTTP/1.1 200 OK
Content-Type: application/json

{"value": "280", "decimal": "280", "kind": "exact"}
```

Error example:

```http
HTTP/1.1 422 Unprocessable Entity
Content-Type: application/problem+json

{
  "type": "https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#division_by_zero",
  "title": "Division by zero",
  "status": 422,
  "detail": "The expression divides by zero at position 1.",
  "code": "division_by_zero",
  "position": 1
}
```

## Decisions log

### Block 0: Foundation

| # | Decision | Reason |
|---|---|---|
| D1 | **Partly superseded by D40 and D41.** Form-based calculator with chaining. Operators (+, −, ×, ÷, xʸ, √, %) are clickable buttons. Operands are entered through an on-screen keypad (digits, `.`, ±, backspace) or the physical keyboard. After a result, pressing an operator reuses that result as the first operand. | Reusing results and entering calculations comfortably is the top UI priority, and users must not need to know keyboard shortcuts. One API call per step keeps the backend the source of truth. Some mobile number keyboards lack a minus key, which the ± button covers. |
| D2 | Division is presented in the UI as a fraction (a/b). | The owner asked for a clickable fraction element. It needs no API change. |
| D3 | **Extended by D48.** A history icon in the display opens the last 100 calculations, saved in the browser (localStorage). Clicking an entry loads its operands, operator and result. The 101st entry drops the oldest. Frontend only. | The owner wants to go back to earlier calculations and results, and the history must survive a reload. |
| D4 | Numbers travel as JSON strings in requests and responses. JSON numbers are rejected as invalid input. | An experiment showed that JavaScript `JSON.parse` turns `12345678901234567890.123` into `12345678901234567000`, and Go's default decoder turns it into a `float64`. Strings are exact in both languages without special parsing. |
| D5 | Errors use RFC 9457 Problem Details (`application/problem+json`) with a machine-readable `code` extension member. | It is a standard, self-describing format, and `code` gives clients a stable value to branch on. |
| D6 | The API contract is written in Markdown in this file. | One source of truth, readable on GitHub and by CodeRabbit, with no extra tooling. Nothing in scope consumes an OpenAPI file. |
| D7 | **Extended by D43.** Percentage divides its number by 100 and follows the preceding operator: A + B% = A + A·B/100, A − B% = A − A·B/100, A × B% = A·B/100, A ÷ B% = A ÷ (B/100), and B% alone = B/100. | These are the pocket calculator semantics the owner described: 200 + 15% = 230, 200 − 15% = 170, 200 × 15% = 30, 200 ÷ 15% = 1333.33…, 15% = 0.15. |
| D8 | Exponentiation accepts integer exponents, negative ones included. | Positive exponents stay exact, and integer exponents never produce complex-number cases. |
| D9 | **Superseded by D21 and D24.** Exact results are returned exactly, with up to 1000 fractional digits. Results that need approximation (repeating decimals, irrational roots) are rounded to 20 decimal places with round half even. An exact result that exceeds the supported limits is an error and is never silently rounded. | This keeps calculation precision separate from display precision. Half even avoids the upward bias of half up across many operations. |
| D10 | **Superseded by D23.** Every successful response states the kind of result: `exact`, `repeating` or `irrational`. | The display rule (D11) depends on it, and the backend can prove it. A reduced fraction repeats unless its denominator has no prime factors other than 2 and 5, and the square root of a rational number is irrational unless the numerator and the denominator are perfect squares. |
| D11 | **The stored value and the copy rule are superseded by D21 and D25.** Stored precision and display precision are separate. The stored value is the complete result the API returned (D9): up to 1000 fractional digits for exact results and 20 for approximated ones. The display is frontend only and shows at most 100 fractional digits, truncated and never rounded. At first it shows 4 decimals and an ellipsis for `repeating` results, and 5 decimals and an ellipsis for every other result with more than 5 decimals. Results with 5 or fewer decimals show in full with no ellipsis. Each click on the ellipsis expands the result to 20, 40, 60, 80 and then 100 fractional digits. Expansion stops early when the result has fewer digits, and stops at 100 even if more are stored. The ellipsis disappears once every displayable digit is visible. Expanding never recalculates or changes the stored value. Copying and chaining always use the complete stored value, whatever is visible. Long numbers never break the layout. | The owner prioritizes both comfortable reading and mathematical precision: the display stays short and readable, and no stored digit is ever lost. |
| D12 | **The limits are superseded by D24, and D22 adds formats.** Operands are plain decimals with an optional leading `-` and an optional `.`, and scientific notation (`1e-3`, `2.5E10`) is also accepted. A leading `+`, thousands separators and `,` as the decimal separator are rejected. Operands and results share the same limits, 40 integer digits and 1000 fractional digits, measured after expanding scientific notation. | One unambiguous syntax plus scientific notation for API users. Shared limits guarantee that every result is a valid operand, so chaining always works. |
| D13 | **Extended by D22.** Results are canonical: no trailing zeros, no leading zeros, no `-0` and no scientific notation (`1.10 + 1.20` returns `2.3`). | Each value has one representation, which keeps the API and the tests simple. The UI can still format results. |
| D14 | **Extended by D41.** One endpoint per operation: `POST /api/v1/{operation}` with a JSON body. Percentage takes the percent plus an optional base and preceding operator. | Versioned paths and JSON bodies of strings follow from D4. Percentage needs the operator for D7, and keeping it in one call puts every percent rule in Go, where it is tested. |
| D15 | Repository layout: the Go module in `backend/`, the Vite app in `frontend/`, and documentation at the root. | Each layer is self-contained, and the layout matches the path instructions in `.coderabbit.yaml`. |
| D16 | **The response shape is superseded by D21.** Request bodies: `{"a","b"}` for add, subtract, multiply and divide; `{"base","exponent"}` for power; `{"value"}` for square root; `{"percent"}` for percentage, with optional `"base"` and `"operator"` (`add`, `subtract`, `multiply` or `divide`), which must be given together. A successful response is `200 {"result": "...", "kind": "..."}`. | Field names state each operand's role, and the success shape is the same for every operation. |
| D17 | **D24 restates the limits behind `result_too_large` and `result_too_precise`.** HTTP status codes: `400` when the request itself is wrong (malformed JSON, missing field, invalid or out-of-range number, non-integer or out-of-range exponent, invalid operator); `422` when the request is well formed but the result does not exist or cannot be represented (division by zero, square root of a negative number, `result_too_large` for more than 40 integer digits, `result_too_precise` for more than 1000 fractional digits); `404`, `405`, `413` for an oversized body, and `415` for a body that is not JSON. | Clients can tell a request they must fix from a calculation that has no representable answer. |
| D18 | Power edge cases: `0^0 = 1`, `0` raised to a negative exponent is a division by zero, and the exponent must be between −1000 and 1000. | `0^0 = 1` is the usual convention in calculators and programming languages. The exponent range bounds the cost of a single request. Most larger exponents would exceed the result limits anyway, but a few representable results are rejected too, such as `1.0001^-2000` (about 0.82). |
| D19 | The UI text is in English. | Everything in the repository is in English. |
| D20 | The out of scope list below is approved. | It keeps the calculator focused on the operations and behaviors in this document. |
| D21 | Every result keeps its exact mathematical value, and the API stays stateless. A successful response is `200 {"value": "...", "decimal": "...", "kind": "..."}`. `value` is used for chaining and stored in the history. `1 ÷ 3` returns `"value": "1/3"`, and chaining `× 3` on it returns exactly `"1"`. | The owner requires exact fractions to survive chaining and history reuse: 1 ÷ 3 × 3 must be 1, not 0.99999999999999999999. |
| D22 | An operand or a `value` takes one of three forms: a decimal (`-12.5`, `1e-3`), a fraction `p/q` (integers, sign on the numerator, `q > 0`), or an approximate decimal with a `~` prefix (`~1.4142135623730950488`). Fractions in requests need not be reduced (`4/2` is read as 2). A fraction `value` in a response is always reduced and is used only when its decimal does not terminate (`1/4` is written `0.25`). If any operand is approximate, the result is approximate. | Each value string describes itself, so the frontend can send back exactly what it received, request bodies stay flat, and the API needs no state to know whether a value is exact. |
| D23 | `kind` is `exact` (a terminating decimal), `repeating` (an exact rational whose decimal does not terminate) or `approximate` (an irrational result, or one computed from an approximate operand). | With exact fractions, only approximation needs a warning. Approximation also spreads to every later result, not just to irrational ones. |
| D24 | `decimal` is for display only. Exact results include every digit. Repeating results are rounded to at most 100 fractional digits and approximate results to at most 20, both with round half even and without trailing zeros (D13). The `value` of an approximate result is `~` followed by its `decimal`. When 0 < \|x\| < 10⁻²⁰, `decimal` uses scientific notation, and the mantissa keeps the same number of fractional digits. Limits: \|x\| < 10⁴⁰ (otherwise `result_too_large`); an exact value has at most 1000 fractional digits when its decimal terminates, and a reduced denominator of at most 10¹⁰⁰⁰ when it does not; and an approximate value needs at most 1000 fractional digits in plain notation, so a non-zero approximate result must not round to 0 at that precision. Breaking either of the last two limits returns `result_too_precise`. A non-zero result never becomes 0. | The owner requires that tiny non-zero results never become 0 and keep their magnitude, for example 1 ÷ 3e30 ≈ 3.3333… × 10⁻³¹. Terminating values keep the owner's limit of 1000 fractional digits. A denominator limit alone would allow up to about 3321 digits (`1/2^3321`), so the owner chose the digit limit in the PR #1 review. Operands share every limit, so chaining always works. |
| D25 | Copying puts `decimal` on the clipboard with all its digits. Chaining and history use `value`. The display rules of D11 apply to `decimal`, so repeating results can expand to 100 digits, and tiny values are shown in scientific notation. | A decimal pastes correctly into other apps, while a fraction such as `1/3` can be misread (a spreadsheet turns it into a date). |
| D26 | Frontend tooling: ESLint with typescript-eslint, and Prettier, with TypeScript pinned to `~6.0.3`. | The master prompt asks for ESLint and Prettier. The default Vite React-TS template (create-vite 9.3.4) ships oxlint and TypeScript 7.0, but typescript-eslint 8.72 supports only TypeScript below 6.1. The template's `--eslint` option ships ESLint with TypeScript `~6.0.3`, so the frontend is scaffolded with it. |
| D27 | During local development the backend listens only on `127.0.0.1:8080`. The Docker setup (Block 3) sets the address its environment needs. | The owner does not want the development server reachable from other machines on the network. |

### Block 1: Backend in Go

| # | Decision | Reason |
|---|---|---|
| D28 | The arithmetic uses only Go's standard `math/big`: `big.Rat` holds every exact value and `big.Int.Sqrt` computes square roots. The backend implements its own operand parser and its own round half even. | An experiment with Go 1.27.3 compared shopspring/decimal 1.6.0, cockroachdb/apd 3.2.3 and `math/big`. Both decimal libraries store `1 ÷ 3` as a rounded decimal, so `1 ÷ 3 × 3` returns `0.999…` and breaks D21. shopspring also loses precision silently: by default `1 ÷ 3e30` returns `0` and `0 ^ −1` returns `0` without an error. `big.Rat` keeps `1/3` exactly, and `Rat.FloatPrec` tells whether a decimal terminates. Two documented gaps are covered by our own code: `Rat.FloatString` rounds halves away from zero, and `Rat.SetString` accepts forms outside the operand syntax, such as `+5` and `0x10`. |
| D29 | The HTTP layer is the standard `net/http` `ServeMux` with method patterns (`POST /api/v1/add`). No router dependency. The backend answers `not_found` and `method_not_allowed` itself in Problem Details. | Seven fixed routes need no router. Both the standard mux and chi need custom 404 and 405 handlers anyway, because the built-in ones answer in `text/plain`. |
| D30 | Package layout: `backend/main.go` starts the server, `backend/internal/calc` holds the number model (parsing, arithmetic, rounding, formatting) and `backend/internal/api` holds the HTTP handlers. | It separates the arithmetic from HTTP and allows coverage per package (D34). `go run .` keeps working, and with a single binary a `cmd/` directory adds nothing. |
| D31 | During development the Vite dev server proxies `/api` to the backend. The backend sends no CORS headers. | The browser sees one origin, so no CORS code is needed. A cross-origin `POST` with `Content-Type: application/json` triggers an `OPTIONS` preflight, which the API answers with `405`. |
| D32 | `GET /healthz`, outside `/api/`, returns `200` with `{"status": "ok"}`. | Docker (Block 3) can check that the server is alive, and the rule that every path under `/api/` accepts only `POST` stays intact. |
| D33 | The listen address comes from the environment variable `PRECIZZLE_ADDR`, `127.0.0.1:8080` by default. | It is the only setting D27 needs to vary. The prefix avoids clashes with other variables. |
| D34 | CI fails when statement coverage drops below 95% in `internal/calc` or below 90% in `internal/api`. | It protects the arithmetic without forcing artificial tests of trivial code such as server startup. |
| D35 | The request body is decoded strictly with `encoding/json/v2`: member names are case-sensitive (`"A"` is `unknown_field`), a duplicate member is `invalid_json`, and any data after the object is `invalid_json`. | `encoding/json` (v1) accepts `"A"` as `a` and keeps the last of two duplicate members. A client could send two values without knowing which one was used. In Go 1.27.3, v2 needs no experiment flag. |
| D36 | The notation of a `decimal` is chosen from the value the API returns, after rounding. | The owner requires that a returned value reads the same when it is sent back in a later calculation. Choosing the notation before rounding wrote one value two ways: an approximate √(0.99…9 × 10⁻⁴⁰) gave `1e-20`, but sending that value back gave `0.00000000000000000001`. |
| D37 | If package `calc` returns an error that the API does not map, the API logs the error and answers `500 internal_error` in Problem Details, with a generic detail and no `field`. | The owner wants an internal failure reported as one: never a crash, and never a math code that hides the bug. The cause stays in the server log, so the response reveals no implementation details. |
| D38 | Paths that are not in clean form keep the standard `ServeMux` behavior: `307 Temporary Redirect` to the clean path. | The owner chose not to change the standard router behavior. A `307` keeps the method and the body, so following it reaches the right endpoint. CodeRabbit noted in PR #2 that the SPEC did not document it. |

### Block 2: Expressions and frontend

| # | Decision | Reason |
|---|---|---|
| D39 | Visual direction "Equation". The calculation is an editable equation of tappable boxes, the result sits in its own mint card, and division is drawn as a stacked fraction. The background is white. Dark purple (`#392757`) is used for text and selected elements. Accents come from the reference palette: mint (`#3ED9AB`) and green (`#12C488`) for results and the equals key, purple (`#A346D0`) for the active box and focus, coral (`#EE6072`) for errors, and orange (`#FF7925`) for approximate results. The typeface is Nunito with tabular figures. | The owner chose it from three mockups (Pocket, Equation, Ledger). Every part of a calculation has its own place and color, chaining is visible, and D2's fraction is drawn literally. Mint, green and orange reach only 1.8 to 2.6:1 contrast as text on white, below the WCAG minimum of 4.5:1, so they are used as fills under dark text or in darker shades for text. |
| D40 | Input is an expression editor with a real cursor: a thin vertical bar that the user places by clicking or tapping between characters, or moves with the arrow keys. Digits and operators are inserted at the cursor without replacing anything, and any part can be deleted. Fractions, square roots and exponents are drawn as structures with editable parts. On-screen buttons and the physical keyboard behave the same way. | The owner wants to insert, change or delete any part of a calculation without starting over. The original prompt is in `PROMPTS.md`. |
| D41 | A new endpoint, `POST /api/v1/evaluate`, evaluates expressions with several operators. Its body is `{"expression": "..."}`: an ASCII string with numbers, values in brackets (D50), the operators `+ - * / ^ %`, `sqrt(...)` and parentheses. A success has the usual shape. An error about the text of the expression or one of its operations carries a `position` member in place of `field`; request errors keep the usual members. The seven operation endpoints stay unchanged, and the frontend uses only `evaluate`. This removes "Expressions with several operators and operator precedence in one request" from the out of scope list, and replaces "one API call per step" in D1. | The owner chose free expressions after weighing three alternatives. Precedence is mathematical logic, so it stays in the backend. A frontend that applied precedence and chained requests would move math rules into the browser. A text string reads well in `curl` and README examples, and it follows D4. |
| D42 | Precedence follows the mathematical convention. Parentheses and structures come first. Then `^`, from right to left (`2^3^2` = 2⁹ = 512). Then unary minus (`-2^2` = −4). Then `*` and `/`, from left to right. Then `+` and `-`, from left to right. | It matches the notation the editor draws, where an exponent sits on its base. Spreadsheet conventions (−2² = 4, `^` from left to right) contradict that notation. |
| D43 | `%` extends D7 by applying to the operand it belongs to. A term that is only a percentage, B% with or without minus signs in front, applies to the term T right before it when it follows `+` or `-`: T + B% = T + T × B ÷ 100, and T − B% = T − T × B ÷ 100. The pair acts as one term and keeps the sign in front of T: `50 + 200 + 15%` = 50 + 230 = 280, and `50 - 200 + 15%` = 50 − 230 = −180. Percentages in a row compound: `200 + 15% + 10%` = 253. A group is a term, so `(50 + 200) + 15%` = 287.5. Anywhere else, B% = B ÷ 100: `200 * 15%` = 30, `200 / 15%` = 4000/3 and `15% + 200` = 200.15. | The owner wants a percentage to apply to the operand it follows, not to the subtotal, and corrected the first proposal, which gave 287.5 for `50 + 200 + 15%`. The sign in front of T covers T with its percentage, so a markup makes an expense larger and a discount makes it smaller: `50 - 200 - 15%` = −120. Successive percentages compound, like successive markups. Every case of D7 returns the same result as the `percentage` endpoint. |
| D44 | Every operation inside an expression applies the rules and limits of [Numbers](#2-numbers) and [Exactness and precision](#3-exactness-and-precision), as if it were computed alone. Each intermediate result must be a valid operand, an approximate intermediate result goes into the next operation as its `value`, and an error points to the operation that failed. `10^50 / 10^20` is `result_too_large`, and `sqrt(2) * sqrt(2)` returns `~2`. | The owner does not want expressions to change the math rules already defined. An expression returns exactly what the same steps return when they are chained. |
| D45 | An expression has at most 100 operations, and at most 10 of them are powers. A longer expression is `400 expression_too_complex`. | On the largest operands, a power takes about 85 ms on the development machine and every other operation takes at most 0.25 ms. The worst case stays near 1 s, far below the 10 s write timeout, and a user can still add a list of 100 amounts. |
| D46 | Two keys enter parentheses, `(` and `)`. Each one inserts exactly its parenthesis at the cursor, and the user can move, insert and delete parentheses like any other character. An unbalanced expression is reported when it is evaluated. Pressing √ or `(` right after a number, a `)` or a `%` first inserts a visible ×, so 2 followed by `(` gives 2 × (. The API accepts only explicit operators. The Equation layout places both keys so that it stays balanced on desktop and on mobile. | The owner prefers two explicit keys to one smart key, and kept the visible × before `(`. Fractions, roots and exponents already group their parts, and parentheses cover cases such as (2 + 3) × 4. The visible × keeps what the user sees equal to what is computed, with no implicit multiplication rule. |
| D47 | An error is shown next to the part of the expression at fault, with a message that the frontend writes for each `code`: what happened and how to fix it. The keypad and the keyboard produce only well-formed numbers, and the backend checks the limits. The expression is kept so the user can fix it. | Frontend messages can use the words the user sees, and `position` marks the exact spot. The limits stay written in one place, the backend. |
| D48 | Each history entry offers two actions. "Load" restores its expression and result, as in D3. "Use" inserts its exact `value` at the cursor as a result chip. A "Clear history" action deletes every entry after a confirmation inside the history view. This extends D3. | The owner prioritizes reusing results anywhere in a calculation. The history can hold sensitive amounts on a shared device, so the user can clear it. |
| D49 | Frontend tests use Vitest, React Testing Library and jsdom. Mock Service Worker (MSW) intercepts the API requests at the network level. CI fails when statement coverage of `frontend/src` drops below 90%, not counting the entry point `main.tsx`. | MSW lets tests check the requests the app really sends (URL, method, headers and body), as the CodeRabbit path instructions ask. The threshold matches `internal/api` (D34). |
| D50 | Expression syntax details. Spaces may separate tokens. A result goes back into an expression as a value in brackets, such as `[1/7]` or `[~1.4142135623730950488]`, read with the operand syntax of [Numbers](#2-numbers). A number in an expression has no sign, so a leading `-` is unary minus, and `sqrt(` is one token. Groups and square roots nest at most 100 levels deep. An error about the expression carries `position`, the 0-based index of the character at fault, and there are two new codes: `invalid_expression` and `expression_too_complex`, both `400`. Every `+`, `-`, `*`, `/`, `^`, `%` and `sqrt` counts as one operation for D45. | A fraction cannot travel as a division: as a number on its own, its denominator can exceed the operand limit of 10⁴⁰. The brackets keep exact fractions and the `~` mark (D21, D22). `position` lets the frontend mark the exact spot (D47). The nesting limit bounds the parser's recursion, as the body limit bounds its input. |

### Block 2a: Expression API

Decisions taken while building the expression API, numbered from D51.

| # | Decision | Reason |
|---|---|---|

### Block 2b: Frontend

Decisions taken while building the frontend, numbered from D71.

| # | Decision | Reason |
|---|---|---|

## Out of scope

- User accounts and authentication.
- Server-side storage of the history (the history lives in the browser, D3).
- Fractional exponents, roots other than the square root, and complex numbers.
- Scientific functions such as trigonometry and logarithms.
- Memory keys (M+, M−, MR).
- Thousands separators, `,` as the decimal separator, currency formatting and internationalization.
