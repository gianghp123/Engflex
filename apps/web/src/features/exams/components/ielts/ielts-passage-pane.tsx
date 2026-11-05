import { Highlighter, Trash2 } from "lucide-react";
import { useRef } from "react";
import { Button } from "#/components/ui/button";
import type { IeltsPassage } from "#/features/exams/ielts/mock-data";
import { m } from "#/paraglide/messages";

export function IeltsPassagePane({
	passages,
	highlightMode,
	onClearHighlights,
}: {
	passages: IeltsPassage[];
	highlightMode: boolean;
	onClearHighlights: () => void;
}) {
	const containerRef = useRef<HTMLDivElement>(null);

	const handleMouseUp = () => {
		if (!highlightMode) return;
		const selection = window.getSelection();
		if (!selection || selection.isCollapsed || selection.rangeCount === 0)
			return;

		const range = selection.getRangeAt(0);
		if (!containerRef.current?.contains(range.commonAncestorContainer)) return;

		try {
			const span = document.createElement("mark");
			span.className =
				"bg-amber-200/90 text-amber-950 dark:bg-amber-400/40 dark:text-amber-100 rounded px-1 py-0.5 transition-colors";
			range.surroundContents(span);
			selection.removeAllRanges();
		} catch {
			// surroundContents fails if range crosses tag boundaries; safe no-op
		}
	};

	return (
		<div className="flex h-full flex-col overflow-hidden border-r border-border/80 bg-background">
			{/* Header */}
			<div className="flex h-10 shrink-0 items-center justify-between border-b border-border/70 px-4 bg-muted/30">
				<div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-muted-foreground">
					<Highlighter className="size-3.5 text-amber-500" />
					<span>{m["exams.ielts.take.passageTitle"]()}</span>
				</div>

				{highlightMode && (
					<div className="flex items-center gap-2">
						<span className="hidden lg:inline text-[11px] text-muted-foreground italic">
							{m["exams.ielts.take.highlightHint"]()}
						</span>
						<Button
							variant="ghost"
							size="sm"
							onClick={onClearHighlights}
							className="h-7 text-xs text-rose-600 hover:text-rose-700 hover:bg-rose-50 dark:hover:bg-rose-950/40 gap-1"
							title={m["exams.ielts.take.clearHighlight"]()}
						>
							<Trash2 className="size-3" />
							<span>{m["exams.ielts.take.clearHighlight"]()}</span>
						</Button>
					</div>
				)}
			</div>

			{/* Passage Content */}
			<section
				ref={containerRef}
				aria-label={m["exams.ielts.take.passageTitle"]()}
				onMouseUp={handleMouseUp}
				className="flex-1 overflow-y-auto p-4 md:p-6 space-y-8 select-text"
			>
				{passages.map((passage) => (
					<article
						key={passage.id}
						id={`passage-${passage.id}`}
						className="space-y-4 rounded-xl border border-border/60 bg-card p-5 md:p-6 shadow-xs"
					>
						<div>
							<h2 className="text-lg md:text-xl font-bold tracking-tight text-foreground">
								{passage.title}
							</h2>
							{passage.subtitle && (
								<p className="mt-1 text-sm italic text-muted-foreground">
									{passage.subtitle}
								</p>
							)}
						</div>

						<div className="space-y-3.5 text-sm md:text-[15px] leading-relaxed text-foreground/90 font-normal">
							{passage.paragraphs.map((p) => (
								<p
									key={`${passage.id}-${p.slice(0, 24)}`}
									className="whitespace-pre-line text-justify"
								>
									{p}
								</p>
							))}
						</div>
					</article>
				))}
			</section>
		</div>
	);
}
