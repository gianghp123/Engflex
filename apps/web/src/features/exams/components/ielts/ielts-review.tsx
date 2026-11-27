import type { ExamReviewAttempt } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { Award, CheckCircle2, Clock, Target, XCircle } from "lucide-react";
import { useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { BackButton } from "#/components/common/back-button";
import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { StatCard } from "#/components/common/stat-card";
import { Button } from "#/components/ui/button";
import { AiReviewPanel } from "#/features/exams/components/result/ai-review-panel";
import { answerValue, isCorrectAnswer } from "#/features/exams/ielts/review";
import { m } from "#/paraglide/messages";
import { IeltsAudio } from "./ielts-audio";
import { IeltsContext } from "./ielts-context";

export function IeltsReview({ attempt }: { attempt: ExamReviewAttempt }) {
	const [filter, setFilter] = useState<"all" | "correct" | "incorrect">("all");
	const content = attempt.examTest?.ieltsContent;
	const result = attempt.ieltsResult;
	const answers = new Map(
		(result?.answers ?? []).map((answer) => [answer.questionId, answer]),
	);
	const total =
		content?.sections.reduce(
			(sum, section) =>
				sum +
				section.questionGroups.reduce(
					(count, group) => count + group.questions.length,
					0,
				),
			0,
		) ?? 0;
	const correct = result?.correctCount ?? 0;
	const duration = attempt.durationSec ?? 0;
	return (
		<PageLayout
			action={
				<BackButton
					to={APP_ROUTES.EXAMS.IELTS.HISTORY}
					label={m["exams.ielts.history.title"]()}
				/>
			}
		>
			<div className="space-y-6">
				<div>
					<h1 className="text-2xl font-bold">
						{m["exams.ielts.result.title"]()}
					</h1>
					<p className="text-muted-foreground">{attempt.examTest?.title}</p>
				</div>
				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
					<StatCard
						title={m["exams.ielts.result.bandScore"]()}
						value={(result?.bandScore ?? attempt.score ?? 0).toFixed(1)}
						aside={<Award className="size-7 text-amber-500" />}
					/>
					<StatCard
						title={m["exams.ielts.result.metrics.correct"]()}
						value={`${correct}/${total}`}
						aside={<CheckCircle2 className="size-7 text-emerald-500" />}
					/>
					<StatCard
						title={m["exams.ielts.result.metrics.accuracy"]()}
						value={`${Math.round(result?.accuracyPct ?? 0)}%`}
						aside={<Target className="size-7 text-sky-500" />}
					/>
					<StatCard
						title={m["exams.ielts.result.metrics.time"]()}
						value={`${Math.floor(duration / 60)}:${String(duration % 60).padStart(2, "0")}`}
						aside={<Clock className="size-7 text-purple-500" />}
					/>
				</div>
				<AiReviewPanel
					attemptId={attempt.id}
					examReviewCost={attempt.aiReviewCreditCost}
					studyPlanCost={attempt.studyPlanCreditCost}
				/>
				<div className="flex flex-wrap gap-2">
					<Button
						variant={filter === "all" ? "default" : "outline"}
						onClick={() => setFilter("all")}
					>
						{m["exams.ielts.result.filterAll"]()} ({total})
					</Button>
					<Button
						variant={filter === "correct" ? "default" : "outline"}
						onClick={() => setFilter("correct")}
					>
						{m["exams.ielts.result.filterCorrect"]()} ({correct})
					</Button>
					<Button
						variant={filter === "incorrect" ? "default" : "outline"}
						onClick={() => setFilter("incorrect")}
					>
						{m["exams.ielts.result.filterIncorrect"]()} ({total - correct})
					</Button>
				</div>
				{content?.sections.map((section) => (
					<CardShell
						key={section.sectionNumber}
						className="space-y-5 p-4 md:p-6"
					>
						<h2 className="text-lg font-bold">
							{section.title ||
								m["exams.ielts.take.section"]({
									number: section.sectionNumber,
								})}
						</h2>
						{section.audioAssetId && (
							<IeltsAudio identity={section.audioAssetId} />
						)}
						{section.passages.map((passage) => (
							<details key={passage.id} className="rounded-lg border p-3">
								<summary className="cursor-pointer font-semibold">
									{passage.title || m["exams.ielts.take.passageTitle"]()}
								</summary>
								<div className="mt-3 space-y-3">
									{passage.content.map((paragraph, index) => (
										<p
											className="whitespace-pre-line text-sm leading-relaxed"
											key={paragraph.id || index}
										>
											{paragraph.label && <strong>{paragraph.label} </strong>}
											{paragraph.text}
										</p>
									))}
								</div>
							</details>
						))}
						{Boolean(section.transcript?.length) && (
							<details className="rounded-lg border p-3">
								<summary className="cursor-pointer font-semibold">
									{m["exams.ielts.media.transcript"]()}
								</summary>
								<div className="mt-3 space-y-2">
									{section.transcript?.map((line, index) => (
										<p
											key={line.id || index}
											className="whitespace-pre-line text-sm leading-relaxed"
										>
											{line.text}
										</p>
									))}
								</div>
							</details>
						)}
						{section.questionGroups.map((group) => {
							const questions = group.questions.filter(
								(question) =>
									filter === "all" ||
									isCorrectAnswer(question, answers.get(question.id)) ===
										(filter === "correct"),
							);
							if (!questions.length) return null;
							return (
								<div
									key={group.id}
									className="space-y-4 rounded-lg border border-border p-4"
								>
									{group.context?.length ? (
										<IeltsContext blocks={group.context} />
									) : (
										<p className="text-sm font-semibold">
											{group.instructions}
										</p>
									)}
									{questions.map((question) => {
										const answer = answers.get(question.id);
										const passed = isCorrectAnswer(question, answer);
										return (
											<div
												key={question.id}
												className="space-y-2 border-t pt-4"
											>
												<div className="flex items-center gap-2">
													{passed ? (
														<CheckCircle2 className="size-5 text-emerald-500" />
													) : (
														<XCircle className="size-5 text-destructive" />
													)}
													<h3 className="text-sm font-bold">
														{m["exams.ielts.take.question"]({
															number: question.questionNumber,
														})}
													</h3>
												</div>
												{question.stem && (
													<p className="text-sm">{question.stem}</p>
												)}
												{Boolean(question.options?.length) && (
													<ul className="space-y-1 text-sm">
														{question.options?.map((option) => (
															<li key={option.key}>
																{option.key === option.text
																	? option.text
																	: `${option.key}. ${option.text}`}
															</li>
														))}
													</ul>
												)}
												<p className="text-sm">
													{m["exams.ielts.result.yourAnswer"]({
														answer:
															answerValue(answer) ||
															m["exams.ielts.result.unanswered"](),
													})}
												</p>
												<p className="text-sm text-emerald-700 dark:text-emerald-400">
													{m["exams.ielts.result.correctAnswer"]({
														answer:
															question.sourceAnswer ||
															question.correctKey ||
															question.correctAnswers?.join(" / ") ||
															question.correctKeys?.join(", ") ||
															"",
													})}
												</p>
												{(question.vietnameseExplanation ||
													question.explanation) && (
													<div className="whitespace-pre-line rounded-lg bg-muted/50 p-3 text-sm leading-relaxed">
														<strong>
															{m["exams.ielts.result.explanation"]()}{" "}
														</strong>
														{question.vietnameseExplanation ||
															question.explanation}
													</div>
												)}
											</div>
										);
									})}
								</div>
							);
						})}
					</CardShell>
				))}
				{attempt.examTest && (
					<Button asChild>
						<Link
							to={APP_ROUTES.EXAMS.IELTS.TEST}
							params={{
								categoryId: attempt.examTest.categoryId,
								testId: attempt.examTest.id,
							}}
						>
							{m["exams.ielts.result.retakeButton"]()}
						</Link>
					</Button>
				)}
			</div>
		</PageLayout>
	);
}
