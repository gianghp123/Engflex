import { createFileRoute } from "@tanstack/react-router";
import { useStore } from "@tanstack/react-store";
import { LibraryBig, Pencil } from "lucide-react";
import { useState } from "react";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { EmptyList } from "#/components/common/empty-list";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { ProgressBar } from "#/components/common/progress-bar";
import { Button } from "#/components/ui/button";
import { AddWordDialog } from "#/features/vocabulary/components/add-word-dialog";
import { VocabCard } from "#/features/vocabulary/components/vocab-card";
import {
	DEFAULT_VOCAB_FILTERS,
	VocabFilterBar,
	type VocabFilters,
} from "#/features/vocabulary/components/vocab-filter-bar";
import { VOCABULARY_STATS } from "#/features/vocabulary/fixtures";
import {
	getVocabularyItems,
	vocabularyStore,
} from "#/features/vocabulary/store";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/vocabulary/")({
	component: VocabularyPage,
});

function VocabularyPage() {
	useBreadcrumbs([{ label: m["nav.item.vocabulary"]() }]);
	const [filters, setFilters] = useState<VocabFilters>(DEFAULT_VOCAB_FILTERS);
	useStore(vocabularyStore); // subscribe to mastered/note/add changes

	const items = getVocabularyItems();
	const masteryPct = Math.round(
		(VOCABULARY_STATS.mastered / VOCABULARY_STATS.totalSaved) * 100,
	);

	const visible = items.filter((item) => {
		if (filters.search.trim()) {
			const haystack = `${item.term} ${item.definition}`.toLowerCase();
			if (!haystack.includes(filters.search.trim().toLowerCase())) return false;
		}
		if (item.cefr !== filters.cefr && filters.cefr !== "all") return false;
		if (item.domain !== filters.domain && filters.domain !== "all")
			return false;
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
		<PageLayout
			hero={{
				icon: <MoMascot variant="open" size={84} />,
				title: m["vocabulary.hub.title"](),
				description: m["vocabulary.hub.subtitle"](),
			}}
		>
			<div className="grid gap-4 md:grid-cols-3">
				<div className="surface-card flex flex-col gap-2 p-5">
					<span className="text-[11px] font-bold text-muted-foreground">
						{m["vocabulary.stats.mastered"]()}
					</span>
					<span className="stat-display">{masteryPct}%</span>
					<ProgressBar value={masteryPct} tone="accuracy" />
					<span className="text-xs font-medium text-muted-foreground">
						{m["vocabulary.hub.masteredOf"]({
							done: VOCABULARY_STATS.mastered,
							total: VOCABULARY_STATS.totalSaved,
						})}
					</span>
				</div>
				<div className="surface-card flex flex-col gap-2 p-5">
					<span className="inline-flex items-center gap-1 text-[11px] font-bold text-muted-foreground">
						{m["vocabulary.stats.needsReview"]()}
						<span className="size-1.5 rounded-full bg-secondary" />
					</span>
					<span className="text-2xl font-bold text-foreground">
						{VOCABULARY_STATS.dueToday}{" "}
						<span className="text-sm font-medium text-muted-foreground">
							{m["vocabulary.stats.needsReviewSub"]({
								count: VOCABULARY_STATS.dueToday,
							})}
						</span>
					</span>
					<div className="mt-auto">
						<Button
							type="button"
							variant="outline"
							size="sm"
							className="btn btn-outline"
							onClick={() =>
								setFilters((current) => ({ ...current, sort: "interval" }))
							}
						>
							{m["vocabulary.hub.reviewNow"]()}
						</Button>
					</div>
				</div>
				<div className="surface-card flex flex-col gap-2 p-5">
					<span className="inline-flex items-center gap-1 text-[11px] font-bold text-muted-foreground">
						<LibraryBig className="size-3.5" />
						{m["vocabulary.stats.custom"]()}
					</span>
					<span className="text-2xl font-bold text-foreground">
						{VOCABULARY_STATS.customAdditions}
					</span>
					<span className="text-xs font-medium text-muted-foreground">
						{m["vocabulary.stats.customSub"]()}
					</span>
					<span className="tile mt-auto bg-secondary text-secondary-foreground">
						<Pencil className="size-5" />
					</span>
				</div>
			</div>

			<VocabFilterBar filters={filters} onChange={setFilters} />

			{sorted.length === 0 ? (
				<EmptyList
					title={m["vocabulary.hub.emptyTitle"]()}
					onReset={() => setFilters(DEFAULT_VOCAB_FILTERS)}
				/>
			) : (
				<div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
					{sorted.map((item) => (
						<VocabCard key={item.id} item={item} />
					))}
				</div>
			)}

			<p className="text-xs font-medium text-muted-foreground">
				{m["vocabulary.hub.showing"]({
					shown: sorted.length,
					total: VOCABULARY_STATS.totalSaved,
				})}
			</p>

			<section className="surface-card flex flex-col gap-4 p-5 md:flex-row md:items-center">
				<span className="inline-flex h-16 w-16 shrink-0 items-center justify-center rounded-[13px] bg-secondary shadow-edge-neutral-sm">
					<MoMascot variant="wave" size={52} />
				</span>
				<div className="min-w-0 flex-1">
					<h2 className="text-lg font-bold text-foreground">
						{m["vocabulary.addWord.title"]()}
					</h2>
					<p className="text-[15px] font-medium text-muted-foreground">
						{m["vocabulary.addWord.description"]()}
					</p>
				</div>
				<div className="[&_button]:btn [&_button]:btn-primary">
					<AddWordDialog />
				</div>
			</section>
		</PageLayout>
	);
}
