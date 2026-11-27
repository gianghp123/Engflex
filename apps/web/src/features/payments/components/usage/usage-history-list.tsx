import type { UsageHistoryItem } from "@engflex/contracts";
import { CardShell } from "#/components/common/card-shell";
import { Pager } from "#/components/common/pager";
import { Badge } from "#/components/ui/badge";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "#/components/ui/table";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";

/** User-facing label for a usage action string. The wire sends open
 *  strings, so anything unrecognized renders as "other" rather than
 *  leaking an internal ref type. Every branch is a static message key. */
export function usageActionLabel(action: string): string {
	switch (action) {
		case "voice_session_start":
			return m["payments.usage.actions.voice_session_start"]();
		case "exam_ai_review":
			return m["payments.usage.actions.exam_ai_review"]();
		case "study_plan":
			return m["payments.usage.actions.study_plan"]();
		case "writing_score":
			return m["payments.usage.actions.writing_score"]();
		default:
			return m["payments.usage.actions.other"]();
	}
}

function formatDateTime(iso: string): string {
	return new Date(iso).toLocaleString(getLocale(), {
		day: "numeric",
		month: "short",
		year: "numeric",
		hour: "2-digit",
		minute: "2-digit",
	});
}

/** The paginated charge tail: one row per settled charge (committed or
 *  reserved), newest first. Reserved rows are in-progress holds explaining
 *  missing credits, so they carry a status chip; committed rows need none.
 *  The caller owns the page in the URL — this list only reports page
 *  changes back. */
export function UsageHistoryList({
	items,
	page,
	pageSize,
	total,
	onPageChange,
}: {
	items: UsageHistoryItem[];
	page: number;
	pageSize: number;
	total: number;
	onPageChange: (page: number) => void;
}) {
	const totalPages = Math.max(1, Math.ceil(total / pageSize));
	return (
		<section className="flex flex-col gap-3">
			<h2 className="text-base font-bold">
				{m["payments.usage.history.title"]()}
			</h2>
			{items.length === 0 ? (
				<CardShell className="flex flex-col items-center gap-2 p-12 text-center">
					<h3 className="font-bold">
						{m["payments.usage.history.emptyTitle"]()}
					</h3>
					<p className="text-muted-foreground">
						{m["payments.usage.history.emptyBody"]()}
					</p>
				</CardShell>
			) : (
				<>
					<CardShell className="p-4 sm:p-6">
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>{m["payments.usage.history.action"]()}</TableHead>
									<TableHead className="text-right">
										{m["payments.usage.history.credits"]()}
									</TableHead>
									<TableHead className="text-right">
										{m["payments.usage.history.date"]()}
									</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{items.map((item) => (
									<TableRow key={item.id}>
										<TableCell>
											<span className="flex items-center gap-2">
												{usageActionLabel(item.action)}
												{item.status === "reserved" ? (
													<Badge variant="secondary">
														{m["payments.usage.history.inProgress"]()}
													</Badge>
												) : null}
											</span>
										</TableCell>
										<TableCell className="text-right tabular-nums">
											{item.credits}
										</TableCell>
										<TableCell className="text-right text-muted-foreground">
											{formatDateTime(item.createdAt)}
										</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</CardShell>
					<Pager
						page={page}
						totalPages={totalPages}
						onPageChange={onPageChange}
						summary={
							totalPages > 1
								? m["payments.usage.history.range"]({
										from: (page - 1) * pageSize + 1,
										to: Math.min(page * pageSize, total),
										total,
									})
								: m["payments.usage.history.all"]({ count: total })
						}
					/>
				</>
			)}
		</section>
	);
}
