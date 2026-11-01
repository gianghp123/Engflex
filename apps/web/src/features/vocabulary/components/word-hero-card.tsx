import type { VocabularyItem } from "@engflex/contracts";
import { cn } from "cn";
import { Check, Copy, Share2 } from "lucide-react";
import { toast } from "sonner";
import { AudioButton } from "#/components/common/audio-button";
import { CefrBadge } from "#/components/common/cefr-badge";
import { SourcePill } from "#/components/common/source-pill";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";

export function WordHeroCard({
	item,
	onToggleMastered,
}: {
	item: VocabularyItem;
	onToggleMastered: () => void;
}) {
	const state = item.userState;
	const mastered = state?.mastered ?? false;

	async function copy(value: string, message: string) {
		try {
			await navigator.clipboard.writeText(value);
			toast.success(message);
		} catch {
			toast.error(m["common.toast.clipboardUnavailable"]());
		}
	}

	return (
		<div className="surface-hero flex flex-col gap-4 p-5">
			<div className="flex flex-wrap items-center gap-2">
				<h1 className="text-2xl font-bold text-foreground">{item.term}</h1>
				<CefrBadge value={item.cefr} />
				{item.partOfSpeech ? (
					<span className="chip px-1.5 py-0.5 text-[11px]">
						{item.partOfSpeech}
					</span>
				) : null}
				{state ? (
					<SourcePill source={state.sourceType} suffix={state.sourceId} />
				) : null}
			</div>

			<div className="flex flex-wrap items-center gap-2">
				<span className="chip font-mono text-xs">{item.ipa}</span>
				<AudioButton label={m["vocabulary.card.listenUs"]()} />
				<AudioButton label={m["vocabulary.card.listenUk"]()} />
			</div>

			<p className="text-[15px] font-medium text-muted-foreground">
				{item.definition}
			</p>

			<div className="flex flex-wrap items-center gap-2 border-t-2 border-border pt-4">
				<Button
					type="button"
					variant={mastered ? "outline" : "default"}
					onClick={onToggleMastered}
					className={cn(
						"btn",
						mastered ? "btn-outline" : "btn-primary",
						mastered && "text-accuracy",
					)}
				>
					<Check data-icon="inline-start" />
					{mastered
						? m["vocabulary.card.mastered"]()
						: m["vocabulary.card.markAsMastered"]()}
				</Button>
				<Button
					type="button"
					variant="ghost"
					onClick={() =>
						copy(
							`${item.term} — ${item.definition}`,
							m["common.toast.copiedClipboard"](),
						)
					}
				>
					<Copy data-icon="inline-start" />
					{m["common.actions.copy"]()}
				</Button>
				<Button
					type="button"
					variant="ghost"
					onClick={() =>
						copy(window.location.href, m["common.toast.linkCopied"]())
					}
				>
					<Share2 data-icon="inline-start" />
					{m["vocabulary.card.share"]()}
				</Button>
			</div>
		</div>
	);
}
