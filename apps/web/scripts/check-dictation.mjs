import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { checkDictation } from "../src/features/videos/dictation.ts";

const fixtures = new URL(
	"../../api/internal/modules/videos/services/testdata/dictation.json",
	import.meta.url,
);
for (const item of JSON.parse(readFileSync(fixtures, "utf8"))) {
	const result = checkDictation(item.content, item.draft, item.level, item.answers);
	const positions = result.slots
		.filter((slot) => slot.blankIndex !== undefined)
		.map((slot) => slot.position);
	assert.deepEqual(positions, item.positions, item.name);
	assert.equal(result.complete, item.complete, item.name);
}
