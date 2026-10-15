import type { CEFR } from "@engflex/contracts";
import { CEFR_VALUES } from "@engflex/contracts";
import { cn } from "cn";
import { Search } from "lucide-react";
import { Input } from "#/components/ui/input";
import { enumOptions } from "#/lib/enum-options";
import { m } from "#/paraglide/messages";

export interface LessonFilters {
	search: string;
	level: "all" | CEFR;
}

export const DEFAULT_LESSON_FILTERS: LessonFilters = {
	search: "",
	level: "all",
};

const LEVEL_PILLS = [
	{ value: "all" as const, label: () => m["lessons.hub.levelAll"]() },
	...enumOptions(CEFR_VALUES),
];

export function LessonFilterBar({
	filters,
	onChange,
}: {
	filters: LessonFilters;
	onChange: (next: LessonFilters) => void;
}) {
	return (
		<div className="flex flex-col gap-3">
			<div className="relative">
				<Search className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
				<Input
					value={filters.search}
					onChange={(event) =>
						onChange({ ...filters, search: event.target.value })
					}
					placeholder={m["lessons.filter.search"]()}
					className="field pl-9"
				/>
			</div>
			<div className="flex flex-wrap items-center gap-2">
				<span className="text-[11px] font-bold text-muted-foreground">
					{m["lessons.hub.levelLabel"]()}
				</span>
				{LEVEL_PILLS.map((pill) => (
					<button
						key={pill.value}
						type="button"
						onClick={() => onChange({ ...filters, level: pill.value })}
						className={cn(
							"chip px-3 py-1 text-xs transition",
							filters.level === pill.value
								? "chip-selected"
								: "hover:border-primary",
						)}
					>
						{pill.label()}
					</button>
				))}
			</div>
		</div>
	);
}
