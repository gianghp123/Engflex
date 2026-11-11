import type { ExamCategory, ExamReviewContent } from "@engflex/contracts";
import { Store } from "@tanstack/store";

// Local authoring drafts only: no admin read/write endpoint exists yet.
export interface ExamDraft {
	id: string;
	slug: string;
	sourceKey: string;
	content: ExamReviewContent;
}
export type CategoryDraft = Pick<
	ExamCategory,
	"id" | "slug" | "title" | "type"
> & { tests: ExamDraft[] };
export function newExamDraft(): ExamDraft {
	const id = crypto.randomUUID();
	return {
		id,
		slug: "",
		sourceKey: "",
		content: {
			schemaVersion: "0.1.0",
			collectionId: "",
			testId: id,
			testNumber: 1,
			title: "",
			audioAssetId: "",
			assets: [],
			parts: [],
		},
	};
}
export function createExamsStore() {
	return new Store<CategoryDraft[]>([
		{
			id: "demo-category",
			slug: "sample-collection",
			title: "Sample TOEIC Collection",
			type: "toeic",
			tests: [
				{
					id: "demo-test",
					slug: "sample-test",
					sourceKey: "sample:test:1",
					content: {
						schemaVersion: "0.1.0",
						collectionId: "sample-collection",
						testId: "demo-test",
						testNumber: 1,
						title: "Sample TOEIC Test",
						audioAssetId: "",
						assets: [],
						parts: [
							{
								partNumber: 5,
								section: "reading",
								items: [
									{
										id: "sample-item",
										type: "single",
										questions: [
											{
												id: "sample-question",
												questionNumber: 101,
												stem: "The meeting starts ___ nine.",
												options: [
													{ key: "A", text: "at" },
													{ key: "B", text: "on" },
													{ key: "C", text: "in" },
													{ key: "D", text: "by" },
												],
												correctKey: "A",
												explanation: "Use at for a specific time.",
												vietnameseText: null,
												vocabulary: [],
												targetIds: [],
											},
										],
										passages: [],
										transcript: [],
										media: [],
										itemAudio: null,
										stimulusAudio: null,
									},
								],
							},
						],
					},
				},
			],
		},
	]);
}

/** Draft integrity, not a claim that an incomplete draft is publishable. */
export function validExamDraft(test: ExamDraft): boolean {
	if (!test.slug.trim() || !test.sourceKey.trim() || !test.content.title.trim())
		return false;
	const ids = new Set<string>(),
		numbers = new Set<number>(),
		parts = new Set<number>();
	const assets = new Set(test.content.assets.map((asset) => asset.sourceId));
	if (
		assets.size !== test.content.assets.length ||
		test.content.assets.some(
			(asset) =>
				!asset.sourceId.trim() || !/^[a-f0-9]{64}$/i.test(asset.sha256),
		)
	)
		return false;
	if (test.content.audioAssetId && !assets.has(test.content.audioAssetId))
		return false;
	function unique(id: string) {
		if (!id.trim() || ids.has(id)) return false;
		ids.add(id);
		return true;
	}
	for (const part of test.content.parts) {
		if (
			!Number.isInteger(part.partNumber) ||
			part.partNumber < 1 ||
			part.partNumber > 7 ||
			parts.has(part.partNumber) ||
			part.section !== (part.partNumber <= 4 ? "listening" : "reading")
		)
			return false;
		parts.add(part.partNumber);
		for (const item of part.items) {
			if (!unique(item.id) || !["single", "group"].includes(item.type))
				return false;
			const questionIds = new Set(
				item.questions.map((question) => question.id),
			);
			const nodes = new Set<string>();
			for (const block of [
				...item.media,
				...item.passages.flatMap((passage) => passage.blocks),
			]) {
				if (!unique(block.id) || (block.assetId && !assets.has(block.assetId)))
					return false;
				for (const node of block.inline) {
					if (
						!unique(node.id) ||
						(node.questionId && !questionIds.has(node.questionId))
					)
						return false;
					nodes.add(node.id);
				}
			}
			for (const passage of item.passages)
				if (!unique(passage.id)) return false;
			for (const line of item.transcript) if (!unique(line.id)) return false;
			for (const audio of [item.itemAudio, item.stimulusAudio]) {
				if (
					audio &&
					(!assets.has(audio.assetId) ||
						audio.startTime == null ||
						audio.endTime == null ||
						!Number.isFinite(audio.startTime) ||
						!Number.isFinite(audio.endTime) ||
						audio.startTime < 0 ||
						audio.endTime <= audio.startTime)
				)
					return false;
			}
			for (const question of item.questions) {
				if (
					!unique(question.id) ||
					!Number.isInteger(question.questionNumber) ||
					question.questionNumber < 1 ||
					question.questionNumber > 200 ||
					numbers.has(question.questionNumber)
				)
					return false;
				numbers.add(question.questionNumber);
				if (
					!question.options.length ||
					new Set(question.options.map((option) => option.key)).size !==
						question.options.length ||
					!question.options.some(
						(option) => option.key === question.correctKey,
					) ||
					question.targetIds.some((id) => !nodes.has(id))
				)
					return false;
			}
		}
	}
	return true;
}
