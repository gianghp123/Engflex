import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import { AddWordDialog } from "#/features/vocabulary/components/add-word-dialog";
import { DeckInfo } from "#/features/vocabulary/components/deck-info";
import {
  deckStats,
  VocabularyStatsRow,
  WordGrid,
} from "#/features/vocabulary/components/word-grid";
import {
  deckQueryOptions,
  useDeck,
  useRemoveDeckItem,
} from "#/features/vocabulary/queries";
import { m } from "#/paraglide/messages";
import { createFileRoute, Link } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/_app/vocabulary/decks/$deckId/")({
  loader: async ({ context: { queryClient }, params: { deckId } }) => {
    await queryClient.query(deckQueryOptions(deckId));
  },
  component: VocabularyDeckPage,
});

function VocabularyDeckPage() {
  const { deckId } = Route.useParams();
  const deck = useDeck(deckId);
  const removeDeckItem = useRemoveDeckItem();
  useBreadcrumbs([
    { label: m["nav.item.vocabulary"](), to: APP_ROUTES.VOCABULARY.DECKS.LIST },
    // The current page is the deck itself — no link on the last crumb.
    { label: deck.data?.name ?? m["vocabulary.decks.title"]() },
  ]);
  if (deck.isError) throw deck.error;

  const rows = deck.data?.items ?? [];
  const items = rows.flatMap((row) => (row.item ? [row.item] : []));
  const pending = deck.isPending;
  // The remove target is the deck ROW id, not the item id.
  const rowIdByItem = new Map(rows.map((row) => [row.itemId, row.id]));

  return (
    <PageLayout
      action={
        <Button asChild variant="ghost" size="sm">
          <Link to={APP_ROUTES.VOCABULARY.DECKS.LIST}>
            <ArrowLeft /> {m["vocabulary.deck.back"]()}
          </Link>
        </Button>
      }
    >
      {deck.data ? (
        <DeckInfo deck={deck.data} />
      ) : (
        <CardShell
          pending
          className="flex flex-col gap-3 p-5"
          skeleton={
            <>
              <Skeleton className="h-8 w-48" />
              <Skeleton className="h-4 w-64" />
            </>
          }
        />
      )}
      <WordGrid
        items={items}
        total={items.length}
        statsRow={
          <VocabularyStatsRow
            stats={deckStats(items)}
            pending={pending}
            action={
              <AddWordDialog deckId={deckId} className="btn-accent" />
            }
          />
        }
        emptyTitle={m["vocabulary.deck.emptyTitle"]()}
        deckId={deckId}
        pending={pending}
        removeTarget={(itemId) => rowIdByItem.get(itemId)}
        removePending={(rowId) =>
          removeDeckItem.isPending && removeDeckItem.variables === rowId
        }
        onRemoveItem={(rowId) =>
          removeDeckItem.mutate(rowId, {
            onSuccess: () => toast.success(m["vocabulary.deck.removed"]()),
            onError: () => toast.error(m["vocabulary.deck.removeFailed"]()),
          })
        }
      />
    </PageLayout>
  );
}
