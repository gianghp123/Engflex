import { createFileRoute, Link } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { AcousticBreakdownCard } from "#/features/vocabulary/components/acoustic-breakdown-card";
import { CollocationsGrid } from "#/features/vocabulary/components/collocations-grid";
import { ContextsSection } from "#/features/vocabulary/components/contexts-section";
import { MorphologyCard } from "#/features/vocabulary/components/morphology-card";
import { PersonalNoteCard } from "#/features/vocabulary/components/personal-note-card";
import { WordHeroCard } from "#/features/vocabulary/components/word-hero-card";
import {
	useSaveUserVocabulary,
	useVocabularyItemDetail,
	vocabularyItemQueryOptions,
} from "#/features/vocabulary/queries";
import { loadOr404 } from "#/lib/router/route-loader";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/vocabulary/$itemId")({
	// Preload here: a draft/unpublished word 404s, a premium-gated word 403s
	// (RouteErrorFallback renders the upgrade prompt), both via loadOr404's
	// shared routing — the component below never special-cases either.
	loader: ({ context: { queryClient }, params: { itemId } }) =>
		loadOr404(queryClient, vocabularyItemQueryOptions(itemId)),
	component: WordDetailPage,
});

function WordDetailPage() {
	const { itemId } = Route.useParams();
	const itemQuery = useVocabularyItemDetail(itemId);
	const saveState = useSaveUserVocabulary();

	if (itemQuery.isError) throw itemQuery.error;
	const item = itemQuery.data;
	useBreadcrumbs(
		item
			? [
					{ label: m["nav.item.vocabulary"](), to: APP_ROUTES.VOCABULARY.LIST },
					{ label: item.term },
				]
			: [],
	);
	if (!item) return null;
	const details = item.details;

	const userState = item.userState;

	function toggleMastered() {
		if (!userState) return;
		saveState.mutate(
			{
				itemId,
				sourceType: userState.sourceType,
				sourceId: userState.sourceId,
				note: userState.note,
				mastered: !userState.mastered,
			},
			{ onError: () => toast.error(m["common.toast.error"]()) },
		);
	}

	function saveNote(note: string) {
		if (!userState) return;
		saveState.mutate(
			{
				itemId,
				sourceType: userState.sourceType,
				sourceId: userState.sourceId,
				mastered: userState.mastered,
				note,
			},
			{ onError: () => toast.error(m["common.toast.error"]()) },
		);
	}

	return (
		<PageLayout
			action={
				<Button asChild variant="ghost" size="sm" className="w-fit">
					<Link to={APP_ROUTES.VOCABULARY.LIST}>
						<ArrowLeft data-icon="inline-start" />
						{m["common.actions.back"]()}
					</Link>
				</Button>
			}
		>
			<WordHeroCard item={item} onToggleMastered={toggleMastered} />

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
					note={item.userState.note}
					createdAt={item.userState.createdAt}
					onSave={saveNote}
				/>
			) : null}
		</PageLayout>
	);
}
