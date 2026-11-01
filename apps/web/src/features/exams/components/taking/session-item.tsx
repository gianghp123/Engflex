import type { ExamAnswer, ExamTakingContent } from "@engflex/contracts";
import { Bookmark } from "lucide-react";
import { Button } from "#/components/ui/button";
import { mediaSource } from "#/features/exams/review";
import { m } from "#/paraglide/messages";
import { ToeicContentBlocks } from "../toeic-content-blocks";
import { SessionAudio } from "./session-audio";

type Item = ExamTakingContent["parts"][number]["items"][number];
export function SessionItem({
	item,
	assets,
	answers,
	speed,
	onEdit,
	onPlay,
	locked,
}: {
	item: Item;
	assets: ExamTakingContent["assets"];
	answers: ReadonlyMap<string, ExamAnswer>;
	speed: number;
	onEdit: (answer: ExamAnswer) => void;
	onPlay: (audio: HTMLAudioElement) => void;
	locked: boolean;
}) {
	const numbers = new Map(
		item.questions.map((question) => [question.id, question.questionNumber]),
	);
	const audio = item.itemAudio ?? item.stimulusAudio;
	const copy = {
		imageAlt: m["exams.toeic.take.imageAlt"](),
		mediaFailed: m["exams.toeic.take.mediaFailed"](),
		blank: (number: number) => m["exams.toeic.take.blank"]({ number }),
		insertion: m["exams.toeic.take.insertion"](),
	};
	return (
		<div className="flex flex-col gap-5 rounded-2xl border bg-card p-4 md:p-6">
			{audio && (
				<SessionAudio
					key={item.id}
					range={audio}
					src={mediaSource(audio.assetId, assets)}
					speed={speed}
					locked={locked}
					onPlay={onPlay}
				/>
			)}
			<ToeicContentBlocks
				blocks={item.media}
				assets={assets}
				numbers={numbers}
				copy={copy}
			/>
			{item.passages.map((passage) => (
				<div key={passage.id} className="rounded-xl border bg-muted/20 p-4">
					<ToeicContentBlocks
						blocks={passage.blocks}
						assets={assets}
						numbers={numbers}
						copy={copy}
					/>
				</div>
			))}
			{item.questions.map((question) => {
				const answer = answers.get(question.id) ?? {
					questionId: question.id,
					selectedKey: null,
					flagged: false,
				};
				return (
					<section
						key={question.id}
						id={question.id}
						className="flex scroll-mt-4 flex-col gap-3 rounded-xl border bg-background p-4"
					>
						<div className="flex items-center justify-between gap-2">
							<h2 className="font-bold">
								{m["exams.toeic.take.question"]({
									number: question.questionNumber,
								})}
							</h2>
							<Button
								variant="ghost"
								size="sm"
								aria-pressed={answer.flagged}
								onClick={() => onEdit({ ...answer, flagged: !answer.flagged })}
							>
								<Bookmark
									className={
										answer.flagged ? "fill-amber-400 text-amber-500" : ""
									}
								/>
								{answer.flagged
									? m["exams.toeic.take.bookmarked"]()
									: m["exams.toeic.take.bookmark"]()}
							</Button>
						</div>
						{question.stem && (
							<p className="whitespace-pre-wrap text-sm leading-relaxed">
								{question.stem}
							</p>
						)}
						{question.targetIds.length > 0 && (
							<div className="flex flex-wrap gap-2">
								{question.targetIds.map((target, index) => (
									<Button
										key={target}
										variant="link"
										size="sm"
										onClick={() =>
											document.getElementById(target)?.scrollIntoView({
												behavior: "smooth",
												block: "center",
											})
										}
									>
										{m["exams.toeic.take.reference"]({ number: index + 1 })}
									</Button>
								))}
							</div>
						)}
						<div className="flex flex-col gap-2">
							{question.options.map((option) => (
								<button
									key={option.key}
									type="button"
									aria-pressed={answer.selectedKey === option.key}
									className="rounded-xl border p-3 text-left text-sm hover:bg-muted aria-pressed:border-primary aria-pressed:bg-primary/10"
									onClick={() => onEdit({ ...answer, selectedKey: option.key })}
								>
									<span className="mr-3 font-bold">{option.key}.</span>
									{option.text}
								</button>
							))}
						</div>
						{answer.selectedKey !== null && (
							<Button
								variant="ghost"
								size="sm"
								className="self-start"
								onClick={() => onEdit({ ...answer, selectedKey: null })}
							>
								{m["exams.toeic.take.clear"]()}
							</Button>
						)}
					</section>
				);
			})}
		</div>
	);
}
