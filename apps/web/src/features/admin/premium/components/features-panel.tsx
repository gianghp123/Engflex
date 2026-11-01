import type { PremiumFeature } from "@engflex/contracts";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Field, FieldError, FieldLabel } from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import { Switch } from "#/components/ui/switch";
import { Textarea } from "#/components/ui/textarea";
import { ConfirmDeleteDialog } from "#/features/admin/components/confirm-delete-dialog";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import { ResourceFormDialog } from "#/features/admin/components/resource-form-dialog";
import {
	useAdminPremiumFeatures,
	useCreatePremiumFeature,
	useDeletePremiumFeature,
	useUpdatePremiumFeature,
} from "#/features/admin/premium/queries";
import { m } from "#/paraglide/messages";

interface FeatureFormState {
	key: string;
	label: string;
	description: string;
	isActive: boolean;
}

const EMPTY_FORM: FeatureFormState = {
	key: "",
	label: "",
	description: "",
	isActive: true,
};

export function FeaturesPanel() {
	const [page, setPage] = useState(1);
	const pageSize = 20;
	const query = useAdminPremiumFeatures(page, pageSize);
	const createMutation = useCreatePremiumFeature();
	const updateMutation = useUpdatePremiumFeature();
	const deleteMutation = useDeletePremiumFeature();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<PremiumFeature | null>(null);
	const [form, setForm] = useState<FeatureFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] = useState<PremiumFeature | null>(null);

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(feature: PremiumFeature) {
		setEditing(feature);
		setForm({
			key: feature.key,
			label: feature.label,
			description: feature.description,
			isActive: feature.isActive,
		});
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.key.trim() || !form.label.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		if (editing) {
			updateMutation.mutate(
				{
					id: editing.id,
					payload: {
						label: form.label.trim(),
						description: form.description.trim(),
						isActive: form.isActive,
					},
				},
				{
					onSuccess: () => {
						toast.success(m["admin.toast.updated"]());
						setDialogOpen(false);
					},
					onError: () => toast.error(m["admin.toast.error"]()),
				},
			);
		} else {
			createMutation.mutate(
				{
					key: form.key.trim(),
					label: form.label.trim(),
					description: form.description.trim(),
					isActive: form.isActive,
				},
				{
					onSuccess: () => {
						toast.success(m["admin.toast.created"]());
						setDialogOpen(false);
					},
					onError: () => toast.error(m["admin.toast.error"]()),
				},
			);
		}
	}

	function confirmDelete() {
		if (!deleteTarget) return;
		deleteMutation.mutate(deleteTarget.id, {
			onSuccess: () => {
				toast.success(m["admin.toast.deleted"]());
				setDeleteTarget(null);
			},
			onError: () => toast.error(m["admin.toast.error"]()),
		});
	}

	const columns: DataTableColumn<PremiumFeature>[] = [
		{
			id: "key",
			header: m["admin.premium.features.key"](),
			cell: (row) => row.key,
		},
		{
			id: "label",
			header: m["admin.premium.features.label"](),
			cell: (row) => row.label,
		},
		{
			id: "isActive",
			header: m["admin.table.status"](),
			cell: (row) => (
				<Badge variant={row.isActive ? "default" : "secondary"}>
					{row.isActive
						? m["admin.premium.features.gated"]()
						: m["admin.premium.features.ungated"]()}
				</Badge>
			),
		},
		{
			id: "actions",
			header: m["admin.table.actions"](),
			cell: (row) => (
				<div className="flex items-center gap-2">
					<Button
						type="button"
						variant="outline"
						size="sm"
						onClick={() => openEdit(row)}
					>
						{m["admin.actions.edit"]()}
					</Button>
					<Button
						type="button"
						variant="ghost"
						size="sm"
						onClick={() => setDeleteTarget(row)}
					>
						{m["admin.actions.delete"]()}
					</Button>
				</div>
			),
		},
	];

	return (
		<div className="flex flex-col gap-4">
			<div className="flex justify-end">
				<Button type="button" onClick={openCreate}>
					<Plus data-icon="inline-start" />
					{m["admin.actions.create"]()}
				</Button>
			</div>
			<DataTable
				columns={columns}
				rows={query.data?.items ?? []}
				rowKey={(row) => row.id}
				isLoading={query.isPending}
				pagination={query.data?.pagination}
				onPageChange={setPage}
			/>
			<ResourceFormDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				title={
					editing ? m["admin.actions.edit"]() : m["admin.actions.create"]()
				}
				description={m["admin.premium.features.formHint"]()}
				onSubmit={submit}
				pending={createMutation.isPending || updateMutation.isPending}
			>
				<Field data-invalid={Boolean(error)}>
					<FieldLabel htmlFor="feature-key">
						{m["admin.premium.features.key"]()}
					</FieldLabel>
					<Input
						id="feature-key"
						disabled={Boolean(editing)}
						value={form.key}
						onChange={(event) =>
							setForm((current) => ({ ...current, key: event.target.value }))
						}
					/>
				</Field>
				<Field data-invalid={Boolean(error)}>
					<FieldLabel htmlFor="feature-label">
						{m["admin.premium.features.label"]()}
					</FieldLabel>
					<Input
						id="feature-label"
						value={form.label}
						onChange={(event) =>
							setForm((current) => ({ ...current, label: event.target.value }))
						}
					/>
				</Field>
				<Field>
					<FieldLabel htmlFor="feature-description">
						{m["admin.premium.features.description"]()}
					</FieldLabel>
					<Textarea
						id="feature-description"
						value={form.description}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								description: event.target.value,
							}))
						}
					/>
				</Field>
				<Field className="flex-row items-center justify-between">
					<FieldLabel htmlFor="feature-active">
						{m["admin.premium.features.gated"]()}
					</FieldLabel>
					<Switch
						id="feature-active"
						checked={form.isActive}
						onCheckedChange={(checked) =>
							setForm((current) => ({ ...current, isActive: checked }))
						}
					/>
				</Field>
				{error ? <FieldError>{error}</FieldError> : null}
			</ResourceFormDialog>
			<ConfirmDeleteDialog
				open={Boolean(deleteTarget)}
				onOpenChange={(open) => !open && setDeleteTarget(null)}
				onConfirm={confirmDelete}
				pending={deleteMutation.isPending}
			/>
		</div>
	);
}
