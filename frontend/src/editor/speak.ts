import type { Row } from "./model.ts";

const words = {
  "+": "plus",
  "-": "minus",
  "*": "times",
  "%": "percent",
  "(": "open parenthesis",
  ")": "close parenthesis",
};

function numberWords(value: string): string {
  return value
    .replace(/e/i, " times ten to the power of ")
    .replace("~", "approximately ")
    .replace("/", " over ")
    .replaceAll("-", "minus ")
    .replaceAll("+", "plus ");
}

export function speak(row: Row): string {
  const parts: string[] = [];
  let number = "";
  const flush = () => {
    if (number) parts.push(number);
    number = "";
  };
  for (const node of row) {
    if (node.type === "char" && /^[0-9.]$/.test(node.char)) {
      number += node.char;
      continue;
    }
    flush();
    switch (node.type) {
      case "char":
        parts.push(words[node.char as keyof typeof words]);
        break;
      case "chip":
        parts.push(numberWords(node.value));
        break;
      case "frac":
        parts.push(
          `fraction ${speak(node.num)} over ${speak(node.den)} end fraction`,
        );
        break;
      case "sqrt":
        parts.push(`square root of ${speak(node.body)} end root`);
        break;
      case "pow": {
        const exp = speak(node.exp);
        parts.push(
          exp === "2"
            ? "squared"
            : exp === "3"
              ? "cubed"
              : `to the power of ${exp} end exponent`,
        );
        break;
      }
    }
  }
  flush();
  return parts.join(" ") || "empty";
}
