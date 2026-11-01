import type { VideoCategoryResponse } from "@engflex/contracts";
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
	useAdminVideoCategories,
	useCreateVideoCategory,
	useDeleteVideoCategory,
	useUpdateVideoCategory,
} from "#/features/admin/videos/queries";
import { m } from "#/paraglide/messages";

interface CategoryFormState {
	slug: string;
	name: string;
}

const EMPTY_FORM: CategoryFormState = { slug: "", name: "" };

export function VideoCategoriesPanel() {
	const [page, setPage] = useState(1);
	const pageSize = 20;
	const query = useAdminVideoCategories(page, pageSize);
	const createMutation = useCreateVideoCategory();
	const updateMutation = useUpdateVideoCategory();
	const deleteMutation = useDeleteVideoCategory();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<VideoCategoryResponse | null>(null);
	const [form, setForm] = useState<CategoryFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] =
		useState<VideoCategoryResponse | null>(null);

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(category: VideoCategoryResponse) {
		setEditing(category);
		setForm({ slug: category.slug, name: category.name });
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.slug.trim() || !form.name.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		const payload = { slug: form.slug.trim(), name: form.name.trim() };
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

	const columns: DataTableColumn<VideoCategoryResponse>[] = [
		{ id: "name", header: "Name", cell: (row) => row.name },
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
				<Field data-invalid={Boolean(error)}>
					<FieldLabel htmlFor="category-slug">Slug</FieldLabel>
					<Input
						id="category-slug"
						value={form.slug}
						onChange={(event) =>
							setForm((current) => ({ ...current, slug: event.target.value }))
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
