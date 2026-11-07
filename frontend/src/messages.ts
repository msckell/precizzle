const messages: Record<string, string> = {
  invalid_json:
    "The calculation request could not be read. Reload the page and try again.",
  missing_field:
    "The calculation request is missing a required field. Reload the page and try again.",
  unknown_field:
    "The calculation request includes an unexpected field. Reload the page and try again.",
  invalid_number:
    "This number is not valid. Enter digits with at most one decimal point.",
  operand_out_of_range:
    "This number exceeds the supported limits. Use a smaller number or fewer digits.",
  exponent_not_integer:
    "The exponent must be an exact integer. Replace it with a whole number without an approximate value.",
  exponent_out_of_range:
    "The exponent is outside the supported range. Use an integer between -1000 and 1000.",
  invalid_operator:
    "This operator is not supported. Choose an operator from the keypad.",
  invalid_expression:
    "The expression is incomplete or contains an unexpected symbol. Fill empty slots and check the operators and parentheses.",
  expression_too_complex:
    "The expression exceeds 100 operations, 10 powers or 100 nested groups. Shorten it or calculate part of it first.",
  division_by_zero:
    "This operation divides by zero. Change the denominator or the base of the negative power.",
  negative_square_root:
    "A negative number has no real square root. Use a non-negative number inside the root.",
  result_too_large:
    "The result is too large to represent. Use smaller numbers or reduce the exponent.",
  result_too_precise:
    "The result needs more digits than the supported limits allow. Simplify the expression or use fewer fractional digits.",
  not_found:
    "The calculation service could not be found. Reload the page and try again.",
  method_not_allowed:
    "The service does not accept this request method. Reload the page and try again.",
  payload_too_large:
    "The expression is too long to send. Shorten it or calculate part of it first.",
  unsupported_media_type:
    "The service cannot read this request format. Reload the page and try again.",
  internal_error:
    "The service could not complete the calculation. Try again in a moment.",
  network_failure:
    "The calculation service could not be reached. Check your connection and try again.",
  unexpected_response:
    "The service returned an unexpected response. Reload the page and try again.",
};

export function messageFor(code: string): string {
  return Object.hasOwn(messages, code)
    ? messages[code]
    : messages.unexpected_response;
}
