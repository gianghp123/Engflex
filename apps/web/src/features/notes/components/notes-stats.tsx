import { BookmarkCheck, Sparkles, StickyNote } from "lucide-react";
import { StatCard } from "#/components/common/stat-card";
import { m } from "#/paraglide/messages";
import { useNotesStats } from "../store";

export function NotesStats() {
	const stats = useNotesStats();

	return (
		<div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
			<StatCard
				title={m["notes.stats.total"]()}
				value={stats.total}
				size="md"
				aside={<BookmarkCheck className="size-6 text-primary" />}
			/>
			<StatCard
				title={m["notes.stats.withNotes"]()}
				value={stats.withNotes}
				size="md"
				aside={<StickyNote className="size-6 text-amber-500" />}
			/>
			<StatCard
				title={m["notes.stats.vocabCount"]()}
				value={stats.vocabCount}
				size="md"
				aside={<Sparkles className="size-6 text-emerald-500" />}
			/>
		</div>
	);
}
