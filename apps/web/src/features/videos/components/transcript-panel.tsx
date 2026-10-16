import type { VideoTranscriptResponse } from "@engflex/contracts";
import { cn } from "cn";
import { SquarePen } from "lucide-react";
import type { ReactNode } from "react";
import { Button } from "#/components/ui/button";
import { ScrollArea } from "#/components/ui/scroll-area";
import { maskSentence } from "#/features/videos/dictation";
import { m } from "#/paraglide/messages";

/**
 * Right pane shared by both practice modes: transcript list + overall
 * progress. Dictation masks each line; shadowing shows it with optional
 * IPA and translation lines.
 */
export function TranscriptPanel({
	transcripts,
	activeIndex,
	progress,
	masked = true,
	showPhonetic = false,
	showTranslation = false,
	actions,
	onSelect,
	onEdit,
}: {
	transcripts: VideoTranscriptResponse[];
	activeIndex: number;
	progress: number;
	masked?: boolean;
	showPhonetic?: boolean;
	showTranslation?: boolean;
	actions?: ReactNode;
	onSelect?: (index: number) => void;
	onEdit?: (index: number) => void;
}) {
	const value = Math.round(progress);
	return (
		<section className="surface-card flex flex-col gap-3 p-3">
			<div className="flex items-center justify-between gap-2 px-1">
				<h2 className="text-base font-extrabold tracking-wide text-foreground uppercase">
					{m["videos.practice.transcript.title"]()}
				</h2>
				{actions ? (
					<div className="flex flex-1 items-center gap-1">{actions}</div>
				) : null}
				<span className="chip px-3 py-1">
					{m["videos.practice.transcript.progress"]({ value })}
				</span>
			</div>
			<div
				className="meter"
				role="progressbar"
				aria-valuemin={0}
				aria-valuemax={100}
				aria-valuenow={value}
				aria-label={m["videos.practice.transcript.title"]()}
			>
				<div
					className="h-full rounded-full bg-primary transition-[width]"
					style={{ width: `${value}%` }}
				/>
			</div>

			<ScrollArea className="h-[560px] pr-3">
				<ol className="flex flex-col gap-2 pb-2">
					{transcripts.map((line, index) => {
						const active = index === activeIndex;
						const itemLabel = m["videos.practice.transcript.item"]({
							index: line.sequence,
						});
						return (
							<li
								key={line.id}
								className={cn(
									"flex flex-col gap-2 rounded-2xl border-2 bg-card px-3 py-2 transition-colors",
									active
										? "border-primary bg-accent/40 shadow-[0_4px_0_var(--primary)]"
										: "border-border shadow-[0_3px_0_var(--border)]",
								)}
							>
								<div className="flex items-center justify-between gap-2">
									<span
										className={cn(
											"chip rounded-lg px-2 py-0.5 text-xs",
											active && "chip-selected",
										)}
									>
										#{line.sequence}
									</span>
									<div className="flex items-center">
										<Button
											type="button"
											variant="ghost"
											size="icon"
											className="size-8"
											aria-label={`${m["videos.practice.transcript.edit"]()} — ${itemLabel}`}
											title={m["videos.practice.transcript.edit"]()}
											onClick={() => onEdit?.(index)}
										>
											<SquarePen />
										</Button>
									</div>
								</div>
								<button
									type="button"
									aria-current={active ? "step" : undefined}
									aria-label={masked ? itemLabel : undefined}
									onClick={() => onSelect?.(index)}
									className="flex cursor-pointer flex-col gap-1 rounded-lg text-left focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-ring"
								>
									{masked ? (
										<span className="font-mono text-sm leading-6 tracking-widest break-words text-muted-foreground">
											{maskSentence(line.content)}
										</span>
									) : (
										<>
											<span className="text-[15px] leading-6 font-medium text-foreground">
												{line.content}
											</span>
											{showPhonetic && line.phonetic ? (
												<span className="font-mono text-xs leading-5 text-muted-foreground">
													/{line.phonetic}/
												</span>
											) : null}
											{showTranslation && line.vietnamese ? (
												<span className="text-sm leading-5 text-muted-foreground italic">
													{line.vietnamese}
												</span>
											) : null}
										</>
									)}
								</button>
							</li>
						);
					})}
				</ol>
			</ScrollArea>
		</section>
	);
}
