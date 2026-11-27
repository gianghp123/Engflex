import type {
	ExamReviewAttempt,
	IeltsReviewQuestion,
} from "@engflex/contracts";

type Answer = NonNullable<ExamReviewAttempt["ieltsResult"]>["answers"][number];

export function answerValue(answer: Answer | undefined): string {
	return (
		answer?.selectedKey ??
		answer?.textValue ??
		answer?.selectedKeys?.join(", ") ??
		""
	);
}

function normalize(value: string): string {
	return value
		.trim()
		.toLowerCase()
		.replace(/\s+/g, " ")
		.replace(/[.,;!?:]+$/, "")
		.trim();
}

export function isCorrectAnswer(
	question: IeltsReviewQuestion,
	answer: Answer | undefined,
): boolean {
	if (!answer) return false;
	if (
		question.correctKey &&
		answer.selectedKey &&
		question.correctKey.trim().toLowerCase() ===
			answer.selectedKey.trim().toLowerCase()
	)
		return true;
	if (question.correctKeys?.length && answer.selectedKeys?.length) {
		const expected = [...question.correctKeys]
			.map((value) => value.trim().toLowerCase())
			.sort();
		const actual = [...answer.selectedKeys]
			.map((value) => value.trim().toLowerCase())
			.sort();
		if (
			expected.length === actual.length &&
			expected.every((value, index) => value === actual[index])
		)
			return true;
	}
	const value = normalize(answer.textValue ?? "");
	return Boolean(
		value &&
			question.correctAnswers?.some(
				(expected) => normalize(expected) === value,
			),
	);
}
