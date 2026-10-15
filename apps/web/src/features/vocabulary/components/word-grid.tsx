import { APP_ROUTES } from "#/app/app-route";
import { CardShell } from "#/components/common/card-shell";
import { EmptyList } from "#/components/common/empty-list";
import { ProgressBar } from "#/components/common/progress-bar";
import { StatCard } from "#/components/common/stat-card";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import { VocabCard } from "#/features/vocabulary/components/vocab-card";
import {
  DEFAULT_VOCAB_FILTERS,
  VocabFilterBar,
  type VocabFilters,
} from "#/features/vocabulary/components/vocab-filter-bar";
import { isToday } from "#/features/vocabulary/dates";
import { m } from "#/paraglide/messages";
import type {
  VocabularyItem,
  VocabularySort,
  VocabularyStats,
} from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { LibraryBig } from "lucide-react";
import { type ReactNode, useDeferredValue, useState } from "react";

/** Mastery % and the due/custom counts, computed from a VocabularyStats shape. */
export function masteryPct(stats: VocabularyStats): number {
  const total = stats.totalSaved;
  return total === 0 ? 0 : Math.round((stats.mastered / total) * 100);
}

/**
 * The vocabulary word grid: stat row (a slot — the caller owns what the cards
 * say), filter bar, cards, and the filter-empty state. All words feeds it the
 * library; a deck detail feeds it the deck's items.
 */
export function WordGrid({
  items,
  total,
  statsRow,
  emptyTitle,
  deckId,
  initialSort,
  pending,
  removeTarget,
  onRemoveItem,
  removePending,
}: {
  items: VocabularyItem[];
  total: number;
  statsRow: ReactNode;
  emptyTitle: string;
  /** Set on a deck page: cards link into that deck's word URLs. */
  deckId?: string;
  initialSort?: VocabularySort;
  /** True while the caller's data loads: the grid renders as skeletons. */
  pending?: boolean;
  /** Deck rows keyed by item id; a deck page needs the row id to remove. */
  removeTarget?: (itemId: string) => string | undefined;
  /** Present only on a deck page: removes a deck row, not the word. */
  onRemoveItem?: (deckItemId: string) => void;
  /** True for the row whose removal is in flight, so its control spins. */
  removePending?: (deckItemId: string) => boolean;
}) {
  const [filters, setFilters] = useState<VocabFilters>(() => ({
    ...DEFAULT_VOCAB_FILTERS,
    sort: initialSort ?? DEFAULT_VOCAB_FILTERS.sort,
  }));
  const deferredSearch = useDeferredValue(filters.search.trim().toLowerCase());

  const visible = items.filter((item) => {
    if (deferredSearch) {
      const haystack = `${item.term} ${item.definition}`.toLowerCase();
      if (!haystack.includes(deferredSearch)) return false;
    }
    if (item.cefr !== filters.cefr && filters.cefr !== "all") return false;
    if (
      item.userState?.sourceType !== filters.source &&
      filters.source !== "all"
    )
      return false;
    return true;
  });

  const sorted = [...visible];
  if (filters.sort === "alphabetical") {
    sorted.sort((a, b) => a.term.localeCompare(b.term));
  } else if (filters.sort === "mastery") {
    sorted.sort(
      (a, b) =>
        Number(b.userState?.mastered ?? false) -
        Number(a.userState?.mastered ?? false),
    );
  } else if (filters.sort === "interval") {
    sorted.sort((a, b) =>
      (a.userState?.srsDueAt ?? "9999").localeCompare(
        b.userState?.srsDueAt ?? "9999",
      ),
    );
  }

  return (
    <>
      {statsRow}
      {pending ? (
        <CardSkeletons count={6} />
      ) : (
        <>
          <VocabFilterBar filters={filters} onChange={setFilters} />
          {sorted.length === 0 ? (
            <EmptyList
              title={emptyTitle}
              onReset={() => setFilters(DEFAULT_VOCAB_FILTERS)}
            />
          ) : (
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              {sorted.map((item) => {
                const rowId = removeTarget?.(item.id);
                return (
                  <VocabCard
                    key={item.id}
                    item={item}
                    deckId={deckId}
                    onRemove={
                      rowId && onRemoveItem
                        ? () => onRemoveItem(rowId)
                        : undefined
                    }
                    removePending={rowId ? removePending?.(rowId) : undefined}
                  />
                );
              })}
            </div>
          )}
          <p className="text-xs font-medium text-muted-foreground">
            {m["vocabulary.hub.showing"]({ shown: sorted.length, total })}
          </p>
        </>
      )}
    </>
  );
}

