import { useEffect } from "react";
import { CardShell } from "#/components/common/card-shell";
import { ProgressRing, StatCard } from "#/components/common/stat-card";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "#/components/ui/table";
import {
	CreditGuideButton,
	CreditGuideDialog,
} from "#/features/payments/components/credit-guide-dialog";
import {
	UsageHistoryList,
	usageActionLabel,
} from "#/features/payments/components/usage/usage-history-list";
import { useMyUsage } from "#/features/payments/entitlements";
import { useCreditGuide } from "#/features/payments/hooks/use-credit-guide";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";

const PAGE_SIZE = 10;

/** Shape-matched pending frame for a balance card: title, metric, caption,
 *  and the ring the loaded card carries in its aside slot. */
function BalanceSkeleton() {
	return (
		<>
			<div className="flex min-w-0 flex-1 flex-col gap-2">
				<Skeleton className="h-3 w-20" />
				<Skeleton className="h-7 w-24" />
				<Skeleton className="h-3 w-28" />
			</div>
			<Skeleton className="size-16 rounded-full" />
		</>
	);
}

function resetLabel(resetAt?: string): string {
	if (!resetAt) return m["payments.usage.balance.noReset"]();
	const date = new Date(resetAt);
	if (Number.isNaN(date.getTime()))
		return m["payments.usage.balance.noReset"]();
	return m["payments.usage.balance.resetOn"]({
		date: date.toLocaleDateString(getLocale()),
	});
}

function UsagePending() {
	return (
		<>
			<div className="grid gap-4 md:grid-cols-2">
				{[0, 1].map((card) => (
					<StatCard
						key={card}
						title=""
						value=""
						pending
						skeleton={<BalanceSkeleton />}
					/>
				))}
			</div>
			<CardShell
				pending
				className="p-4 sm:p-6"
				skeleton={<Skeleton className="h-40 w-full" />}
			/>
			<CardShell
				pending
				className="p-4 sm:p-6"
				skeleton={<Skeleton className="h-64 w-full" />}
			/>
		</>
	);
}

/** The caller's usage surface: both balance cards, the per-action costs
 *  table (numbers always from the response, never hardcoded), and the
 *  paginated charge history. Errors render inline with a retry — the same
 *  convention as the sibling authed pages (pricing, notes). */
export function UsagePage({
	page,
	onPageChange,
}: {
	page: number;
	onPageChange: (page: number, opts?: { replace?: boolean }) => void;
}) {
	const query = useMyUsage(page, PAGE_SIZE);
	const guide = useCreditGuide();
	const total = query.data?.history.total ?? 0;
	const historyPageSize = query.data?.history.pageSize ?? PAGE_SIZE;
	const lastPage = Math.max(1, Math.ceil(total / historyPageSize));

	useEffect(() => {
		if (query.data && page > lastPage) {
			onPageChange(lastPage, { replace: true });
		}
	}, [query.data, page, lastPage, onPageChange]);

	const counters = query.data?.counters ?? [];
	const credits = counters.find((counter) => counter.counter === "ai_credit");
	const voice = counters.find((counter) => counter.counter === "voice_session");

	return (
		<div className="flex flex-col gap-6">
			<div className="flex items-start justify-between gap-3">
				<div>
					<h1 className="text-2xl font-bold">{m["payments.usage.title"]()}</h1>
					<p className="mt-1 text-muted-foreground">
						{m["payments.usage.subtitle"]()}
					</p>
				</div>
				{query.data ? (
					<CreditGuideButton onClick={() => guide.setOpen(true)} />
				) : null}
			</div>
			{query.isPending ? (
				<UsagePending />
			) : query.isError || !query.data ? (
				<CardShell className="flex flex-col items-start gap-3 p-5">
					<p role="alert" className="text-destructive">
						{m["payments.usage.error"]()}
					</p>
					<Button
						type="button"
						variant="outline"
						disabled={query.isFetching}
						onClick={() => void query.refetch()}
					>
						{m["payments.usage.retry"]()}
					</Button>
				</CardShell>
			) : (
				<>
					<div className="grid gap-4 md:grid-cols-2">
						<StatCard
							title={m["payments.usage.balance.creditsTitle"]()}
							value={
								credits
									? m["payments.usage.balance.creditsValue"]({
											remaining: credits.remaining,
											limit: credits.limit,
										})
									: "–"
							}
							aside={
								credits ? (
									<ProgressRing
										value={
											credits.limit > 0
												? (credits.remaining / credits.limit) * 100
												: 0
										}
									/>
								) : null
							}
						>
							<span className="text-xs text-muted-foreground">
								{resetLabel(credits?.resetAt)}
							</span>
						</StatCard>
						<StatCard
							title={m["payments.usage.balance.voiceTitle"]()}
							value={
								voice
									? m["payments.usage.balance.voiceValue"]({
											used: voice.used,
											limit: voice.limit,
										})
									: "–"
							}
							aside={
								voice ? (
									<ProgressRing
										tone="violet"
										value={
											voice.limit > 0 ? (voice.used / voice.limit) * 100 : 0
										}
									/>
								) : null
							}
						>
							<span className="text-xs text-muted-foreground">
								{resetLabel(voice?.resetAt)}
							</span>
						</StatCard>
					</div>
					<section className="flex flex-col gap-3">
						<h2 className="text-base font-bold">
							{m["payments.usage.costs.title"]()}
						</h2>
						<CardShell className="p-4 sm:p-6">
							<Table>
								<TableHeader>
									<TableRow>
										<TableHead>{m["payments.usage.costs.action"]()}</TableHead>
										<TableHead className="text-right">
											{m["payments.usage.costs.credits"]()}
										</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody>
									{query.data.costs.map((cost) => (
										<TableRow key={cost.action}>
											<TableCell>{usageActionLabel(cost.action)}</TableCell>
											<TableCell className="text-right tabular-nums">
												{cost.credits}
											</TableCell>
										</TableRow>
									))}
								</TableBody>
							</Table>
						</CardShell>
					</section>
					<UsageHistoryList
						items={query.data.history.items}
						page={page}
						pageSize={historyPageSize}
						total={total}
						onPageChange={(next) => onPageChange(next)}
					/>
				</>
			)}
			{query.data ? (
				<CreditGuideDialog
					open={guide.open}
					onClose={guide.close}
					costs={query.data.costs}
				/>
			) : null}
		</div>
	);
}
