// Derives runtime member arrays from the tygo-emitted enums.ts. Run by
// `make contracts` after `tygo generate`. Its output is generated — never
// hand-edit enum-values.ts.
//
// Why this exists: tygo's `union` enum style (the only style this codebase can
// use — a TS `enum` is nominal and rejects string literals) emits each member as
// an individual `export const` plus a `export type X = typeof A | typeof B`.
// There is no runtime object to iterate, so pickers would otherwise need a
// hand-written member list. This script is the single generator of that list.
//
// It fails loudly rather than emitting a shorter list: an unresolved member
// const, or a shape change that parses zero enums, is an error.
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const dir = join(dirname(fileURLToPath(import.meta.url)), "..", "src");
const src = readFileSync(join(dir, "enums.ts"), "utf8");

const snake = (s) => s.replace(/([a-z0-9])([A-Z])/g, "$1_$2").toUpperCase();

// `export const CEFRA1 = "A1";` — the plain union members. Typed consts
// (`export const X: SomeType = "…";`, the `= string` enums) are intentionally
// not matched: those enums have no closed member list to generate.
const consts = new Map();
for (const m of src.matchAll(/^export const (\w+) = ("[^"]*");$/gm)) {
  consts.set(m[1], m[2]);
}

const enums = [];
for (const m of src.matchAll(/^export type (\w+) = (typeof \w+(?: \| typeof \w+)*);$/gm)) {
  const members = [...m[2].matchAll(/typeof (\w+)/g)].map((x) => {
    if (!consts.has(x[1])) {
      throw new Error(`enum ${m[1]}: unresolved member const ${x[1]}`);
    }
    return consts.get(x[1]);
  });
  if (members.length === 0) throw new Error(`enum ${m[1]}: no members parsed`);
  enums.push([m[1], members]);
}
if (enums.length === 0) {
  throw new Error("parsed 0 enums from enums.ts — the emitted shape changed");
}

const body = enums
  .map(
    ([name, members]) =>
      `export const ${snake(name)}_VALUES = [${members.join(", ")}] as const satisfies readonly Enums.${name}[];`,
  )
  .join("\n");

writeFileSync(
  join(dir, "enum-values.ts"),
  "// Code generated from ./enums.ts by scripts/gen-enum-values.mjs. DO NOT EDIT.\n" +
    'import type * as Enums from "./enums.ts";\n\n' +
    `${body}\n`,
);

console.log(`enum-values.ts: ${enums.length} enums`);
