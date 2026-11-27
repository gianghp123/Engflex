import type { AccessAction } from "@engflex/contracts";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { SubmitButton } from "#/components/common/submit-button";
import { Input } from "#/components/ui/input";
import { Skeleton } from "#/components/ui/skeleton";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import {
	useAdminActionCosts,
	useAdminTrialTerms,
	useUpdateActionCosts,
	useUpdateTrialTerms,
} from "#/features/admin/premium/queries";
import type { ActionCostInput } from "#/features/admin/premium/service";
import { m } from "#/paraglide/messages";

const PAID_ACTIONS = [
	"voice_session_start",
	"exam_ai_review",
	"study_plan",
] as const satisfies readonly AccessAction[];

const ACTION_LABELS: Record<(typeof PAID_ACTIONS)[number], () => string> = {
	voice_session_start: () => m["admin.premium.costs.voiceSession"](),
	exam_ai_review: () => m["admin.premium.costs.examReview"](),
	study_plan: () => m["admin.premium.costs.studyPlan"](),
};

type TrialKey = "days" | "credits" | "voiceSessions";

interface TrialRow {
	id: TrialKey;
	label: () => string;
	min: number;
	max: number;
}

const TRIAL_ROWS: TrialRow[] = [
	{
		id: "days",
		label: () => m["admin.premium.trial.days"](),
		min: 1,
		max: 365,
	},
	{
		id: "credits",
		label: () => m["admin.premium.trial.credits"](),
		min: 0,
		max: 100000,
	},
	{
		id: "voiceSessions",
		label: () => m["admin.premium.trial.voiceSessions"](),
		min: 0,
		max: 100000,
	},
];

function formatUpdated(iso: string | undefined): string {
	if (!iso) return "—";
	const date = new Date(iso);
	return Number.isNaN(date.getTime()) ? "—" : date.toLocaleString();
}

export function PricingPanel() {
	const costsQuery = useAdminActionCosts();
	const trialQuery = useAdminTrialTerms();
	const saveCosts = useUpdateActionCosts();
	const saveTrial = useUpdateTrialTerms();
	const [costs, setCosts] = useState<ActionCostInput[]>([]);
	const [trial, setTrial] = useState({
		days: 7,
		credits: 30,
		voiceSessions: 7,
	});

	useEffect(() => {
		if (costsQuery.data) {
			const byAction = new Map(
				costsQuery.data.map((row) => [row.action, row.credits]),
			);
			setCosts(
				PAID_ACTIONS.map((action) => ({
					action,
					credits: byAction.get(action) ?? 0,
				})),
			);
		}
	}, [costsQuery.data]);
	useEffect(() => {
		if (trialQuery.data) {
			setTrial({
				days: trialQuery.data.days,
				credits: trialQuery.data.credits,
				voiceSessions: trialQuery.data.voiceSessions,
			});
		}
	}, [trialQuery.data]);

	if (costsQuery.isPending || trialQuery.isPending) {
		return <Skeleton className="h-64 w-full" />;
	}
	if (costsQuery.isError || trialQuery.isError) {
		return (
			<p className="text-sm text-destructive">
				{m["admin.premium.loadError"]()}
			</p>
		);
	}

	const costRows = PAID_ACTIONS.map((action) => {
		const draft = costs.find((row) => row.action === action);
		const meta = costsQuery.data?.find((row) => row.action === action);
		return {
			action,
			credits: draft?.credits ?? 0,
			updatedAt: meta?.updatedAt,
		};
	});

	function updateCost(action: AccessAction, credits: number) {
		setCosts((current) =>
			current.map((row) => (row.action === action ? { ...row, credits } : row)),
		);
	}

	function updateTrial(key: TrialKey, value: number) {
		setTrial((current) => ({ ...current, [key]: value }));
	}

	const costColumns: DataTableColumn<(typeof costRows)[number]>[] = [
		{
			id: "action",
			header: m["admin.premium.costs.action"](),
			cell: (row) => (
				<span className="font-medium">{ACTION_LABELS[row.action]()}</span>
			),
		},
		{
			id: "credits",
			header: m["admin.premium.costs.credits"](),
			className: "w-40",
			cell: (row) => (
				<Input
					aria-label={`${ACTION_LABELS[row.action]()} ${m["admin.premium.costs.credits"]()}`}
					type="number"
					min={0}
					max={100000}
					className="w-28"
					value={row.credits}
					onChange={(event) =>
						updateCost(row.action, Number(event.target.value))
					}
				/>
			),
		},
		{
			id: "updated",
			header: m["admin.premium.costs.updated"](),
			className: "w-56",
			cell: (row) => (
				<span className="text-muted-foreground">
					{formatUpdated(row.updatedAt)}
				</span>
			),
		},
	];

	const trialColumns: DataTableColumn<TrialRow>[] = [
		{
			id: "setting",
			header: m["admin.premium.trial.setting"](),
			cell: (row) => <span className="font-medium">{row.label()}</span>,
		},
		{
			id: "value",
			header: m["admin.premium.trial.value"](),
			className: "w-40",
			cell: (row) => (
				<Input
					aria-label={row.label()}
					type="number"
					min={row.min}
					max={row.max}
					className="w-28"
					value={trial[row.id]}
					onChange={(event) => updateTrial(row.id, Number(event.target.value))}
				/>
			),
		},
	];

	return (
		<div className="flex flex-col gap-6">
			<section className="flex flex-col gap-3">
				<div>
					<h2 className="text-sm font-bold">
						{m["admin.premium.costs.title"]()}
					</h2>
					<p className="text-xs text-muted-foreground">
						{m["admin.premium.costs.hint"]()}
					</p>
				</div>
				<DataTable
					columns={costColumns}
					rows={costRows}
					rowKey={(row) => row.action}
				/>
				<div className="flex justify-end">
					<SubmitButton
						type="button"
						pending={saveCosts.isPending}
						onClick={() => {
							saveCosts.mutate(
								{ costs },
								{
									onSuccess: () =>
										toast.success(m["admin.premium.costs.saved"]()),
									onError: () => toast.error(m["admin.toast.error"]()),
								},
							);
						}}
					>
						{m["admin.premium.costs.save"]()}
					</SubmitButton>
				</div>
			</section>

			<section className="flex flex-col gap-3">
				<div>
					<h2 className="text-sm font-bold">
						{m["admin.premium.trial.title"]()}
					</h2>
					<p className="text-xs text-muted-foreground">
						{m["admin.premium.trial.description"]()}
					</p>
				</div>
				<DataTable
					columns={trialColumns}
					rows={TRIAL_ROWS}
					rowKey={(row) => row.id}
				/>
				<div className="flex items-center justify-between gap-3">
					{trialQuery.data?.updatedAt ? (
						<p className="text-xs text-muted-foreground">
							{m["admin.premium.trial.updated"]({
								date: formatUpdated(trialQuery.data.updatedAt),
							})}
						</p>
					) : (
						<span />
					)}
					<SubmitButton
						type="button"
						pending={saveTrial.isPending}
						onClick={() => {
							saveTrial.mutate(trial, {
								onSuccess: () =>
									toast.success(m["admin.premium.trial.saved"]()),
								onError: () => toast.error(m["admin.toast.error"]()),
							});
						}}
					>
						{m["admin.premium.trial.save"]()}
					</SubmitButton>
				</div>
			</section>
		</div>
	);
}
