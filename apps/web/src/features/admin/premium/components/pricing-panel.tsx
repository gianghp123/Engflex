import type { AccessAction } from "@engflex/contracts";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { SubmitButton } from "#/components/common/submit-button";
import { Field, FieldLabel } from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import { Skeleton } from "#/components/ui/skeleton";
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
			setCosts(
				costsQuery.data.map((row) => ({
					action: row.action,
					credits: row.credits,
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

	return (
		<div className="grid gap-4 lg:grid-cols-2">
			<section className="surface-card flex flex-col gap-3 p-4 sm:p-5">
				<h2 className="text-sm font-bold">
					{m["admin.premium.costs.title"]()}
				</h2>
				<p className="text-xs text-muted-foreground">
					{m["admin.premium.costs.hint"]()}
				</p>
				{PAID_ACTIONS.map((action) => {
					const row = costs.find((item) => item.action === action);
					return (
						<Field key={action}>
							<FieldLabel htmlFor={`cost-${action}`}>
								{ACTION_LABELS[action]()}
							</FieldLabel>
							<Input
								id={`cost-${action}`}
								type="number"
								min={0}
								max={100000}
								value={row?.credits ?? 0}
								onChange={(event) =>
									setCosts((current) =>
										current.map((item) =>
											item.action === action
												? { ...item, credits: Number(event.target.value) }
												: item,
										),
									)
								}
							/>
						</Field>
					);
				})}
				<SubmitButton
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
			</section>

			<section className="surface-card flex flex-col gap-3 p-4 sm:p-5">
				<h2 className="text-sm font-bold">
					{m["admin.premium.trial.title"]()}
				</h2>
				<p className="text-xs text-muted-foreground">
					{m["admin.premium.trial.description"]()}
				</p>
				<Field>
					<FieldLabel htmlFor="trial-days">
						{m["admin.premium.trial.days"]()}
					</FieldLabel>
					<Input
						id="trial-days"
						type="number"
						min={1}
						max={365}
						value={trial.days}
						onChange={(event) =>
							setTrial({ ...trial, days: Number(event.target.value) })
						}
					/>
				</Field>
				<Field>
					<FieldLabel htmlFor="trial-credits">
						{m["admin.premium.trial.credits"]()}
					</FieldLabel>
					<Input
						id="trial-credits"
						type="number"
						min={0}
						max={100000}
						value={trial.credits}
						onChange={(event) =>
							setTrial({ ...trial, credits: Number(event.target.value) })
						}
					/>
				</Field>
				<Field>
					<FieldLabel htmlFor="trial-voice">
						{m["admin.premium.trial.voiceSessions"]()}
					</FieldLabel>
					<Input
						id="trial-voice"
						type="number"
						min={0}
						max={100000}
						value={trial.voiceSessions}
						onChange={(event) =>
							setTrial({
								...trial,
								voiceSessions: Number(event.target.value),
							})
						}
					/>
				</Field>
				<SubmitButton
					pending={saveTrial.isPending}
					onClick={() => {
						saveTrial.mutate(trial, {
							onSuccess: () => toast.success(m["admin.premium.trial.saved"]()),
							onError: () => toast.error(m["admin.toast.error"]()),
						});
					}}
				>
					{m["admin.premium.trial.save"]()}
				</SubmitButton>
			</section>
		</div>
	);
}
