import type { AccessAction } from "@engflex/contracts";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { SubmitButton } from "#/components/common/submit-button";
import { Input } from "#/components/ui/input";
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

const TRIAL_ROWS: {
	id: TrialKey;
	label: () => string;
	min: number;
	max: number;
}[] = [
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
				<div className="overflow-hidden rounded-md border bg-card text-card-foreground">
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>{m["admin.premium.costs.action"]()}</TableHead>
								<TableHead className="w-40">
									{m["admin.premium.costs.credits"]()}
								</TableHead>
								<TableHead className="w-56">
									{m["admin.premium.costs.updated"]()}
								</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{costRows.map((row) => (
								<TableRow key={row.action}>
									<TableCell className="font-medium">
										{ACTION_LABELS[row.action]()}
									</TableCell>
									<TableCell>
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
									</TableCell>
									<TableCell className="text-muted-foreground">
										{formatUpdated(row.updatedAt)}
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				</div>
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
				<div className="overflow-hidden rounded-md border bg-card text-card-foreground">
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>{m["admin.premium.trial.setting"]()}</TableHead>
								<TableHead className="w-40">
									{m["admin.premium.trial.value"]()}
								</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{TRIAL_ROWS.map((row) => (
								<TableRow key={row.id}>
									<TableCell className="font-medium">{row.label()}</TableCell>
									<TableCell>
										<Input
											aria-label={row.label()}
											type="number"
											min={row.min}
											max={row.max}
											className="w-28"
											value={trial[row.id]}
											onChange={(event) =>
												updateTrial(row.id, Number(event.target.value))
											}
										/>
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				</div>
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
