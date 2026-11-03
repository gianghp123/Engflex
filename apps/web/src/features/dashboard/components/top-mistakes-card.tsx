import { Link } from "@tanstack/react-router";
import { ClipboardList } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { CardShell } from "#/components/common/card-shell";
import { Skeleton } from "#/components/ui/skeleton";
import { useMistakes, useTopErrors } from "#/features/progress/queries";
import type { Mistake } from "#/features/progress/service";
import { m } from "#/paraglide/messages";

const TOP_ERRORS_LIMIT = 5;
const EXAMPLES_LIMIT = 20;

function TopMistakeRow({
	errorType,
	count,
	example,
}: {
	errorType: string;
	count: number;
	example?: Mistake;
}) {
	return (
		<li className="flex items-center justify-between gap-2">
			<div className="flex min-w-0 flex-col gap-0.5">
				<span className="text-sm font-bold text-foreground">{errorType}</span>
				{example ? (
					<span className="truncate text-xs font-medium text-muted-foreground">
						{m["mistakes.row.example"]({
							text: example.text,
							correction: example.correction,
						})}
					</span>
				) : null}
			</div>
			<span className="chip shrink-0 px-2.5 py-1 text-[11px] font-bold">
				{m["mistakes.top.count"]({ count })}
			</span>
		</li>
	);
}

/**
 * Dashboard "Top mistakes" card: the caller's most common error types with
 * an example each, linking out to the mistakes history view. Buckets come
 * from `useTopErrors`; the one example per bucket comes from the recent
 * mistakes feed (newest row matching the bucket).
 */
export function TopMistakesCard() {
	const top = useTopErrors(TOP_ERRORS_LIMIT);
	const recent = useMistakes({ limit: EXAMPLES_LIMIT });
	const rows = top.data ?? [];
	const examples = new Map<string, Mistake>();
	for (const mistake of recent.data ?? []) {
		if (!examples.has(mistake.errorType))
			examples.set(mistake.errorType, mistake);
	}
	return (
		<CardShell
			pending={top.isPending || recent.isPending}
			skeleton={
				<div className="flex flex-col gap-3">
					<Skeleton className="h-5 w-32" />
					<Skeleton className="h-4 w-full" />
					<Skeleton className="h-4 w-5/6" />
				</div>
			}
			className="flex h-full flex-col gap-3 p-5"
		>
			<div className="flex items-center gap-2">
				<span className="tile bg-secondary text-secondary-foreground">
					<ClipboardList className="size-5" />
				</span>
				<h2 className="text-base font-bold tracking-tight text-foreground">
					{m["mistakes.top.title"]()}
				</h2>
			</div>
			{rows.length === 0 ? (
				<p className="text-sm text-muted-foreground">
					{m["mistakes.top.empty"]()}
				</p>
			) : (
				<ul className="flex flex-col gap-2.5">
					{rows.map((stat) => (
						<TopMistakeRow
							key={stat.errorType}
							errorType={stat.errorType}
							count={stat.count}
							example={examples.get(stat.errorType)}
						/>
					))}
				</ul>
			)}
			<Link
				to={APP_ROUTES.MISTAKES}
				className="mt-auto text-sm font-bold text-primary hover:underline"
			>
				{m["mistakes.top.viewAll"]()}
			</Link>
		</CardShell>
	);
}
