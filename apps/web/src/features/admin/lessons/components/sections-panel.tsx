import type { CEFR, Section } from "@engflex/contracts";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Button } from "#/components/ui/button";
import { Field, FieldError, FieldLabel } from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { Textarea } from "#/components/ui/textarea";
import { ConfirmDeleteDialog } from "#/features/admin/components/confirm-delete-dialog";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import { ResourceFormDialog } from "#/features/admin/components/resource-form-dialog";
import {
	useAdminLessonSections,
	useCreateLessonSection,
	useDeleteLessonSection,
	useUpdateLessonSection,
} from "#/features/admin/lessons/queries";
import { m } from "#/paraglide/messages";

const CEFR_OPTIONS: CEFR[] = ["A1", "A2", "B1", "B2", "C1", "C2"];

interface SectionFormState {
	slug: string;
	title: string;
	description: string;
	cefrBand: CEFR;
	position: string;
}

const EMPTY_FORM: SectionFormState = {
	slug: "",
	title: "",
	description: "",
	cefrBand: "A1",
	position: "0",
};

export function SectionsPanel() {
	const query = useAdminLessonSections();
	const createMutation = useCreateLessonSection();
	const updateMutation = useUpdateLessonSection();
	const deleteMutation = useDeleteLessonSection();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<Section | null>(null);
	const [form, setForm] = useState<SectionFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] = useState<Section | null>(null);

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(section: Section) {
		setEditing(section);
		setForm({
			slug: section.slug,
			title: section.title,
			description: section.description,
			cefrBand: section.cefrBand,
			position: String(section.position),
		});
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.slug.trim() || !form.title.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		const payload = {
			slug: form.slug.trim(),
			title: form.title.trim(),
			description: form.description,
			cefrBand: form.cefrBand,
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

	const columns: DataTableColumn<Section>[] = [
		{ id: "position", header: "#", cell: (row) => row.position },
		{ id: "title", header: "Title", cell: (row) => row.title },
		{ id: "cefrBand", header: "CEFR band", cell: (row) => row.cefrBand },
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
					<FieldLabel htmlFor="section-title">Title</FieldLabel>
					<Input
						id="section-title"
						value={form.title}
						onChange={(event) =>
							setForm((current) => ({ ...current, title: event.target.value }))
						}
					/>
				</Field>
				<div className="grid gap-3 sm:grid-cols-3">
					<Field data-invalid={Boolean(error)}>
						<FieldLabel htmlFor="section-slug">Slug</FieldLabel>
						<Input
							id="section-slug"
							value={form.slug}
							onChange={(event) =>
								setForm((current) => ({ ...current, slug: event.target.value }))
							}
						/>
					</Field>
					<Field>
						<FieldLabel>CEFR band</FieldLabel>
						<Select
							value={form.cefrBand}
							onValueChange={(value) =>
								setForm((current) => ({ ...current, cefrBand: value as CEFR }))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								{CEFR_OPTIONS.map((option) => (
									<SelectItem key={option} value={option}>
										{option}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
					<Field>
						<FieldLabel htmlFor="section-position">Position</FieldLabel>
						<Input
							id="section-position"
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
					<FieldLabel htmlFor="section-description">Description</FieldLabel>
					<Textarea
						id="section-description"
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
