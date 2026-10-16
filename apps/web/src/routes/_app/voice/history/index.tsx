import type { ConversationSummary } from "@engflex/contracts";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { X } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { CardShell } from "#/components/common/card-shell";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import {
	HistoryRow,
	HistoryRowSkeleton,
} from "#/features/voice/components/history/history-row";
import { useConversationHistory, useScenario } from "#/features/voice/queries";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";

export const Route = createFileRoute("/_app/voice/history/")({
	validateSearch: (
		search: Record<string, unknown>,
	): { scenarioId?: string } => ({
		scenarioId:
			typeof search.scenarioId === "string" ? search.scenarioId : undefined,
	}),
	component: VoiceHistoryPage,
});

function dayLabel(date: Date): string {
	const today = new Date();
	const yesterday = new Date(today);
	yesterday.setDate(today.getDate() - 1);
	if (date.toDateString() === today.toDateString()) {
		return m["voice.history.today"]();
	}
	if (date.toDateString() === yesterday.toDateString()) {
		return m["voice.history.yesterday"]();
	}
	return new Intl.DateTimeFormat(getLocale(), {
		day: "numeric",
		month: "long",
		year: "numeric",
	}).format(date);
}

type DayGroup = { key: string; label: string; rows: ConversationSummary[] };

/** Rows arrive newest first, so day groups are contiguous and stay in order. */
function groupByDay(rows: ConversationSummary[]): DayGroup[] {
	const groups: DayGroup[] = [];
	for (const row of rows) {
		const date = new Date(row.startedAt);
		const key = date.toDateString();
		const last = groups[groups.length - 1];
		if (last?.key === key) {
			last.rows.push(row);
		} else {
			groups.push({ key, label: dayLabel(date), rows: [row] });
		}
	}
	return groups;
}

/** The scenario scope set by the scenario detail's "View all" link: a chip
 *  plus a clear action — no filter UI. Hidden until the title resolves. */
function ScenarioFilterChip({ scenarioId }: { scenarioId: string }) {
	const navigate = useNavigate();
	const scenario = useScenario(scenarioId);
	if (!scenario.data) return null;
	return (
		<div className="flex flex-wrap items-center gap-2">
			<span className="chip chip-selected px-3 py-1 text-xs">
				{m["voice.history.filter.scenario"]({ title: scenario.data.title })}
			</span>
			<button
				type="button"
				onClick={() => {
					void navigate({ to: APP_ROUTES.VOICE.HISTORY.LIST, search: {} });
				}}
				className="inline-flex items-center gap-1 text-xs font-bold text-muted-foreground hover:text-foreground"
			>
				<X className="size-3.5" />
				{m["voice.history.filter.clear"]()}
			</button>
		</div>
	);
}

function VoiceHistoryPage() {
	useBreadcrumbs([
		{ label: m["nav.item.voice"](), to: APP_ROUTES.VOICE.LIST },
		{ label: m["voice.crumb.history"]() },
	]);
	const { scenarioId } = Route.useSearch();
	const history = useConversationHistory(scenarioId);
	if (history.isError) throw history.error;

	const rows = history.data?.pages.flatMap((page) => page.items) ?? [];

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="thinking" size={84} />,
				title: m["voice.history.title"](),
				description: m["voice.history.subtitle"](),
			}}
		>
			{scenarioId ? <ScenarioFilterChip scenarioId={scenarioId} /> : null}

			{history.isPending ? (
				<CardShell
					pending
					skeleton={
						<>
							<HistoryRowSkeleton />
							<HistoryRowSkeleton />
							<HistoryRowSkeleton />
						</>
					}
				/>
			) : rows.length === 0 ? (
				<div className="surface-card flex flex-col items-center gap-3 p-8 text-center">
					<MoMascot variant="confused" size={72} />
					<h3 className="text-lg font-bold text-foreground">
						{m["voice.history.emptyTitle"]()}
					</h3>
					<p className="max-w-md text-sm font-medium text-muted-foreground">
						{m["voice.history.emptyBody"]()}
					</p>
					<Button asChild className="btn btn-primary">
						<Link to={APP_ROUTES.VOICE.LIST}>
							{m["voice.history.emptyCta"]()}
						</Link>
					</Button>
				</div>
			) : (
				groupByDay(rows).map((group) => (
					<section key={group.key} className="flex flex-col gap-2">
						<h2 className="text-[11px] font-bold tracking-wide text-muted-foreground uppercase">
							{group.label}
						</h2>
						<div className="surface-card divide-y overflow-hidden">
							{group.rows.map((conversation) => (
								<HistoryRow key={conversation.id} conversation={conversation} />
							))}
						</div>
					</section>
				))
			)}

			{history.hasNextPage ? (
				<SubmitButton
					type="button"
					className="btn w-fit"
					pending={history.isFetchingNextPage}
					onClick={() => {
						void history.fetchNextPage();
					}}
				>
					{m["voice.history.loadMore"]()}
				</SubmitButton>
			) : null}
		</PageLayout>
	);
}
