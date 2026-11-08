import type { SavedLibrary } from "@engflex/contracts";
import { BookmarkCheck, Sparkles, StickyNote } from "lucide-react";
import { StatCard } from "#/components/common/stat-card";
import { m } from "#/paraglide/messages";
export function NotesStats({
	stats,
	pending,
}: {
	stats?: SavedLibrary;
	pending: boolean;
}) {
	return (
		<div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
			<StatCard
				title={m["notes.stats.total"]()}
				value={stats?.total ?? "—"}
				pending={pending}
				size="md"
				aside={<BookmarkCheck className="size-6 text-primary" />}
			/>
			<StatCard
				title={m["notes.stats.withNotes"]()}
				value={stats?.withNotes ?? "—"}
				pending={pending}
				size="md"
				aside={<StickyNote className="size-6 text-amber-500" />}
			/>
			<StatCard
				title={m["notes.stats.vocabCount"]()}
				value={stats?.vocab ?? "—"}
				pending={pending}
				size="md"
				aside={<Sparkles className="size-6 text-emerald-500" />}
			/>
		</div>
	);
}
