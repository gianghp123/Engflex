import type { VocabularyDeckDetail } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { ArrowRight } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { MoMascot } from "#/components/common/mo-mascot";
import { Empty, EmptyTitle } from "#/components/ui/empty";
import { AddWordDialog } from "#/features/vocabulary/components/add-word-dialog";
import { VocabCard } from "#/features/vocabulary/components/vocab-card";
import { m } from "#/paraglide/messages";

/**
 * One deck as a landing card: its name (the way in), what it holds, one quiet
 * add-word control, and the preview cards — or a dashed slot when it is empty.
 *
 * The add control is deliberately not the page's primary: the New-deck card
 * above is, and a solid button in every section would flatten that.
 */
export function DeckSection({
	deck,
	index,
	previewK,
}: {
	deck: VocabularyDeckDetail;
	index: number;
	/** The landing shows a bounded preview; a full preview means "or more". */
	previewK: number;
}) {
	const words = deck.items ?? [];
	// `items` here is the list endpoint's preview, not the deck's membership, so
	// a full preview must not be reported as the deck's size.
	const truncated = words.length >= previewK;
	return (
		<section className="surface-card flex flex-col gap-4 p-5">
			<div className="flex flex-wrap items-start justify-between gap-3">
				<div className="min-w-0">
					<div className="flex min-w-0 items-center gap-3">
						<span className="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-sm bg-primary text-sm font-bold text-primary-foreground shadow-edge-primary-xs">
							{index + 1}
						</span>
						<Link
							to={APP_ROUTES.VOCABULARY.DECKS.DETAIL}
							params={{ deckId: deck.id }}
							className="group inline-flex min-w-0 items-center gap-1 text-lg font-bold text-foreground hover:underline"
						>
							<span className="truncate">{deck.name}</span>
							<ArrowRight className="size-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
						</Link>
						<span className="shrink-0 text-xs font-bold text-muted-foreground">
							{truncated
								? m["vocabulary.decks.wordCountMore"]({ count: words.length })
								: m["vocabulary.decks.wordCount"]({ count: words.length })}
						</span>
					</div>
					{deck.description ? (
						<p className="mt-1.5 max-w-prose text-sm font-medium text-muted-foreground">
							{deck.description}
						</p>
					) : null}
				</div>
				<div className="flex flex-wrap items-center gap-3">
					<Link
						to={APP_ROUTES.VOCABULARY.DECKS.DETAIL}
						params={{ deckId: deck.id }}
						className="text-sm font-bold text-primary hover:underline"
					>
						{m["vocabulary.decks.seeDeck"]()}
					</Link>
					<AddWordDialog deckId={deck.id} size="sm" variant="outline" />
				</div>
			</div>
			{words.length === 0 ? (
				<Empty className="gap-3 rounded-2xl border-2 border-dashed border-border p-6 md:p-6">
					<MoMascot variant="thinking" size={48} />
					<EmptyTitle className="text-sm font-bold text-muted-foreground">
						{m["vocabulary.decks.sectionEmpty"]()}
					</EmptyTitle>
				</Empty>
			) : (
				<div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
					{words.map((row) =>
						row.item ? (
							<VocabCard key={row.id} item={row.item} deckId={deck.id} />
						) : null,
					)}
				</div>
			)}
		</section>
	);
}
