import { useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { AddWordDialog } from "#/features/vocabulary/components/add-word-dialog";
import {
  VocabularyStatsRow,
  WordGrid,
} from "#/features/vocabulary/components/word-grid";
import {
  libraryQueryOptions,
  useMyVocabulary,
  useVocabularyStats,
  vocabularyStatsQueryOptions,
} from "#/features/vocabulary/queries";
import { m } from "#/paraglide/messages";
import type { VocabularySort } from "@engflex/contracts";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_app/vocabulary/words/")({
  validateSearch: (
    search: Record<string, unknown>,
  ): { sort?: VocabularySort } => ({
    sort:
      typeof search.sort === "string"
        ? (search.sort as VocabularySort)
        : undefined,
  }),
  loader: async ({ context: { queryClient } }) => {
    await Promise.all([
      queryClient.query(libraryQueryOptions()),
      queryClient.query(vocabularyStatsQueryOptions()),
    ]);
  },
  component: VocabularyWordsPage,
});

function VocabularyWordsPage() {
  const { sort } = Route.useSearch();
  useBreadcrumbs([
    { label: m["nav.item.vocabulary"]() },
    { label: m["vocabulary.decks.allWords"]() },
  ]);
  const library = useMyVocabulary();
  const stats = useVocabularyStats();
  if (library.isError) throw library.error;
  if (stats.isError) throw stats.error;
  // Never return null: the frame renders while the data loads (CardShell holds
  // the shape), so the layout does not flash empty.
  const pending = library.isPending || stats.isPending;

  const items = library.data?.items ?? [];
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
        icon: <MoMascot variant="open" size={84} />,
        title: m["vocabulary.hub.title"](),
        description: m["vocabulary.hub.subtitle"](),
      }}
    >
      <WordGrid
        items={items}
        total={vocabularyStats.totalSaved}
        statsRow={
          <VocabularyStatsRow
            stats={vocabularyStats}
            pending={pending}
            action={<AddWordDialog className="btn btn-primary" />}
          />
        }
        emptyTitle={m["vocabulary.hub.emptyTitle"]()}
        initialSort={sort}
        pending={pending}
      />
    </PageLayout>
  );
}
