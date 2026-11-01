import type { ExamReviewContent } from "@engflex/contracts";
import { Bookmark } from "lucide-react";
import { CardShell } from "#/components/common/card-shell";
import { Badge } from "#/components/ui/badge";
import {
	type AnswersById,
	matchesReviewFilter,
	mediaSource,
	questionStatus,
	type ReviewFilter,
	type ReviewItem,
	type ReviewPart,
} from "#/features/exams/review";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
import { ToeicReviewAudio } from "./toeic-review-audio";
import { ToeicReviewBlocks } from "./toeic-review-blocks";

function ReviewItemCard({
	item,
	assets,
	answers,
	filter,
	focusedQuestion,
}: {
	item: ReviewItem;
	assets: ExamReviewContent["assets"];
	answers: AnswersById;
	filter: ReviewFilter;
	focusedQuestion: string | null;
}) {
	const questions = item.questions.filter((question) =>
		matchesReviewFilter(question, answers, filter),
	);
	const numbers = new Map(
		item.questions.map((question) => [question.id, question.questionNumber]),
	);
	if (!questions.length) return null;
	const audio = item.itemAudio ?? item.stimulusAudio;
	return (
		<div className="flex flex-col gap-5 rounded-xl border p-5">
			<ToeicReviewBlocks
				blocks={item.media}
				assets={assets}
				numbers={numbers}
			/>
			{audio && (
				<ToeicReviewAudio
					key={`audio-${item.id}`}
					range={audio}
					src={mediaSource(audio.assetId, assets)}
					label={m["exams.toeic.result.review.itemAudio"]()}
				/>
			)}
			{item.passages.map((passage) => (
				<div key={passage.id} className="rounded-xl border bg-muted/20 p-4">
					<ToeicReviewBlocks
						blocks={passage.blocks}
						assets={assets}
						numbers={numbers}
					/>
				</div>
			))}
			{item.transcript.length > 0 && (
				<details className="rounded-lg bg-muted/40 p-3">
					<summary className="cursor-pointer text-sm font-semibold">
						{m["exams.toeic.result.review.transcript"]()}
					</summary>
					<div className="mt-3 flex flex-col gap-2">
						{item.transcript.map((line) => (
							<div key={line.id} className="text-sm">
								<p>
									{line.speaker ? `${line.speaker}: ` : ""}
									{line.optionKey ? `${line.optionKey}. ` : ""}
									{line.text}
								</p>
								{getLocale() === "vi" && line.vietnameseText && (
									<p className="text-xs text-muted-foreground">
										{line.vietnameseText}
									</p>
								)}
							</div>
						))}
					</div>
				</details>
			)}
			{questions.map((question) => {
				const status = questionStatus(question, answers);
				const selected = answers.get(question.id)?.selectedKey;
				return (
					<section
						key={question.id}
						id={question.id}
						className="flex scroll-mt-24 flex-col gap-3 rounded-xl border bg-background p-4"
						ref={(node) => {
							if (node && question.id === focusedQuestion)
								node.scrollIntoView({ behavior: "smooth", block: "start" });
						}}
					>
						<div className="flex flex-wrap items-center gap-3">
							<h4 className="font-bold">
								{m["exams.toeic.result.review.question"]({
									number: question.questionNumber,
								})}
							</h4>
							{answers.get(question.id)?.flagged && (
								<Badge variant="outline">
									<Bookmark
										className="size-3 fill-amber-400"
										aria-hidden="true"
									/>
									{m["exams.toeic.result.flagged"]()}
								</Badge>
							)}
							<Badge variant="outline">
								{status === "correct"
									? m["exams.toeic.result.correctBadge"]()
									: status === "skipped"
										? m["exams.toeic.result.skippedBadge"]()
										: m["exams.toeic.result.incorrectBadge"]()}
							</Badge>
						</div>
						{question.stem && <p className="text-sm">{question.stem}</p>}
						{question.targetIds.length > 0 && (
							<p className="text-xs text-muted-foreground">
								{m["exams.toeic.result.review.reference"]({
									number: question.questionNumber,
								})}
							</p>
						)}
						{getLocale() === "vi" && question.vietnameseText && (
							<p className="text-sm text-muted-foreground">
								{question.vietnameseText}
							</p>
						)}
						<div className="flex flex-col gap-2">
							{question.options.map((option) => {
								const transcript = item.transcript.find(
									(line) =>
										line.role === "option" && line.optionKey === option.key,
								);
								return (
									<div
										key={option.key}
										data-correct={option.key === question.correctKey}
										data-selected={option.key === selected}
										className={`rounded-lg border p-3 text-sm ${option.key === question.correctKey ? "border-emerald-500 bg-emerald-500/10" : option.key === selected ? "border-rose-500 bg-rose-500/10" : "border-border"}`}
									>
										<span className="mr-2 font-bold">{option.key}.</span>
										{option.text ?? transcript?.text}
										{getLocale() === "vi" && transcript?.vietnameseText && (
											<p className="mt-1 text-xs text-muted-foreground">
												{transcript.vietnameseText}
											</p>
										)}
									</div>
								);
							})}
						</div>
						{question.explanation && (
							<p className="rounded-lg bg-primary/5 p-3 text-sm">
								<strong>{m["exams.toeic.result.explanationLabel"]()} </strong>
								{question.explanation}
							</p>
						)}
						{question.vocabulary.length > 0 && (
							<div className="text-sm">
								<h5 className="font-semibold">
									{m["exams.toeic.result.review.vocabulary"]()}
								</h5>
								<ul className="list-inside list-disc">
									{question.vocabulary.map((entry, index) => (
										<li key={JSON.stringify([index, entry.word, entry.raw])}>
											{entry.word}
											{entry.pos ? ` (${entry.pos})` : ""}
											{getLocale() === "vi" && entry.meaning
												? ` — ${entry.meaning}`
												: ""}
											{!entry.word && entry.raw}
										</li>
									))}
								</ul>
							</div>
						)}
					</section>
				);
			})}
		</div>
	);
}
export function ToeicResultReviewSection({
	parts,
	assets,
	answers,
	activePart,
	onSelectPart,
	selectedFilter,
	onSelectFilter,
	focusedQuestion,
}: {
	parts: ReviewPart[];
	assets: ExamReviewContent["assets"];
	answers: AnswersById;
	activePart: number;
	onSelectPart: (part: number) => void;
	selectedFilter: ReviewFilter;
	onSelectFilter: (filter: ReviewFilter) => void;
	focusedQuestion: string | null;
}) {
	const part = parts.find((entry) => entry.partNumber === activePart);
	const hasQuestions = part?.items.some((item) =>
		item.questions.some((question) =>
			matchesReviewFilter(question, answers, selectedFilter),
		),
	);
	return (
		<CardShell className="flex flex-col gap-5 p-6">
			<h2 className="font-bold">{m["exams.toeic.result.reviewTitle"]()}</h2>
			<div className="flex flex-wrap gap-2">
				<button
					type="button"
					aria-pressed={selectedFilter === "all"}
					className="rounded-lg border px-3 py-2 text-sm aria-pressed:bg-primary/10"
					onClick={() => onSelectFilter("all")}
				>
					{m["exams.toeic.result.filterAll"]()}
				</button>
				<button
					type="button"
					aria-pressed={selectedFilter === "incorrect"}
					className="rounded-lg border px-3 py-2 text-sm aria-pressed:bg-primary/10"
					onClick={() => onSelectFilter("incorrect")}
				>
					{m["exams.toeic.result.filterIncorrect"]()}
				</button>
				<button
					type="button"
					aria-pressed={selectedFilter === "correct"}
					className="rounded-lg border px-3 py-2 text-sm aria-pressed:bg-primary/10"
					onClick={() => onSelectFilter("correct")}
				>
					{m["exams.toeic.result.filterCorrect"]()}
				</button>
				<button
					type="button"
					aria-pressed={selectedFilter === "skipped"}
					className="rounded-lg border px-3 py-2 text-sm aria-pressed:bg-primary/10"
					onClick={() => onSelectFilter("skipped")}
				>
					{m["exams.toeic.result.filterSkipped"]()}
				</button>
			</div>
			<div className="flex flex-wrap gap-2 border-b pb-4">
				{parts.map((entry) => (
					<button
						key={entry.partNumber}
						type="button"
						aria-pressed={entry.partNumber === activePart}
						className="rounded-lg border px-3 py-2 text-sm aria-pressed:bg-primary/10"
						onClick={() => onSelectPart(entry.partNumber)}
					>
						{m["exams.toeic.result.review.part"]({ part: entry.partNumber })}
					</button>
				))}
			</div>
			{!hasQuestions ? (
				<p className="py-10 text-center text-sm text-muted-foreground">
					{m["exams.toeic.result.review.empty"]()}
				</p>
			) : (
				part?.items.map((item) => (
					<ReviewItemCard
						key={item.id}
						item={item}
						assets={assets}
						answers={answers}
						filter={selectedFilter}
						focusedQuestion={focusedQuestion}
					/>
				))
			)}
		</CardShell>
	);
}
