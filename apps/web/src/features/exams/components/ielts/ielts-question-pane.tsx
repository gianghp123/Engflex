import type { IeltsTakingQuestionGroup } from "@engflex/contracts";
import { Bookmark } from "lucide-react";
import { useEffect } from "react";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import { IeltsContext } from "#/features/exams/components/ielts/ielts-context";
import { m } from "#/paraglide/messages";

export function IeltsQuestionPane({
	groups,
	answers,
	bookmarks,
	activeQuestionId,
	onAnswerChange,
	onToggleBookmark,
	disabled = false,
}: {
	groups: IeltsTakingQuestionGroup[];
	answers: Record<string, string>;
	bookmarks: Set<string>;
	activeQuestionId: string | null;
	onAnswerChange: (id: string, value: string) => void;
	onToggleBookmark: (id: string) => void;
	disabled?: boolean;
}) {
	useEffect(() => {
		if (activeQuestionId)
			document
				.getElementById(activeQuestionId)
				?.scrollIntoView({ behavior: "smooth", block: "nearest" });
	}, [activeQuestionId]);
	return (
		<section
			className="h-full overflow-y-auto p-4 md:p-6"
			aria-label={m["exams.ielts.take.questionsTitle"]()}
		>
			<div className="space-y-6">
				{groups.map((group) => (
					<div
						key={group.id}
						className="space-y-5 rounded-xl border border-border bg-card p-4 shadow-xs md:p-6"
					>
						{group.title && <h2 className="font-semibold">{group.title}</h2>}
						{group.context?.length ? (
							<IeltsContext blocks={group.context} />
						) : (
							<p className="whitespace-pre-line text-sm font-semibold">
								{group.instructions}
							</p>
						)}
						{group.questions.map((question) => {
							const answer = answers[question.id] ?? "";
							const multiple = group.type === "multiple_choice_multiple";
							return (
								<div
									id={question.id}
									key={question.id}
									className={`scroll-mt-4 space-y-3 rounded-lg border p-4 ${activeQuestionId === question.id ? "border-primary ring-1 ring-primary" : "border-border"}`}
								>
									<div className="flex items-center justify-between gap-3">
										<label
											htmlFor={`answer-${question.id}`}
											className="text-sm font-bold"
										>
											{m["exams.ielts.take.question"]({
												number: question.questionNumber,
											})}
										</label>
										<Button
											size="sm"
											variant="ghost"
											disabled={disabled}
											onClick={() => onToggleBookmark(question.id)}
											aria-pressed={bookmarks.has(question.id)}
										>
											<Bookmark
												className={
													bookmarks.has(question.id)
														? "size-4 fill-amber-400 text-amber-500"
														: "size-4"
												}
											/>
											{bookmarks.has(question.id)
												? m["exams.ielts.take.bookmarked"]()
												: m["exams.ielts.take.bookmark"]()}
										</Button>
									</div>
									{question.stem && (
										<p className="text-sm leading-relaxed">{question.stem}</p>
									)}
									{question.options?.length ? (
										<fieldset
											id={`answer-${question.id}`}
											className="flex flex-wrap gap-2"
											aria-label={m["exams.ielts.take.question"]({
												number: question.questionNumber,
											})}
										>
											{question.options.map((option) => {
												const selected = multiple
													? answer.split(",").includes(option.key)
													: answer === option.key;
												return (
													<Button
														key={option.key}
														variant={selected ? "default" : "outline"}
														className="h-auto min-h-9 whitespace-normal text-left"
														disabled={disabled}
														aria-pressed={selected}
														onClick={() => {
															const values = answer ? answer.split(",") : [];
															onAnswerChange(
																question.id,
																multiple
																	? (selected
																			? values.filter(
																					(value) => value !== option.key,
																				)
																			: [...values, option.key]
																		).join(",")
																	: selected
																		? ""
																		: option.key,
															);
														}}
													>
														{option.key === option.text
															? option.text
															: `${option.key}. ${option.text}`}
													</Button>
												);
											})}
										</fieldset>
									) : (
										<Input
											id={`answer-${question.id}`}
											value={answer}
											disabled={disabled}
											onChange={(event) =>
												onAnswerChange(question.id, event.target.value)
											}
											autoComplete="off"
											placeholder={m["exams.ielts.take.inputPlaceholder"]()}
										/>
									)}
								</div>
							);
						})}
					</div>
				))}
			</div>
		</section>
	);
}
