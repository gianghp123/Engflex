import type { Plan } from "@engflex/contracts";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Field, FieldLabel } from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import { Switch } from "#/components/ui/switch";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import { ResourceFormDialog } from "#/features/admin/components/resource-form-dialog";
import { useAdminPlans, useUpdatePlan } from "#/features/admin/premium/queries";
import { m } from "#/paraglide/messages";

interface PlanFormState {
	months: string;
	amount: string;
	currency: string;
	isActive: boolean;
}

function planLabel(plan: Plan["plan"]): string {
	switch (plan) {
		case "1m":
			return m["payments.pricing.plan.1m"]();
		case "3m":
			return m["payments.pricing.plan.3m"]();
		case "12m":
			return m["payments.pricing.plan.12m"]();
	}
}

export function PlansPanel() {
	const query = useAdminPlans();
	const updateMutation = useUpdatePlan();

	const [editing, setEditing] = useState<Plan | null>(null);
	const [form, setForm] = useState<PlanFormState | null>(null);

	function openEdit(plan: Plan) {
		setEditing(plan);
		setForm({
			months: String(plan.months),
			amount: String(plan.amount),
			currency: plan.currency,
			isActive: plan.isActive,
		});
	}

	function submit() {
		if (!editing || !form) return;
		updateMutation.mutate(
			{
				plan: editing.plan,
				payload: {
					months: Number(form.months),
					amount: Number(form.amount),
					currency: form.currency,
					isActive: form.isActive,
				},
			},
			{
				onSuccess: () => {
					toast.success(m["admin.toast.updated"]());
					setEditing(null);
				},
				onError: () => toast.error(m["admin.toast.error"]()),
			},
		);
	}

	const columns: DataTableColumn<Plan>[] = [
		{
			id: "plan",
			header: m["admin.premium.plans.plan"](),
			cell: (row) => planLabel(row.plan),
		},
		{
			id: "months",
			header: m["admin.premium.plans.months"](),
			cell: (row) => row.months,
		},
		{
			id: "amount",
			header: m["admin.premium.plans.amount"](),
			cell: (row) =>
				new Intl.NumberFormat("vi-VN", {
					style: "currency",
					currency: row.currency,
				}).format(row.amount),
		},
		{
			id: "isActive",
			header: m["admin.table.status"](),
			cell: (row) => (
				<Badge variant={row.isActive ? "default" : "secondary"}>
					{row.isActive
						? m["admin.premium.plans.active"]()
						: m["admin.premium.plans.inactive"]()}
				</Badge>
			),
		},
		{
			id: "actions",
			header: m["admin.table.actions"](),
			cell: (row) => (
				<Button
					type="button"
					variant="outline"
					size="sm"
					onClick={() => openEdit(row)}
				>
					{m["admin.actions.edit"]()}
				</Button>
			),
		},
	];

	return (
		<div className="flex flex-col gap-4">
			<DataTable
				columns={columns}
				rows={query.data ?? []}
				rowKey={(row) => row.plan}
				isLoading={query.isPending}
			/>
			<ResourceFormDialog
				open={Boolean(editing)}
				onOpenChange={(open) => !open && setEditing(null)}
				title={editing ? planLabel(editing.plan) : ""}
				onSubmit={submit}
				pending={updateMutation.isPending}
			>
				{form && (
					<>
						<Field>
							<FieldLabel htmlFor="plan-months">
								{m["admin.premium.plans.months"]()}
							</FieldLabel>
							<Input
								id="plan-months"
								type="number"
								min={1}
								max={36}
								value={form.months}
								onChange={(event) =>
									setForm((current) =>
										current
											? { ...current, months: event.target.value }
											: current,
									)
								}
							/>
						</Field>
						<Field>
							<FieldLabel htmlFor="plan-amount">
								{m["admin.premium.plans.amount"]()}
							</FieldLabel>
							<Input
								id="plan-amount"
								type="number"
								min={0}
								value={form.amount}
								onChange={(event) =>
									setForm((current) =>
										current
											? { ...current, amount: event.target.value }
											: current,
									)
								}
							/>
						</Field>
						<Field>
							<FieldLabel htmlFor="plan-currency">
								{m["admin.premium.plans.currency"]()}
							</FieldLabel>
							<Input
								id="plan-currency"
								value={form.currency}
								onChange={(event) =>
									setForm((current) =>
										current
											? {
													...current,
													currency: event.target.value.toUpperCase(),
												}
											: current,
									)
								}
							/>
						</Field>
						<Field className="flex-row items-center justify-between">
							<FieldLabel htmlFor="plan-active">
								{m["admin.premium.plans.active"]()}
							</FieldLabel>
							<Switch
								id="plan-active"
								checked={form.isActive}
								onCheckedChange={(checked) =>
									setForm((current) =>
										current ? { ...current, isActive: checked } : current,
									)
								}
							/>
						</Field>
					</>
				)}
			</ResourceFormDialog>
		</div>
	);
}
