import type { CounterKind, PlanKey, UsagePeriod } from "@engflex/contracts";
import {
	COUNTER_KIND_VALUES,
	PLAN_KEY_VALUES,
	USAGE_PERIOD_VALUES,
} from "@engflex/contracts";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { SubmitButton } from "#/components/common/submit-button";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { Skeleton } from "#/components/ui/skeleton";
import { Switch } from "#/components/ui/switch";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import {
	useAdminLimits,
	useUpdatePlanLimits,
} from "#/features/admin/premium/queries";
import type { PlanLimitInput } from "#/features/admin/premium/service";
import { enumOptions } from "#/lib/enum-options";
import { m } from "#/paraglide/messages";

const COUNTER_LABELS: Record<CounterKind, () => string> = {
	ai_credit: () => m["admin.premium.limits.aiCredit"](),
	voice_session: () => m["admin.premium.limits.voiceSession"](),
};

const PERIOD_LABELS: Record<UsagePeriod, () => string> = {
	day: () => m["admin.premium.limits.periodDay"](),
	month: () => m["admin.premium.limits.periodMonth"](),
	total: () => m["admin.premium.limits.periodTotal"](),
};

type Draft = Record<PlanKey, PlanLimitInput[]>;

function emptyDraft(): Draft {
	const next = {} as Draft;
	for (const plan of PLAN_KEY_VALUES) next[plan] = [];
	return next;
}

const EDITABLE_PLANS = PLAN_KEY_VALUES.filter((plan) => plan !== "trial");

function PlanSection({
	plan,
	rows,
	onSave,
	pending,
}: {
	plan: PlanKey;
	rows: PlanLimitInput[];
	onSave: (rows: PlanLimitInput[]) => void;
	pending: boolean;
}) {
	const [form, setForm] = useState<PlanLimitInput[]>(rows);
	useEffect(() => setForm(rows), [rows]);

	function patch(counterKey: CounterKind, next: Partial<PlanLimitInput>) {
		setForm((current) =>
			current.map((row) =>
				row.counterKey === counterKey ? { ...row, ...next } : row,
			),
		);
	}

	const columns: DataTableColumn<PlanLimitInput>[] = [
		{
			id: "counter",
			header: m["admin.premium.limits.counter"](),
			cell: (row) => (
				<span className="font-medium">{COUNTER_LABELS[row.counterKey]()}</span>
			),
		},
		{
			id: "limit",
			header: m["admin.premium.limits.limit"](),
			className: "w-36",
			cell: (row) => (
				<Input
					aria-label={`${COUNTER_LABELS[row.counterKey]()} ${m["admin.premium.limits.limit"]()}`}
					type="number"
					min={0}
					max={100000}
					className="w-28"
					value={row.limitValue}
					onChange={(event) =>
						patch(row.counterKey, {
							limitValue: Number(event.target.value),
						})
					}
				/>
			),
		},
		{
			id: "period",
			header: m["admin.premium.limits.period"](),
			className: "w-44",
			cell: (row) => (
				<Select
					value={row.period}
					onValueChange={(value) =>
						patch(row.counterKey, { period: value as UsagePeriod })
					}
				>
					<SelectTrigger
						aria-label={`${COUNTER_LABELS[row.counterKey]()} ${m["admin.premium.limits.period"]()}`}
						className="w-full"
					>
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						{enumOptions<UsagePeriod>(USAGE_PERIOD_VALUES, PERIOD_LABELS).map(
							(option) => (
								<SelectItem key={option.value} value={option.value}>
									{option.label()}
								</SelectItem>
							),
						)}
					</SelectContent>
				</Select>
			),
		},
		{
			id: "enabled",
			header: m["admin.premium.limits.enabled"](),
			className: "w-28",
			cell: (row) => (
				<Switch
					aria-label={`${COUNTER_LABELS[row.counterKey]()} ${m["admin.premium.limits.enabled"]()}`}
					checked={row.isEnabled}
					onCheckedChange={(checked) =>
						patch(row.counterKey, { isEnabled: checked })
					}
				/>
			),
		},
	];

	return (
		<section className="flex flex-col gap-3">
			<h2 className="text-sm font-bold">
				{m["admin.premium.limits.plan"]({ plan })}
			</h2>
			<DataTable
				columns={columns}
				rows={form}
				rowKey={(row) => row.counterKey}
			/>
			<div className="flex justify-end">
				<SubmitButton
					type="button"
					pending={pending}
					onClick={() => onSave(form)}
				>
					{m["admin.premium.limits.save"]()}
				</SubmitButton>
			</div>
		</section>
	);
}

export function LimitsPanel() {
	const query = useAdminLimits();
	const updateMutation = useUpdatePlanLimits();
	const [savingPlan, setSavingPlan] = useState<PlanKey | null>(null);

	const drafts: Draft = useMemo(() => {
		const next = emptyDraft();
		for (const row of query.data ?? []) {
			next[row.planKey].push({
				counterKey: row.counterKey,
				isEnabled: row.isEnabled,
				limitValue: row.limitValue,
				period: row.period,
			});
		}
		// Every counter the catalog knows about renders, even when a plan has
		// no row for it yet — the PUT upserts the whole plan.
		for (const plan of PLAN_KEY_VALUES) {
			for (const counterKey of COUNTER_KIND_VALUES) {
				if (!next[plan].some((row) => row.counterKey === counterKey)) {
					next[plan].push({
						counterKey,
						isEnabled: true,
						limitValue: 0,
						period: "month",
					});
				}
			}
		}
		return next;
	}, [query.data]);
	const plans = EDITABLE_PLANS.filter((plan) => drafts[plan].length > 0);

	function save(plan: PlanKey, rows: PlanLimitInput[]) {
		setSavingPlan(plan);
		updateMutation.mutate(
			{ plan, payload: { counters: rows } },
			{
				onSuccess: () => {
					toast.success(m["admin.premium.limits.saved"]());
					setSavingPlan(null);
				},
				onError: () => {
					toast.error(m["admin.toast.error"]());
					setSavingPlan(null);
				},
			},
		);
	}

	if (query.isPending) {
		return <Skeleton className="h-64 w-full" />;
	}
	if (query.isError) {
		return (
			<p className="text-sm text-destructive">
				{m["admin.premium.loadError"]()}
			</p>
		);
	}

	return (
		<div className="flex flex-col gap-6">
			{plans.map((plan) => (
				<PlanSection
					key={plan}
					plan={plan}
					rows={drafts[plan]}
					onSave={(rows) => save(plan, rows)}
					pending={updateMutation.isPending && savingPlan === plan}
				/>
			))}
			<p className="text-xs text-muted-foreground">
				{m["admin.premium.limits.trialMoved"]()}
			</p>
		</div>
	);
}
