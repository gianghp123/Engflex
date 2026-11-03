import type {
	ConversationObjective,
	ConversationSummaryDTO,
} from "@engflex/contracts";
import { Lightbulb } from "lucide-react";
import { CardShell } from "#/components/common/card-shell";
import { Badge } from "#/components/ui/badge";
import { Skeleton } from "#/components/ui/skeleton";
import { cn } from "#/lib/utils";
import { m } from "#/paraglide/messages";

/** House answer to dynamic copy: an explicit record keyed by objective
 *  status, never an interpolated message id. `status` narrows the contract's
 *  string so only known verdicts index the copy. */
const STATUS_COPY = {
	done: () => m["voice.room.objectives.done"](),
	partial: () => m["voice.room.objectives.partial"](),
	missed: () => m["voice.room.objectives.missed"](),
} as const;

type SummaryStatus = keyof typeof STATUS_COPY;

const STATUS_TONE: Record<SummaryStatus, string> = {
	done: "text-accuracy",
	partial: "text-warning",
	missed: "text-muted-foreground",
};

function statusOf(status: string): SummaryStatus {
	return status === "done" || status === "partial" ? status : "missed";
}

function SummarySkeleton() {
	return (
		<div className="flex flex-col gap-3">
			<Skeleton className="h-5 w-32" />
			<Skeleton className="h-4 w-full" />
			<Skeleton className="h-4 w-5/6" />
			<div className="flex items-center gap-2">
				<Skeleton className="h-6 w-16 rounded-full" />
				<Skeleton className="h-4 w-40" />
			</div>
			<Skeleton className="h-3 w-48" />
		</div>
	);
}

/**
 * Post-session recap: takeaways plus per-objective verdict pills with the
 * learner quote behind each. Joins the stored summary's objective results
 * (status + evidence) with the objectives snapshot (labels) on the int id.
 */
export function SessionSummaryCard({
	summary,
	objectives,
	pending = false,
}: {
	summary?: ConversationSummaryDTO | null;
	objectives: ConversationObjective[];
	pending?: boolean;
}) {
	const labels = new Map(
		objectives.map((objective) => [objective.id, objective.label]),
	);
	return (
		<CardShell
			pending={pending || summary === undefined}
			skeleton={<SummarySkeleton />}
			className="flex flex-col gap-3 p-5"
		>
			<div className="flex items-center gap-2">
				<span className="tile bg-secondary text-secondary-foreground">
					<Lightbulb className="size-5" />
				</span>
				<h3 className="text-base font-bold tracking-tight text-foreground">
					{m["voice.room.summary.title"]()}
				</h3>
			</div>
			{!summary ? (
				<p className="text-sm text-muted-foreground">
					{m["voice.room.summary.empty"]()}
				</p>
			) : (
				<>
					{summary.takeaways.length > 0 ? (
						<div className="flex flex-col gap-1.5">
							<span className="text-xs font-bold text-muted-foreground">
								{m["voice.room.summary.takeaways"]()}
							</span>
							<ul className="flex list-disc flex-col gap-1 pl-5 marker:text-primary">
								{summary.takeaways.map((takeaway) => (
									<li
										key={takeaway.slice(0, 48)}
										className="text-sm font-medium text-foreground"
									>
										{takeaway}
									</li>
								))}
							</ul>
						</div>
					) : null}
					{summary.objectives.length > 0 ? (
						<ul className="flex flex-col gap-2">
							{summary.objectives.map((result) => (
								<li
									key={result.id}
									className="flex flex-col gap-0.5 rounded-xl border-2 border-border bg-muted p-3"
								>
									<div className="flex items-center justify-between gap-2">
										<span className="text-sm font-bold text-foreground">
											{labels.get(result.id) ?? `#${result.id}`}
										</span>
										<Badge
											variant="outline"
											className={cn(STATUS_TONE[statusOf(result.status)])}
										>
											{STATUS_COPY[statusOf(result.status)]()}
										</Badge>
									</div>
									{result.evidence ? (
										<p className="text-xs font-medium text-muted-foreground">
											{m["voice.room.summary.evidence"]({
												quote: result.evidence,
											})}
										</p>
									) : null}
								</li>
							))}
						</ul>
					) : null}
					{summary.levelNote ? (
						<p className="text-xs text-muted-foreground">{summary.levelNote}</p>
					) : null}
				</>
			)}
		</CardShell>
	);
}
