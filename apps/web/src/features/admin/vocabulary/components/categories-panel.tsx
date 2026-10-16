import type { VocabularyCategoryResponse } from "@engflex/contracts";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "#/components/ui/button";
import { Field, FieldError, FieldLabel } from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import { Textarea } from "#/components/ui/textarea";
import { ConfirmDeleteDialog } from "#/features/admin/components/confirm-delete-dialog";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import { ResourceFormDialog } from "#/features/admin/components/resource-form-dialog";
import {
	useAdminVocabularyCategories,
	useCreateVocabularyCategory,
	useDeleteVocabularyCategory,
	useUpdateVocabularyCategory,
} from "#/features/admin/vocabulary/queries";
import { m } from "#/paraglide/messages";

interface CategoryFormState {
	name: string;
	description: string;
}

const EMPTY_FORM: CategoryFormState = { name: "", description: "" };

export function CategoriesPanel() {
	const [page, setPage] = useState(1);
	const pageSize = 20;
	const query = useAdminVocabularyCategories(page, pageSize);
	const createMutation = useCreateVocabularyCategory();
	const updateMutation = useUpdateVocabularyCategory();
	const deleteMutation = useDeleteVocabularyCategory();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<VocabularyCategoryResponse | null>(
		null,
	);
	const [form, setForm] = useState<CategoryFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] =
		useState<VocabularyCategoryResponse | null>(null);

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(category: VocabularyCategoryResponse) {
		setEditing(category);
		setForm({ name: category.name, description: category.description });
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.name.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		const payload = { name: form.name.trim(), description: form.description };
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

	const columns: DataTableColumn<VocabularyCategoryResponse>[] = [
		{ id: "name", header: "Name", cell: (row) => row.name },
		{
			id: "description",
			header: "Description",
			cell: (row) => row.description || "—",
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
				onSubmit={submit}
				pending={createMutation.isPending || updateMutation.isPending}
			>
				<Field data-invalid={Boolean(error)}>
					<FieldLabel htmlFor="category-name">Name</FieldLabel>
					<Input
						id="category-name"
						value={form.name}
						onChange={(event) =>
							setForm((current) => ({ ...current, name: event.target.value }))
						}
					/>
				</Field>
				<Field>
					<FieldLabel htmlFor="category-description">Description</FieldLabel>
					<Textarea
						id="category-description"
						rows={2}
						value={form.description}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								description: event.target.value,
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
