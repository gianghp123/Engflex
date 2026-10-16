import { Link } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { HistoryRow } from "#/features/voice/components/history/history-row";
import { useRecentConversations } from "#/features/voice/queries";
import { m } from "#/paraglide/messages";

const PREVIEW_K = 3;

/** The voice hub's recent-sessions strip. Renders nothing while loading,
 *  when the learner has no history, or when the fetch fails — a first-run
 *  hub shows the modes, not an empty-state apology (the history page owns
 *  the honest empty state). */
export function HistoryPreview() {
	const recent = useRecentConversations(PREVIEW_K);
	const rows = recent.data?.items ?? [];
	if (recent.isError || rows.length === 0) return null;
	return (
		<section className="flex flex-col gap-3">
			<div className="flex items-center justify-between gap-3">
				<h2 className="text-lg font-bold tracking-tight text-foreground">
					{m["voice.history.recentTitle"]()}
				</h2>
				<Link
					to={APP_ROUTES.VOICE.HISTORY.LIST}
					className="text-sm font-bold text-primary hover:underline"
				>
					{m["voice.history.viewAll"]()}
				</Link>
			</div>
			<div className="surface-card divide-y overflow-hidden">
				{rows.map((conversation) => (
					<HistoryRow key={conversation.id} conversation={conversation} />
				))}
			</div>
		</section>
	);
}
