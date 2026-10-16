import type { VocabularyItem } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { cn } from "cn";
import { Check, CircleCheck, X } from "lucide-react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { CefrBadge } from "#/components/common/cefr-badge";
import { SourcePill } from "#/components/common/source-pill";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { posLabel } from "#/features/vocabulary/components/pos-label";
import { SourceRef } from "#/features/vocabulary/components/source-ref";
import { isDue } from "#/features/vocabulary/dates";
import { usePatchVocabulary } from "#/features/vocabulary/queries";
import { m } from "#/paraglide/messages";

export function VocabCard({
	item,
	deckId,
	onRemove,
	removePending,
}: {
	item: VocabularyItem;
	/** Set inside a deck: the card links to that deck's word URL. */
	deckId?: string;
	/** Present only on a deck page: removes the membership, not the word. */
	onRemove?: () => void;
	/** The removal mutation's pending state, so the control shows progress. */
	removePending?: boolean;
}) {
	const patch = usePatchVocabulary();
	const state = item.userState;
	const mastered = state?.mastered ?? false;
	const due = isDue(state?.srsDueAt);
	const context = item.details?.contexts?.[0];

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

	return (
		<article className="surface-card flex flex-col gap-3 p-5">
			<div className="flex items-start justify-between gap-2">
				<div className="min-w-0">
					<h3 className="text-base font-bold text-foreground">{item.term}</h3>
					<p className="font-mono text-xs text-muted-foreground">
						{item.details?.ipa}
					</p>
				</div>
				{mastered ? (
					<span className="chip shrink-0 border-cefr-b1-border bg-cefr-b1-bg px-2 py-0.5 text-xs text-cefr-b1">
						<Check className="size-3.5" />
						{m["vocabulary.card.mastered"]()}
					</span>
				) : due ? (
					<SubmitButton
						type="button"
						variant="ghost"
						size="sm"
						pending={patch.isPending}
						onClick={toggleMastered}
						className="chip chip-accent shrink-0 px-2 py-0.5 text-xs"
					>
						<CircleCheck className="size-3.5" />
						{m["vocabulary.card.markMastered"]()}
					</SubmitButton>
				) : (
					<SubmitButton
						type="button"
						variant="ghost"
						size="sm"
						pending={patch.isPending}
						onClick={toggleMastered}
						className="chip shrink-0 px-2 py-0.5 text-xs hover:border-primary"
					>
						<Check className="size-3.5" />
						{m["vocabulary.card.markMastered"]()}
					</SubmitButton>
				)}
			</div>

			<div className="flex flex-wrap items-center gap-1.5">
				{item.partOfSpeech ? (
					<span className="chip bg-accent px-1.5 py-0 text-[11px]">
						{posLabel(item.partOfSpeech)}
					</span>
				) : null}
				{state ? <SourcePill source={state.sourceType} /> : null}
				{onRemove ? (
					<SubmitButton
						type="button"
						variant="ghost"
						size="sm"
						pending={removePending ?? false}
						onClick={onRemove}
						className="chip shrink-0 px-2 py-0.5 text-xs hover:border-primary"
					>
						<X className="size-3.5" />
						{m["vocabulary.deck.remove"]()}
					</SubmitButton>
				) : null}
			</div>

			<p className="text-[15px] font-medium text-muted-foreground">
				{item.definition}
			</p>

			{context ? (
				<div className="bubble bubble-partner text-[15px]">
					“{context.quote}”
				</div>
			) : null}

			<div className="mt-auto flex items-center gap-2 pt-1">
				<CefrBadge value={item.cefr} />
				{state ? (
					<SourceRef type={state.sourceType} id={state.sourceId} />
				) : null}
				<Button asChild size="sm" className="btn btn-primary ml-auto">
					{deckId ? (
						<Link
							to={APP_ROUTES.VOCABULARY.DECKS.WORD}
							params={{ deckId, itemId: item.id }}
						>
							{m["vocabulary.card.review"]()}
						</Link>
					) : (
						<Link
							to={APP_ROUTES.VOCABULARY.WORDS.DETAIL}
							params={{ itemId: item.id }}
						>
							{m["vocabulary.card.review"]()}
						</Link>
					)}
				</Button>
			</div>

			<span
				className={cn(
					"text-[11px] font-bold",
					mastered
						? "text-accuracy"
						: due
							? "text-ai-coral"
							: "text-muted-foreground",
				)}
			>
				{mastered
					? m["vocabulary.card.retained"]()
					: due
						? m["vocabulary.card.dueToday"]()
						: state?.srsDueAt?.slice(0, 10)}
			</span>
		</article>
	);
}
