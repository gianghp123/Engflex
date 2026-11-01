import type {
	CEFR,
	VocabularyDomain,
	VocabularySort,
	VocabularySource,
} from "@engflex/contracts";
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
import { m } from "#/paraglide/messages";

const SORT_MESSAGES: Record<VocabularySort, () => string> = {
	recent: () => m["vocabulary.filter.sort.recent"](),
	mastery: () => m["vocabulary.filter.sort.mastery"](),
	alphabetical: () => m["vocabulary.filter.sort.alphabetical"](),
	interval: () => m["vocabulary.filter.sort.interval"](),
};

const DOMAIN_MESSAGES: Record<VocabularyDomain, () => string> = {
	backend_db: () => m["vocabulary.filter.domain.backendDb"](),
	distributed_systems: () => m["vocabulary.filter.domain.distributedSystems"](),
	devops_cloud: () => m["vocabulary.filter.domain.devopsCloud"](),
	frontend_ui: () => m["vocabulary.filter.domain.frontendUi"](),
	ai_ml: () => m["vocabulary.filter.domain.aiMl"](),
};

export interface VocabFilters {
	search: string;
	sort: VocabularySort;
	source: "all" | VocabularySource;
	domain: "all" | VocabularyDomain;
	cefr: "all" | CEFR;
}

export const DEFAULT_VOCAB_FILTERS: VocabFilters = {
	search: "",
	sort: "recent",
	source: "all",
	domain: "all",
	cefr: "all",
};

const SOURCE_TABS: Array<{
	value: VocabFilters["source"];
	label: () => string;
}> = [
	{ value: "all", label: () => m["vocabulary.filter.source.all"]() },
	{ value: "lesson", label: () => m["vocabulary.filter.source.lesson"]() },
	{
		value: "conversation",
		label: () => m["vocabulary.filter.source.conversation"](),
	},
	{ value: "manual", label: () => m["vocabulary.filter.source.manual"]() },
];

const CEFR_OPTIONS: Array<{
	value: VocabFilters["cefr"];
	label: string | (() => string);
}> = [
	{ value: "all", label: () => m["vocabulary.filter.cefrAll"]() },
	{ value: "B1", label: "B1" },
	{ value: "B2", label: "B2" },
	{ value: "C1", label: "C1" },
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
					{m["vocabulary.filter.domainLabel"]()}
				</span>
				<button
					type="button"
					onClick={() => onChange({ ...filters, domain: "all" })}
					className={cn(
						"chip px-2.5 py-1 text-[11px] transition",
						filters.domain === "all" ? "chip-selected" : "hover:border-primary",
					)}
				>
					{m["vocabulary.filter.domainAll"]()}
				</button>
				{(Object.keys(DOMAIN_MESSAGES) as VocabularyDomain[]).map((domain) => (
					<button
						key={domain}
						type="button"
						onClick={() => onChange({ ...filters, domain })}
						className={cn(
							"chip px-2.5 py-1 text-[11px] transition",
							filters.domain === domain
								? "chip-selected"
								: "hover:border-primary",
						)}
					>
						{DOMAIN_MESSAGES[domain]()}
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
						{typeof option.label === "function" ? option.label() : option.label}
					</button>
				))}
			</div>
		</div>
	);
}