/**
 * Stable keys for the skeleton cards. A skeleton list never reorders, but an
 * array index as a `key` is rejected by the linter, so the keys are literal.
 */
const SKELETON_KEYS = [
  "one",
  "two",
  "three",
  "four",
  "five",
  "six",
  "seven",
  "eight",
  "nine",
  "ten",
  "eleven",
  "twelve",
] as const;

/**
 * Deck-scoped stats from the deck's own rows. The API's stats endpoint is
 * user-global, so the deck page derives its numbers instead of lying with
 * library-wide ones.
 */
export function deckStats(items: VocabularyItem[]): VocabularyStats {
  const enrolled = items.filter((item) => item.userState);
  const now = Date.now();
  const due = (item: VocabularyItem) =>
    !item.userState?.mastered &&
    item.userState?.srsDueAt !== null &&
    item.userState?.srsDueAt !== undefined &&
    Date.parse(item.userState.srsDueAt) <= now;
  return {
    totalSaved: enrolled.length,
    mastered: enrolled.filter((item) => item.userState?.mastered).length,
    needsReview: enrolled.filter(due).length,
    // "Custom additions" means words the user created themselves, so the deck
    // scope counts its own manual-sourced rows rather than fabricating 0.
    customAdditions: enrolled.filter(
      (item) => item.userState?.sourceType === "manual",
    ).length,
    dueToday: enrolled.filter(due).length,
    // The deck's own rows saved today — same meaning as the landing's card.
    savedToday: enrolled.filter((item) => isToday(item.userState?.createdAt))
      .length,
  };
}

/** Shape-matched card skeletons for a pending grid or stat row. */
export function CardSkeletons({ count = 3 }: { count?: number }) {
  return (
    <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      {SKELETON_KEYS.slice(0, count).map((key) => (
        <CardShell
          key={key}
          pending
          className="flex flex-col gap-2 p-5"
          skeleton={
            <>
              <Skeleton className="h-3 w-24" />
              <Skeleton className="h-7 w-16" />
              <Skeleton className="h-3 w-32" />
            </>
          }
        />
      ))}
    </div>
  );
}

/** The hub's three stat cards, built from any VocabularyStats shape. */
export function VocabularyStatsRow({
  stats,
  pending = false,
  action,
}: {
  stats: VocabularyStats;
  pending?: boolean;
  /** Rendered at the foot of the third card (the deck page's add-word control). */
  action?: ReactNode;
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
        title={
          <span className="inline-flex items-center gap-1">
            {m["vocabulary.stats.needsReview"]()}
            <span className="size-1.5 rounded-full bg-secondary" />
          </span>
        }
        value={
          <>
            {stats.dueToday}{" "}
            <span className="text-sm font-medium text-muted-foreground">
              {m["vocabulary.stats.needsReviewSub"]()}
            </span>
          </>
        }
      >
        <div className="mt-auto">
          <Button
            asChild
            variant="outline"
            size="sm"
            className="btn btn-outline"
          >
            <Link
              to={APP_ROUTES.VOCABULARY.WORDS.LIST}
              search={{ sort: "interval" }}
            >
              {m["vocabulary.hub.reviewNow"]()}
            </Link>
          </Button>
        </div>
      </StatCard>
      <StatCard
        title={
          <span className="inline-flex items-center gap-1">
            <LibraryBig className="size-3.5" />
            {m["vocabulary.stats.custom"]()}
          </span>
        }
        value={stats.customAdditions}
      >
        <span className="text-xs font-medium text-muted-foreground">
          {m["vocabulary.stats.customSub"]()}
        </span>
        {action ? <div className="mt-auto">{action}</div> : null}
      </StatCard>
    </div>
  );
}
