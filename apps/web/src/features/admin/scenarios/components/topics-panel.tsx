import type { ScenarioTopic } from "@engflex/contracts";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "#/components/ui/button";
import { Field, FieldError, FieldLabel } from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import { ConfirmDeleteDialog } from "#/features/admin/components/confirm-delete-dialog";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import { ResourceFormDialog } from "#/features/admin/components/resource-form-dialog";
import {
	useAdminScenarioTopics,
	useCreateScenarioTopic,
	useDeleteScenarioTopic,
	useUpdateScenarioTopic,
} from "#/features/admin/scenarios/queries";
import { m } from "#/paraglide/messages";

interface TopicFormState {
	slug: string;
	name: string;
	shortName: string;
	position: string;
}

const EMPTY_FORM: TopicFormState = {
	slug: "",
	name: "",
	shortName: "",
	position: "0",
};

export function TopicsPanel() {
	const query = useAdminScenarioTopics();
	const createMutation = useCreateScenarioTopic();
	const updateMutation = useUpdateScenarioTopic();
	const deleteMutation = useDeleteScenarioTopic();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<ScenarioTopic | null>(null);
	const [form, setForm] = useState<TopicFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] = useState<ScenarioTopic | null>(null);

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(topic: ScenarioTopic) {
		setEditing(topic);
		setForm({
			slug: topic.slug,
			name: topic.name,
			shortName: topic.shortName,
			position: String(topic.position),
		});
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.slug.trim() || !form.name.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		const payload = {
			slug: form.slug.trim(),
			name: form.name.trim(),
			shortName: form.shortName,
			position: Number(form.position) || 0,
		};
		if (editing) {
			updateMutation.mutate(
				{ id: editing.id, payload },
				{
					onSuccess: () => {
						toast.success(m["admin.toast.updated"]());
						setDialogOpen(false);
					},
					onError: () => toast.error(m["admin.toast.error"]()),
				},
			);
		} else {
			createMutation.mutate(payload, {
				onSuccess: () => {
					toast.success(m["admin.toast.created"]());
					setDialogOpen(false);
				},
				onError: () => toast.error(m["admin.toast.error"]()),
			});
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

	const columns: DataTableColumn<ScenarioTopic>[] = [
		{ id: "position", header: "#", cell: (row) => row.position },
		{ id: "name", header: "Name", cell: (row) => row.name },
		{ id: "shortName", header: "Short name", cell: (row) => row.shortName },
		{ id: "slug", header: "Slug", cell: (row) => row.slug },
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
				rows={query.data ?? []}
				rowKey={(row) => row.id}
				isLoading={query.isPending}
			/>
			<ResourceFormDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				title={
					editing ? m["admin.actions.edit"]() : m["admin.actions.create"]()
				}
				onSubmit={submit}
				pending={createMutation.isPending || updateMutation.isPending}
			>
				<Field data-invalid={Boolean(error)}>
					<FieldLabel htmlFor="topic-name">Name</FieldLabel>
					<Input
						id="topic-name"
						value={form.name}
						onChange={(event) =>
							setForm((current) => ({ ...current, name: event.target.value }))
						}
					/>
				</Field>
				<div className="grid gap-3 sm:grid-cols-2">
					<Field data-invalid={Boolean(error)}>
						<FieldLabel htmlFor="topic-slug">Slug</FieldLabel>
						<Input
							id="topic-slug"
							value={form.slug}
							onChange={(event) =>
								setForm((current) => ({ ...current, slug: event.target.value }))
							}
						/>
					</Field>
					<Field>
						<FieldLabel htmlFor="topic-position">Position</FieldLabel>
						<Input
							id="topic-position"
							type="number"
							value={form.position}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									position: event.target.value,
								}))
							}
						/>
					</Field>
				</div>
				<Field>
					<FieldLabel htmlFor="topic-short-name">Short name</FieldLabel>
					<Input
						id="topic-short-name"
						value={form.shortName}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								shortName: event.target.value,
							}))
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
