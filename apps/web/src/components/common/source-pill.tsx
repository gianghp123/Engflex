import type { VocabularySource } from "@engflex/contracts";
import { cn } from "cn";
import { BookOpen, Clapperboard, PenLine, Sparkles } from "lucide-react";
import { m } from "#/paraglide/messages";

const SOURCE = {
	lesson: { label: () => m["vocabulary.source.lesson"](), icon: BookOpen },
	conversation: {
		label: () => m["vocabulary.source.conversation"](),
		icon: Sparkles,
	},
	video: { label: () => m["vocabulary.source.video"](), icon: Clapperboard },
	manual: { label: () => m["vocabulary.source.manual"](), icon: PenLine },
} as const;

export function SourcePill({
	source,
	suffix,
	className,
}: {
	source: VocabularySource;
	suffix?: string;
	className?: string;
}) {
	const { label, icon: Icon } = SOURCE[source];
	const labelText = label();
	return (
		<span
			className={cn(
				"inline-flex shrink-0 items-center gap-1 rounded-md border border-border bg-card px-1.5 py-0.5 text-[11px] font-medium text-muted-foreground",
				className,
			)}
		>
			<Icon className="size-3" />
			{labelText}
			{suffix ? `: ${suffix}` : null}
		</span>
	);
}
