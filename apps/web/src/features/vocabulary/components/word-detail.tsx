import { Link } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import { AcousticBreakdownCard } from "#/features/vocabulary/components/acoustic-breakdown-card";
import { CollocationsGrid } from "#/features/vocabulary/components/collocations-grid";
import { ContextsSection } from "#/features/vocabulary/components/contexts-section";
import { MorphologyCard } from "#/features/vocabulary/components/morphology-card";
import { PersonalNoteCard } from "#/features/vocabulary/components/personal-note-card";
import { WordHeroCard } from "#/features/vocabulary/components/word-hero-card";
import { useVocabularyItem } from "#/features/vocabulary/queries";
import { m } from "#/paraglide/messages";

/**
 * One word's detail. `deckId` is the context the word was opened from: set when
 * the word came from a deck route, absent from the flat grid. It drives the back
 * action and the breadcrumb trail, so "back" returns where the reader was rather
 * than guessing.
 */
export function WordDetail({
	itemId,
	deckId,
	deckName,
}: {
	itemId: string;
	deckId?: string;
	/** Only known in the deck context; the nested route loads it. */
	deckName?: string;
}) {
	const itemQuery = useVocabularyItem(itemId);
	const item = itemQuery.data;
	useBreadcrumbs(
		deckId
			? [
					{
						label: m["nav.item.vocabulary"](),
						to: APP_ROUTES.VOCABULARY.DECKS.LIST,
					},
					{
						label: deckName ?? m["vocabulary.decks.title"](),
						to: APP_ROUTES.VOCABULARY.DECKS.DETAIL,
						params: { deckId },
					},
					{ label: item?.term ?? "" },
				]
			: [
					{
						label: m["nav.item.vocabulary"](),
						to: APP_ROUTES.VOCABULARY.DECKS.LIST,
					},
					{
						label: m["vocabulary.decks.allWords"](),
						to: APP_ROUTES.VOCABULARY.WORDS.LIST,
					},
					{ label: item?.term ?? "" },
				],
	);

	if (itemQuery.isError) throw itemQuery.error;

	// The back action is context-derived and needs no data, so it renders while
	// the word loads — the frame never flashes empty.
	const backAction = (
		<Button asChild variant="ghost" size="sm" className="w-fit">
			{deckId ? (
				<Link to={APP_ROUTES.VOCABULARY.DECKS.DETAIL} params={{ deckId }}>
					<ArrowLeft data-icon="inline-start" />
					{m["common.actions.back"]()}
				</Link>
			) : (
				<Link to={APP_ROUTES.VOCABULARY.WORDS.LIST}>
					<ArrowLeft data-icon="inline-start" />
					{m["common.actions.back"]()}
				</Link>
			)}
		</Button>
	);

	if (itemQuery.isPending || !item) {
		return (
			<PageLayout action={backAction}>
				<CardShell
					pending
					className="flex flex-col gap-3 p-5"
					skeleton={
						<>
							<Skeleton className="h-8 w-48" />
							<Skeleton className="h-4 w-64" />
							<Skeleton className="h-4 w-40" />
						</>
					}
				/>
			</PageLayout>
		);
	}
	const details = item.details;

	return (
		<PageLayout action={backAction}>
			<WordHeroCard item={item} />

			{details ? (
				<>
					<ContextsSection term={item.term} contexts={details.contexts} />
					<CollocationsGrid collocations={details.collocations} />
					<div className="grid gap-5 lg:grid-cols-2">
						<MorphologyCard
							wordForms={details.wordForms}
							etymology={details.etymology}
						/>
						<AcousticBreakdownCard
							syllables={details.syllables}
							stressTip={details.stressTip}
							commonSlip={details.commonSlip}
						/>
					</div>
					{details.longDefinition ? (
						<section className="rounded-xl border bg-card p-5">
							<h2 className="text-sm font-bold tracking-tight text-foreground">
								{m["vocabulary.card.fullDefinition"]()}
							</h2>
							<p className="mt-2 text-sm text-muted-foreground">
								{details.longDefinition}
							</p>
						</section>
					) : null}
				</>
			) : null}

			{item.userState ? (
				<PersonalNoteCard
					itemId={item.id}
					note={item.userState.note}
					createdAt={item.userState.createdAt}
				/>
			) : null}
		</PageLayout>
	);
}
