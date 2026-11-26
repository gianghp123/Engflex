import type { CounterKind, PlanKey, UsagePeriod } from "@engflex/contracts";
import {
	COUNTER_KIND_VALUES,
	PLAN_KEY_VALUES,
	USAGE_PERIOD_VALUES,
} from "@engflex/contracts";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { CardShell } from "#/components/common/card-shell";
import { SubmitButton } from "#/components/common/submit-button";
import { Field, FieldLabel } from "#/components/ui/field";
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

	function patch(index: number, next: Partial<PlanLimitInput>) {
		setForm((current) =>
			current.map((row, rowIndex) =>
				rowIndex === index ? { ...row, ...next } : row,
			),
		);
	}

	return (
		<section className="surface-card flex flex-col gap-3 p-4 sm:p-5">
			<h2 className="text-sm font-bold">
				{m["admin.premium.limits.plan"]({ plan })}
			</h2>
			<div className="flex flex-col gap-3">
				{form.map((row, index) => (
					<div
						key={row.counterKey}
						className="grid items-end gap-3 sm:grid-cols-[1fr_8rem_10rem_auto]"
					>
						<div className="text-sm font-medium">
							{COUNTER_LABELS[row.counterKey]()}
						</div>
						<Field>
							<FieldLabel htmlFor={`limits-${plan}-${row.counterKey}-value`}>
								{m["admin.premium.limits.limit"]()}
							</FieldLabel>
							<Input
								id={`limits-${plan}-${row.counterKey}-value`}
								type="number"
								min={0}
								max={100000}
								value={row.limitValue}
								onChange={(event) =>
									patch(index, { limitValue: Number(event.target.value) })
								}
							/>
						</Field>
						<Field>
							<FieldLabel htmlFor={`limits-${plan}-${row.counterKey}-period`}>
								{m["admin.premium.limits.period"]()}
							</FieldLabel>
							<Select
								value={row.period}
								onValueChange={(value) =>
									patch(index, { period: value as UsagePeriod })
								}
							>
								<SelectTrigger
									id={`limits-${plan}-${row.counterKey}-period`}
									className="w-full"
								>
									<SelectValue />
								</SelectTrigger>
								<SelectContent>
									{enumOptions<UsagePeriod>(
										USAGE_PERIOD_VALUES,
										PERIOD_LABELS,
									).map((option) => (
										<SelectItem key={option.value} value={option.value}>
											{option.label()}
										</SelectItem>
									))}
								</SelectContent>
							</Select>
						</Field>
						<Field className="flex-row items-center justify-between gap-2">
							<FieldLabel htmlFor={`limits-${plan}-${row.counterKey}-enabled`}>
								{m["admin.premium.limits.enabled"]()}
							</FieldLabel>
							<Switch
								id={`limits-${plan}-${row.counterKey}-enabled`}
								checked={row.isEnabled}
								onCheckedChange={(checked) =>
									patch(index, { isEnabled: checked })
								}
							/>
						</Field>
					</div>
				))}
			</div>
			<div>
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
	const plans = PLAN_KEY_VALUES.filter((plan) => drafts[plan].length > 0);

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

	return (
		<CardShell
			pending={query.isPending}
			skeleton={<Skeleton className="h-64 w-full" />}
		>
			<div className="flex flex-col gap-4">
				{plans.map((plan) => (
					<PlanSection
						key={plan}
						plan={plan}
						rows={drafts[plan]}
						onSave={(rows) => save(plan, rows)}
						pending={updateMutation.isPending && savingPlan === plan}
					/>
				))}
			</div>
		</CardShell>
	);
}
