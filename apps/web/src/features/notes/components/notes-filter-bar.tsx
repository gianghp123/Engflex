import { cn } from "cn";
import { Search, X } from "lucide-react";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import { m } from "#/paraglide/messages";
import { notesActions, useNotesFilter, useNotesStats } from "../store";
import type { NoteTabFilter } from "../types";
import { NoteCreateDialog } from "./note-create-dialog";

export function NotesFilterBar() {
	const { activeTab, searchQuery } = useNotesFilter();
	const stats = useNotesStats();

	const tabs: Array<{
		key: NoteTabFilter;
		label: () => string;
		count: number;
	}> = [
		{ key: "all", label: () => m["notes.tabs.all"](), count: stats.total },
		{
			key: "vocab",
			label: () => m["notes.tabs.vocab"](),
			count: stats.vocabCount,
		},
		{
			key: "lesson",
			label: () => m["notes.tabs.lesson"](),
			count: stats.lessonCount,
		},
		{
			key: "video",
			label: () => m["notes.tabs.video"](),
			count: stats.videoCount,
		},
		{
			key: "story",
			label: () => m["notes.tabs.story"](),
			count: stats.storyCount,
		},
	];

	return (
		<div className="flex flex-col gap-4">
			<div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
				{/* Tab bar */}
				<div className="flex flex-wrap items-center gap-1.5 rounded-xl bg-muted/60 p-1.5">
					{tabs.map((tab) => {
						const isActive = activeTab === tab.key;
						return (
							<button
								type="button"
								key={tab.key}
								onClick={() => notesActions.setActiveTab(tab.key)}
								className={cn(
									"flex items-center gap-2 rounded-lg px-3 py-1.5 text-xs font-semibold transition-all",
									isActive
										? "bg-background text-foreground shadow-xs"
										: "text-muted-foreground hover:bg-background/50 hover:text-foreground",
								)}
							>
								<span>{tab.label()}</span>
								<Badge
									variant={isActive ? "default" : "secondary"}
									className="px-1.5 py-0 text-[10px] font-bold"
								>
									{tab.count}
								</Badge>
							</button>
						);
					})}
				</div>

				{/* Create action button */}
				<NoteCreateDialog />
			</div>

			{/* Search input bar */}
			<div className="relative">
				<Search className="pointer-events-none absolute left-3.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
				<Input
					value={searchQuery}
					onChange={(e) => notesActions.setSearchQuery(e.target.value)}
					placeholder={m["notes.search.placeholder"]()}
					className="h-10 pl-10 pr-9 text-sm"
				/>
				{searchQuery && (
					<Button
						variant="ghost"
						size="icon"
						className="absolute right-1 top-1/2 size-8 -translate-y-1/2 text-muted-foreground hover:text-foreground"
						onClick={() => notesActions.setSearchQuery("")}
					>
						<X className="size-4" />
					</Button>
				)}
			</div>
		</div>
	);
}
