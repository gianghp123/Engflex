import type { VocabularyItem } from "@engflex/contracts";
import { cn } from "cn";
import { Check, Copy } from "lucide-react";
import { toast } from "sonner";
import { CefrBadge } from "#/components/common/cefr-badge";
import { SourcePill } from "#/components/common/source-pill";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { posLabel } from "#/features/vocabulary/components/pos-label";
import { SourceRef } from "#/features/vocabulary/components/source-ref";
import { usePatchVocabulary } from "#/features/vocabulary/queries";
import { m } from "#/paraglide/messages";

export function WordHeroCard({ item }: { item: VocabularyItem }) {
	const patch = usePatchVocabulary();
	const state = item.userState;
	const mastered = state?.mastered ?? false;

	function toggleMastered() {
		patch.mutate(
			{
				itemId: item.id,
				patch: { mastered: !(item.userState?.mastered ?? false) },
			},
			{
				onError: () => toast.error(m["vocabulary.card.masteredFailed"]()),
			},
		);
	}

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
						{posLabel(item.partOfSpeech)}
					</span>
				) : null}
				{state ? <SourcePill source={state.sourceType} /> : null}
				{state ? (
					<SourceRef
						type={state.sourceType}
						id={state.sourceId}
						showLabel={false}
					/>
				) : null}
			</div>

			{item.details?.ipa ? (
				<div className="flex flex-wrap items-center gap-2">
					<span className="chip font-mono text-xs">{item.details.ipa}</span>
				</div>
			) : null}

			<p className="text-[15px] font-medium text-muted-foreground">
				{item.definition}
			</p>

			<div className="flex flex-wrap items-center gap-2 border-t-2 border-border pt-4">
				<SubmitButton
					type="button"
					variant={mastered ? "outline" : "default"}
					pending={patch.isPending}
					onClick={toggleMastered}
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
				</SubmitButton>
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
			</div>
		</div>
	);
}
