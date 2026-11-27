import type { IeltsPassage } from "@engflex/contracts";
import { Highlighter, Trash2 } from "lucide-react";
import { useRef } from "react";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";

export function IeltsPassagePane({
	passages,
	highlightMode = false,
}: {
	passages: IeltsPassage[];
	highlightMode?: boolean;
}) {
	const ref = useRef<HTMLDivElement>(null);
	const clear = () => {
		for (const mark of ref.current?.querySelectorAll("mark") ?? [])
			mark.replaceWith(...mark.childNodes);
	};
	return (
		<section
			className="flex h-full flex-col overflow-hidden"
			aria-label={m["exams.ielts.take.passageTitle"]()}
		>
			<div className="flex shrink-0 items-center justify-between border-b border-border px-4 py-2 text-sm font-semibold">
				<span className="flex items-center gap-2">
					<Highlighter className="size-4" />
					{m["exams.ielts.take.passageTitle"]()}
				</span>
				<Button size="sm" variant="ghost" onClick={clear}>
					<Trash2 className="size-4" />
					{m["exams.ielts.take.clearHighlight"]()}
				</Button>
			</div>
			<section
				ref={ref}
				aria-label={m["exams.ielts.take.highlightHint"]()}
				className="flex-1 space-y-6 overflow-y-auto p-4 md:p-6"
				onMouseUp={() => {
					if (!highlightMode) return;
					const selection = window.getSelection();
					if (!selection || selection.isCollapsed || !selection.rangeCount)
						return;
					const range = selection.getRangeAt(0);
					if (!ref.current?.contains(range.commonAncestorContainer)) return;
					try {
						const mark = document.createElement("mark");
						mark.className = "rounded bg-amber-200 text-amber-950";
						range.surroundContents(mark);
						selection.removeAllRanges();
					} catch {
						/* Selections crossing paragraph boundaries cannot be wrapped as one inline mark. */
					}
				}}
			>
				{passages.map((passage) => (
					<article
						key={passage.id}
						className="space-y-4 rounded-xl border border-border bg-card p-4 md:p-6"
					>
						{passage.title && (
							<h2 className="text-lg font-bold">{passage.title}</h2>
						)}
						{passage.content.map((paragraph, index) => (
							<p
								key={paragraph.id || `${passage.id}-${index}`}
								className="whitespace-pre-line text-sm leading-relaxed"
							>
								{paragraph.label && <strong>{paragraph.label} </strong>}
								{paragraph.text}
							</p>
						))}
					</article>
				))}
			</section>
		</section>
	);
}
