import type { ConversationSummary } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { ChevronRight } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { MoMascot } from "#/components/common/mo-mascot";
import { Badge } from "#/components/ui/badge";
import { Skeleton } from "#/components/ui/skeleton";
import { personaInitials, tileTone } from "#/features/voice/persona";
import { cn } from "#/lib/utils";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";

function formatWhen(iso: string): string {
	return new Intl.DateTimeFormat(getLocale(), {
		day: "numeric",
		month: "short",
		hour: "2-digit",
		minute: "2-digit",
	}).format(new Date(iso));
}

/** One finished session, as a link to its review. Shared by the history
 *  list, the voice hub preview, and the scenario detail runs section. */
export function HistoryRow({
	conversation,
}: {
	conversation: ConversationSummary;
}) {
	const persona = conversation.scenario?.persona;
	const title =
		conversation.scenario?.title ??
		(conversation.mode === "free_talk"
			? m["voice.history.freeTalk"]()
			: m["voice.history.roleplayFallback"]());
	const minutes =
		conversation.durationSec == null
			? null
			: Math.max(1, Math.round(conversation.durationSec / 60));
	const meta = persona
		? `${persona.name} · ${persona.roleTitle} · ${formatWhen(conversation.startedAt)}`
		: formatWhen(conversation.startedAt);

	return (
		<Link
			to={APP_ROUTES.VOICE.HISTORY.DETAIL}
			params={{ conversationId: conversation.id }}
			className="flex items-center gap-3 px-4 py-3 transition hover:bg-muted/40"
		>
			{persona ? (
				<span
					className={cn(
						"tile shrink-0 text-xs font-bold",
						tileTone(persona.name),
					)}
				>
					{personaInitials(persona.roleTitle)}
				</span>
			) : (
				<MoMascot variant="open" size={36} />
			)}
			<div className="min-w-0 flex-1">
				<p className="truncate text-sm font-bold text-foreground">{title}</p>
				<p className="truncate text-xs font-medium text-muted-foreground">
					{meta}
				</p>
			</div>
			{conversation.status === "failed" ? (
				<Badge variant="secondary" className="shrink-0">
					{m["voice.history.interrupted"]()}
				</Badge>
			) : null}
			{minutes === null ? null : (
				<span className="shrink-0 text-xs font-bold text-muted-foreground">
					{m["voice.history.duration"]({ minutes })}
				</span>
			)}
			<ChevronRight className="size-4 shrink-0 text-muted-foreground" />
		</Link>
	);
}

/** Shape-matched placeholder for a pending HistoryRow list — same avatar +
 *  two-line rhythm, so nothing shifts when the rows land. Used inside a
 *  CardShell (list page) and inside the always-rendered runs section. */
export function HistoryRowSkeleton() {
	return (
		<div className="flex items-center gap-3 px-4 py-3">
			<Skeleton className="size-9 shrink-0 rounded-xl" />
			<div className="flex flex-1 flex-col gap-1.5">
				<Skeleton className="h-4 w-1/2" />
				<Skeleton className="h-3 w-2/3" />
			</div>
		</div>
	);
}
