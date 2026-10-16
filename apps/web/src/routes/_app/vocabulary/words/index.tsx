import type {
	CEFR,
	VocabularySort,
	VocabularySource,
} from "@engflex/contracts";
import {
	CEFR_VALUES,
	VOCABULARY_SORT_VALUES,
	VOCABULARY_SOURCE_VALUES,
} from "@engflex/contracts";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { EmptyList } from "#/components/common/empty-list";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { Pager } from "#/components/common/pager";
import { AddWordDialog } from "#/features/vocabulary/components/add-word-dialog";
import {
	DEFAULT_VOCAB_FILTERS,
	VocabFilterBar,
} from "#/features/vocabulary/components/vocab-filter-bar";
import { VocabularyGuide } from "#/features/vocabulary/components/vocabulary-guide";
import {
	CardSkeletons,
	VocabularyStatsRow,
	WordCards,
} from "#/features/vocabulary/components/word-grid";
import {
	libraryPageQueryOptions,
	useLibraryPage,
	useVocabularyStats,
	vocabularyStatsQueryOptions,
} from "#/features/vocabulary/queries";
import type { VocabFilters } from "#/features/vocabulary/types";
import {
	parseEnumParam,
	parseFilterParam,
	parsePage,
} from "#/lib/search-params";
import { m } from "#/paraglide/messages";

type WordsSearch = {
	page?: number;
	search?: string;
	sort?: VocabularySort;
	source?: "all" | VocabularySource;
	cefr?: "all" | CEFR;
};

export const Route = createFileRoute("/_app/vocabulary/words/")({
	validateSearch: (search: Record<string, unknown>): WordsSearch => ({
		page: parsePage(search.page),
		search:
			typeof search.search === "string" && search.search.trim() !== ""
				? search.search.trim()
				: undefined,
		sort: parseEnumParam(VOCABULARY_SORT_VALUES, search.sort),
		source: parseFilterParam(VOCABULARY_SOURCE_VALUES, search.source),
		cefr: parseFilterParam(CEFR_VALUES, search.cefr),
	}),
	loaderDeps: ({ search }) => search,
	loader: async ({ context: { queryClient }, deps }) => {
		await Promise.all([
			queryClient.query(
				libraryPageQueryOptions(filtersFromSearch(deps), deps.page ?? 1),
			),
			queryClient.query(vocabularyStatsQueryOptions()),
		]);
	},
	component: VocabularyWordsPage,
});

/** The URL's optional params resolved to the filter set every read uses. */
function filtersFromSearch(search: WordsSearch): VocabFilters {
	return {
		search: search.search ?? "",
		sort: search.sort ?? DEFAULT_VOCAB_FILTERS.sort,
		source: search.source ?? "all",
		cefr: search.cefr ?? "all",
	};
}

function VocabularyWordsPage() {
	useBreadcrumbs([
		{ label: m["nav.item.vocabulary"]() },
		{ label: m["vocabulary.decks.allWords"]() },
	]);
	const navigate = useNavigate();
	const search = Route.useSearch();
	const filters = filtersFromSearch(search);
	const page = search.page ?? 1;
	const library = useLibraryPage(filters, page);
	const stats = useVocabularyStats();
	if (library.isError) throw library.error;
	if (stats.isError) throw stats.error;

	const items = library.data?.items ?? [];
	const pagination = library.data?.pagination;
	const total = pagination?.total ?? 0;
	const from = pagination ? (pagination.page - 1) * pagination.pageSize + 1 : 0;
	const to = pagination
		? Math.min(pagination.page * pagination.pageSize, total)
		: 0;
	const vocabularyStats = stats.data ?? {
		totalSaved: 0,
		mastered: 0,
		needsReview: 0,
		customAdditions: 0,
		dueToday: 0,
		savedToday: 0,
	};

	/** Filters/search changes drop `page`; explicit navigation keeps it. */
	function toSearch(next: VocabFilters): WordsSearch {
		return {
			page: undefined,
			search: next.search || undefined,
			sort: next.sort === DEFAULT_VOCAB_FILTERS.sort ? undefined : next.sort,
			source: next.source === "all" ? undefined : next.source,
			cefr: next.cefr === "all" ? undefined : next.cefr,
		};
	}

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="open" size={84} />,
				title: m["vocabulary.decks.allWords"](),
				description: m["vocabulary.hub.subtitle"](),
			}}
		>
			<VocabularyStatsRow
				stats={vocabularyStats}
				pending={stats.isPending}
				action={
					<div className="flex items-center gap-2">
						<AddWordDialog className="btn btn-primary" />
						<VocabularyGuide autoOpen />
					</div>
				}
			/>
			<VocabFilterBar
				filters={filters}
				onChange={(next) =>
					void navigate({
						to: APP_ROUTES.VOCABULARY.WORDS.LIST,
						search: toSearch(next),
					})
				}
			/>
			{library.isPending ? (
				<CardSkeletons count={6} />
			) : total === 0 ? (
				<EmptyList
					title={m["vocabulary.hub.emptyTitle"]()}
					onReset={() =>
						void navigate({
							to: APP_ROUTES.VOCABULARY.WORDS.LIST,
							search: {},
						})
					}
				/>
			) : (
				<>
					<WordCards items={items} />
					<Pager
						page={page}
						totalPages={pagination?.totalPages ?? 1}
						summary={
							pagination && pagination.totalPages > 1
								? m["vocabulary.hub.showingRange"]({ from, to, total })
								: m["vocabulary.hub.showingAll"]({ total })
						}
						onPageChange={(next) =>
							void navigate({
								to: APP_ROUTES.VOCABULARY.WORDS.LIST,
								search: { ...search, page: next <= 1 ? undefined : next },
							})
						}
					/>
				</>
			)}
		</PageLayout>
	);
}
