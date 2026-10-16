import { cn } from "cn";
import { Eye, EyeOff } from "lucide-react";
import { Button } from "#/components/ui/button";
import { maskWord } from "#/features/videos/dictation";
import { m } from "#/paraglide/messages";

/** Masked word slots; each eye button reveals one word. */
export function WordHints({
	words,
	revealed,
	onReveal,
}: {
	words: string[];
	revealed: number[];
	onReveal?: (index: number) => void;
}) {
	return (
		<ul className="flex flex-wrap gap-2">
			{words.map((word, index) => {
				const shown = revealed.includes(index);
				const Icon = shown ? EyeOff : Eye;
				const label = shown
					? m["videos.dictation.hints.hide"]({ index: index + 1 })
					: m["videos.dictation.hints.reveal"]({ index: index + 1 });
				return (
					// biome-ignore lint/suspicious/noArrayIndexKey: words repeat within a sentence; position is the identity.
					<li key={index} className="flex flex-col items-center gap-1">
						<Button
							type="button"
							variant="ghost"
							size="icon"
							aria-label={label}
							title={label}
							aria-pressed={shown}
							onClick={() => onReveal?.(index)}
							className="size-7 text-muted-foreground"
						>
							<Icon />
						</Button>
						<span
							className={cn(
								"chip rounded-xl px-2.5 py-1 font-mono tracking-widest",
								shown && "chip-accent tracking-normal",
							)}
						>
							{shown ? word : maskWord(word)}
						</span>
					</li>
				);
			})}
		</ul>
	);
}
