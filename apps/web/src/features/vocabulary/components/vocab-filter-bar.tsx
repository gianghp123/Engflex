import type { VocabularySort } from "@engflex/contracts";
import { CEFR_VALUES, VOCABULARY_SOURCE_VALUES } from "@engflex/contracts";
import { cn } from "cn";
import { Search } from "lucide-react";
import { type FormEvent, useState } from "react";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import type { VocabFilters } from "#/features/vocabulary/types";
import { enumOptions } from "#/lib/enum-options";
import { m } from "#/paraglide/messages";

const SORT_MESSAGES: Record<VocabularySort, () => string> = {
	recent: () => m["vocabulary.filter.sort.recent"](),
	mastery: () => m["vocabulary.filter.sort.mastery"](),
	alphabetical: () => m["vocabulary.filter.sort.alphabetical"](),
	interval: () => m["vocabulary.filter.sort.interval"](),
};

export const DEFAULT_VOCAB_FILTERS: VocabFilters = {
	search: "",
	sort: "recent",
	source: "all",
	cefr: "all",
};

const SOURCE_TABS = [
	{ value: "all" as const, label: () => m["vocabulary.filter.source.all"]() },
	...enumOptions(VOCABULARY_SOURCE_VALUES, {
		lesson: () => m["vocabulary.filter.source.lesson"](),
		conversation: () => m["vocabulary.filter.source.conversation"](),
		video: () => m["vocabulary.filter.source.video"](),
		manual: () => m["vocabulary.filter.source.manual"](),
	}),
];

const CEFR_OPTIONS = [
	{ value: "all" as const, label: () => m["vocabulary.filter.cefrAll"]() },
	...enumOptions(CEFR_VALUES),
];

export function VocabFilterBar({
	filters,
	onChange,
}: {
	filters: VocabFilters;
	onChange: (next: VocabFilters) => void;
}) {
	const [draft, setDraft] = useState(filters.search);
	const [seen, setSeen] = useState(filters.search);
	if (seen !== filters.search) {
		// The URL (back/forward, reset) is the source of truth; the draft follows it.
		setSeen(filters.search);
		setDraft(filters.search);
	}

	function submitSearch(event: FormEvent) {
		event.preventDefault();
		onChange({ ...filters, search: draft.trim() });
	}

	return (
		<div className="surface-card flex flex-col gap-4 p-5">
			<div className="grid gap-3 sm:grid-cols-[1fr_auto]">
				<form onSubmit={submitSearch} className="flex items-center gap-2">
					<Input
						value={draft}
						onChange={(event) => setDraft(event.target.value)}
						placeholder={m["vocabulary.filter.search"]()}
						className="field"
					/>
					<Button
						type="submit"
						variant="outline"
						size="icon"
						aria-label={m["vocabulary.filter.search"]()}
						disabled={draft.trim() === filters.search}
					>
						<Search />
					</Button>
				</form>
				<Select
					value={filters.sort}
					onValueChange={(value) =>
						onChange({ ...filters, sort: value as VocabularySort })
					}
				>
					<SelectTrigger className="w-[190px]">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{(Object.keys(SORT_MESSAGES) as VocabularySort[]).map((sort) => (
							<SelectItem key={sort} value={sort}>
								{SORT_MESSAGES[sort]()}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</div>

			<div className="flex flex-wrap items-center gap-2">
				{SOURCE_TABS.map((tab) => (
					<button
						key={tab.value}
						type="button"
						onClick={() => onChange({ ...filters, source: tab.value })}
						className={cn(
							"chip px-3 py-1 text-xs transition",
							filters.source === tab.value
								? "chip-selected"
								: "hover:border-primary",
						)}
					>
						{tab.label()}
					</button>
				))}
			</div>

			<div className="flex flex-wrap items-center gap-2">
				<span className="text-[11px] font-semibold text-muted-foreground">
					{m["vocabulary.filter.cefrLabel"]()}
				</span>
				{CEFR_OPTIONS.map((option) => (
					<button
						key={option.value}
						type="button"
						onClick={() => onChange({ ...filters, cefr: option.value })}
						className={cn(
							"chip px-2.5 py-1 text-[11px] transition",
							filters.cefr === option.value
								? "chip-selected"
								: "hover:border-primary",
						)}
					>
						{option.label()}
					</button>
				))}
			</div>
		</div>
	);
}
