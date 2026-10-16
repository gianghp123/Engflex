import type { CEFR } from "@engflex/contracts";
import { Filter, GraduationCap, Search } from "lucide-react";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { m } from "#/paraglide/messages";

export interface VideoFilters {
	search: string;
	level: "all" | CEFR;
}

export const DEFAULT_VIDEO_FILTERS: VideoFilters = {
	search: "",
	level: "all",
};

const LEVELS: CEFR[] = ["A1", "A2", "B1", "B2", "C1", "C2"];

export function VideoFilterBar({
	filters,
	onChange,
}: {
	filters: VideoFilters;
	onChange: (next: VideoFilters) => void;
}) {
	return (
		<section className="surface-card flex flex-col gap-3 p-4">
			<h2 className="flex items-center gap-2 text-sm font-extrabold tracking-wide text-foreground uppercase">
				<Filter aria-hidden="true" className="size-5" />
				{m["videos.category.filter.title"]()}
			</h2>
			<div className="flex flex-col gap-3 sm:flex-row">
				<div className="relative flex-1">
					<Search
						aria-hidden="true"
						className="absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
					/>
					<Input
						type="search"
						value={filters.search}
						onChange={(event) =>
							onChange({ ...filters, search: event.target.value })
						}
						placeholder={m["videos.category.filter.search"]()}
						aria-label={m["videos.category.filter.search"]()}
						className="field h-11 pl-9"
					/>
				</div>
				<Select
					value={filters.level}
					onValueChange={(value) =>
						onChange({ ...filters, level: value as VideoFilters["level"] })
					}
				>
					<SelectTrigger
						aria-label={m["videos.category.filter.level"]()}
						className="field h-11! w-full gap-2 sm:w-60"
					>
						<GraduationCap
							aria-hidden="true"
							className="text-muted-foreground"
						/>
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="all">
							{m["videos.category.filter.levelAll"]()}
						</SelectItem>
						{LEVELS.map((level) => (
							<SelectItem key={level} value={level}>
								{level}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</div>
		</section>
	);
}
