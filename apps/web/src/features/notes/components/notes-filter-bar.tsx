import { type SavedLibrary, TitleMaxLength } from "@engflex/contracts";
import { cn } from "cn";
import { Search } from "lucide-react";
import { Badge } from "#/components/ui/badge";
import { Input } from "#/components/ui/input";
import { m } from "#/paraglide/messages";
import type { NoteTabFilter } from "../types";

export function NotesFilterBar({
	tab,
	search,
	stats,
	onTab,
	onSearch,
}: {
	tab: NoteTabFilter;
	search: string;
	stats?: SavedLibrary;
	onTab: (tab: NoteTabFilter) => void;
	onSearch: (search: string) => void;
}) {
	const tabs: Array<{ key: NoteTabFilter; label: string; count?: number }> = [
		{ key: "all", label: m["notes.tabs.all"](), count: stats?.total },
		{ key: "vocab", label: m["notes.tabs.vocab"](), count: stats?.vocab },
		{ key: "lesson", label: m["notes.tabs.lesson"](), count: stats?.lesson },
		{ key: "video", label: m["notes.tabs.video"](), count: stats?.video },
		{ key: "story", label: m["notes.tabs.story"](), count: stats?.story },
	];

	return (
		<div className="flex flex-col gap-4">
			<div className="flex flex-wrap items-center gap-1.5 rounded-xl bg-muted/60 p-1.5">
				{tabs.map(({ key, label, count }) => (
					<button
						key={key}
						type="button"
						aria-pressed={tab === key}
						onClick={() => onTab(key)}
						className={cn(
							"flex items-center gap-2 rounded-lg px-3 py-1.5 text-xs font-semibold",
							tab === key
								? "bg-background shadow-xs"
								: "text-muted-foreground hover:bg-background/50",
						)}
					>
						{label}
						{count !== undefined && (
							<Badge variant={tab === key ? "default" : "secondary"}>
								{count}
							</Badge>
						)}
					</button>
				))}
			</div>
			<div className="relative">
				<Search className="pointer-events-none absolute left-3.5 top-3 size-4 text-muted-foreground" />
				<Input
					aria-label={m["notes.search.placeholder"]()}
					placeholder={m["notes.search.placeholder"]()}
					value={search}
					maxLength={TitleMaxLength}
					onChange={(event) => onSearch(event.target.value)}
					className="h-10 pl-10"
				/>
			</div>
		</div>
	);
}
