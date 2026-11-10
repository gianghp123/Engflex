import type { ExamReviewAttempt, ExamReviewContent } from "@engflex/contracts";
import { APP_ROUTES } from "#/app/app-route";
import { isToeicMediaKey } from "./media-key";
import { toeicPartLabel } from "./parts";

export type ReviewPart = ExamReviewContent["parts"][number];
export type ReviewItem = ReviewPart["items"][number];
export type ReviewQuestion = ReviewItem["questions"][number];
export type ReviewBlock = ReviewItem["media"][number];
export type ReviewAudio = NonNullable<ReviewItem["itemAudio"]>;
export type ReviewAnswer = NonNullable<
	ExamReviewAttempt["result"]
>["answers"][number];
export type AnswersById = ReadonlyMap<string, ReviewAnswer>;
export type ReviewFilter = "all" | "incorrect" | "correct";

export function questionStatus(
	question: ReviewQuestion,
	answers: AnswersById,
): "correct" | "incorrect" | "skipped" {
	const selected = answers.get(question.id)?.selectedKey;
	if (!selected) return "skipped";
	return question.correctKey !== null && selected === question.correctKey
		? "correct"
		: "incorrect";
}
export function partStatistics(part: ReviewPart, answers: AnswersById) {
	return {
		partNum: part.partNumber,
		name: toeicPartLabel(part.partNumber),
		totalQuestions: part.items.reduce(
			(sum, item) => sum + item.questions.length,
			0,
		),
		correctCount: part.items.reduce(
			(sum, item) =>
				sum +
				item.questions.filter((q) => questionStatus(q, answers) === "correct")
					.length,
			0,
		),
	};
}
export function mediaSource(
	assetId: string | null | undefined,
	assets: ExamReviewContent["assets"],
): string | null {
	const asset = assets.find((asset) => asset.sourceId === assetId);
	if (!assetId || !asset || !/^[a-f0-9]{64}$/i.test(asset.sha256)) return null;
	if (assetId.startsWith("toeic/")) {
		if (!isToeicMediaKey(assetId)) return null;
		return APP_ROUTES.MEDIA.DETAIL.replace("$", assetId);
	}
	try {
		const url = new URL(assetId);
		if (url.protocol !== "https:" && url.protocol !== "http:") return null;
	} catch {
		if (
			assetId.startsWith("/") ||
			assetId.includes("\\") ||
			assetId.split("/").includes("..")
		)
			return null;
	}
	// Both local and imported HTTP(S) identities resolve to verified MinIO
	// objects. Taking and review always use the same-origin media forward.
	return APP_ROUTES.MEDIA.DETAIL.replace(
		"$",
		`exams/toeic/${asset.sha256.toLowerCase()}`,
	);
}
