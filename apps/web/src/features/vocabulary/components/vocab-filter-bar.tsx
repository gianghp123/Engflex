import type {
	CEFR,
	VocabularySort,
	VocabularySource,
} from "@engflex/contracts";
import { CEFR_VALUES, VOCABULARY_SOURCE_VALUES } from "@engflex/contracts";
import { cn } from "cn";
import { Search } from "lucide-react";
import { Kbd } from "#/components/common/kbd";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { enumOptions } from "#/lib/enum-options";
import { m } from "#/paraglide/messages";

const SORT_MESSAGES: Record<VocabularySort, () => string> = {
	recent: () => m["vocabulary.filter.sort.recent"](),
	mastery: () => m["vocabulary.filter.sort.mastery"](),
	alphabetical: () => m["vocabulary.filter.sort.alphabetical"](),
	interval: () => m["vocabulary.filter.sort.interval"](),
};

export interface VocabFilters {
	search: string;
	sort: VocabularySort;
	source: "all" | VocabularySource;
	cefr: "all" | CEFR;
}

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
	return (
		<div className="surface-card flex flex-col gap-4 p-5">
			<div className="grid gap-3 sm:grid-cols-[1fr_auto]">
				<div className="relative">
					<Search className="absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
					<Input
						value={filters.search}
						onChange={(event) =>
							onChange({ ...filters, search: event.target.value })
						}
						placeholder={m["vocabulary.filter.search"]()}
						className="field pr-14 pl-8"
					/>
					<span className="absolute top-1/2 right-2.5 -translate-y-1/2">
						<Kbd>⌘K</Kbd>
					</span>
				</div>
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
