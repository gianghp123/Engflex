import type { VocabularyStats } from "@engflex/contracts";
import { createFileRoute, Link } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { ProgressBar } from "#/components/common/progress-bar";
import { StatCard } from "#/components/common/stat-card";
import { DeckSection } from "#/features/vocabulary/components/deck-section";
import { NewDeckDialog } from "#/features/vocabulary/components/new-deck-dialog";
import {
	CardSkeletons,
	masteryPct,
} from "#/features/vocabulary/components/word-grid";
import { startOfLocalDay } from "#/features/vocabulary/dates";
import {
	decksQueryOptions,
	useDecks,
	useVocabularyStats,
} from "#/features/vocabulary/queries";
import { m } from "#/paraglide/messages";

const PREVIEW_K = 6;

/**
 * The reader's local midnight, RFC3339. "Saved today" is *their* day, so the
 * browser owns the day boundary and the API counts within it.
 */
function localDayStart(): string {
	return startOfLocalDay().toISOString();
}

export const Route = createFileRoute("/_app/vocabulary/decks/")({
	// Only the decks are prefetched: the stats query is keyed by the reader's
	// local day, which the server cannot know, so the client owns that read.
	loader: async ({ context: { queryClient } }) => {
		await queryClient.query(decksQueryOptions(PREVIEW_K));
	},
	component: VocabularyDecksPage,
});

/** The landing's three cards: mastery, today's activity, and the decks. */
function LandingStats({
	stats,
	deckCount,
	pending,
}: {
	stats: VocabularyStats;
	deckCount: number;
	pending: boolean;
}) {
	if (pending) return <CardSkeletons count={3} />;
	const pct = masteryPct(stats);
	return (
		<div className="grid gap-4 md:grid-cols-3">
			<StatCard
				title={m["vocabulary.stats.mastered"]()}
				value={`${pct}%`}
				size="lg"
			>
				<ProgressBar value={pct} tone="accuracy" />
				<span className="text-xs font-medium text-muted-foreground">
					{m["vocabulary.hub.masteredOf"]({
						done: stats.mastered,
						total: stats.totalSaved,
					})}
				</span>
			</StatCard>
			<StatCard
				title={m["vocabulary.stats.savedToday"]()}
				value={stats.savedToday}
			>
				<span className="text-xs font-medium text-muted-foreground">
					{m["vocabulary.stats.savedTodaySub"]()}
				</span>
			</StatCard>
			<StatCard title={m["vocabulary.stats.decks"]()} value={deckCount}>
				<span className="text-xs font-medium text-muted-foreground">
					{m["vocabulary.stats.decksSub"]()}
				</span>
				<div className="mt-auto">
					<NewDeckDialog />
				</div>
			</StatCard>
		</div>
	);
}

function VocabularyDecksPage() {
	useBreadcrumbs([
		{ label: m["nav.item.vocabulary"]() },
		{ label: m["vocabulary.decks.title"]() },
	]);
	const dayStart = localDayStart();
	const decks = useDecks(PREVIEW_K);
	const stats = useVocabularyStats(dayStart);
	if (decks.isError) throw decks.error;
	if (stats.isError) throw stats.error;
	const pending = decks.isPending || stats.isPending;

	const rows = decks.data?.items ?? [];
	const deckCount = decks.data?.pagination.total ?? rows.length;
	const vocabularyStats = stats.data ?? {
		totalSaved: 0,
		mastered: 0,
		needsReview: 0,
		customAdditions: 0,
		dueToday: 0,
		savedToday: 0,
	};

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="nice" size={84} />,
				title: m["vocabulary.hub.title"](),
				description: m["vocabulary.decks.subtitle"](),
			}}
		>
			<LandingStats
				stats={vocabularyStats}
				deckCount={deckCount}
				pending={pending}
			/>

			<div className="flex items-center justify-between gap-3">
				<h2 className="text-lg font-extrabold tracking-tight">
					{m["vocabulary.decks.title"]()}
				</h2>
				<Link
					to={APP_ROUTES.VOCABULARY.WORDS.LIST}
					className="text-sm font-bold text-primary hover:underline"
				>
					{m["vocabulary.decks.allWords"]()}
				</Link>
			</div>

			{pending ? (
				<CardSkeletons count={3} />
			) : rows.length === 0 ? (
				<div className="surface-card flex flex-col items-center gap-3 p-8 text-center">
					<MoMascot variant="confused" size={72} />
					<h3 className="text-lg font-bold text-foreground">
						{m["vocabulary.decks.emptyTitle"]()}
					</h3>
					<p className="max-w-md text-sm font-medium text-muted-foreground">
						{m["vocabulary.decks.emptyBody"]()}
					</p>
				</div>
			) : (
				rows.map((deck, index) => (
					<DeckSection
						key={deck.id}
						deck={deck}
						index={index}
						previewK={PREVIEW_K}
					/>
				))
			)}
		</PageLayout>
	);
}
