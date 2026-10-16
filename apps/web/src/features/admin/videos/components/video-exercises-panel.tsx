import type { CEFR, VideoExerciseResponse } from "@engflex/contracts";
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
import { StatusBadge } from "#/features/admin/components/status-badge";
import {
	useAdminVideoCategories,
	useAdminVideoExercises,
	useCreateVideoExercise,
	useDeleteVideoExercise,
	useUpdateVideoExercise,
} from "#/features/admin/videos/queries";
import { m } from "#/paraglide/messages";

const CEFR_OPTIONS: CEFR[] = ["A1", "A2", "B1", "B2", "C1", "C2"];

interface ExerciseFormState {
	categoryId: string;
	title: string;
	description: string;
	videoUrl: string;
	youtubeUrl: string;
	thumbnailUrl: string;
	cefrLevel: CEFR | "";
	duration: string;
	status: "draft" | "published";
}

const EMPTY_FORM: ExerciseFormState = {
	categoryId: "",
	title: "",
	description: "",
	videoUrl: "",
	youtubeUrl: "",
	thumbnailUrl: "",
	cefrLevel: "",
	duration: "",
	status: "draft",
};

export function VideoExercisesPanel() {
	const [page, setPage] = useState(1);
	const pageSize = 20;
	const query = useAdminVideoExercises(page, pageSize);
	const categoriesQuery = useAdminVideoCategories(1, 100);
	const createMutation = useCreateVideoExercise();
	const updateMutation = useUpdateVideoExercise();
	const deleteMutation = useDeleteVideoExercise();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<VideoExerciseResponse | null>(null);
	const [form, setForm] = useState<ExerciseFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] =
		useState<VideoExerciseResponse | null>(null);

	const categories = categoriesQuery.data?.items ?? [];

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(exercise: VideoExerciseResponse) {
		setEditing(exercise);
		setForm({
			categoryId: exercise.categoryId ?? "",
			title: exercise.title,
			description: exercise.description,
			videoUrl: exercise.videoUrl,
			youtubeUrl: "",
			thumbnailUrl: exercise.thumbnailUrl,
			cefrLevel: exercise.cefrLevel ?? "",
			duration: String(exercise.duration),
			status: exercise.status === "published" ? "published" : "draft",
		});
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.title.trim() || !form.videoUrl.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		const base = {
			categoryId: form.categoryId || undefined,
			title: form.title.trim(),
			description: form.description,
			videoUrl: form.videoUrl.trim(),
			youtubeUrl: form.youtubeUrl,
			thumbnailUrl: form.thumbnailUrl,
			cefrLevel: form.cefrLevel || undefined,
			duration: Number(form.duration) || 0,
		};
		if (editing) {
			updateMutation.mutate(
				{ id: editing.id, payload: { ...base, status: form.status } },
				{
					onSuccess: () => {
						toast.success(m["admin.toast.updated"]());
						setDialogOpen(false);
					},
					onError: () => toast.error(m["admin.toast.error"]()),
				},
			);
		} else {
			createMutation.mutate(base, {
				onSuccess: () => {
					toast.success(m["admin.toast.created"]());
					setDialogOpen(false);
				},
				onError: () => toast.error(m["admin.toast.error"]()),
			});
		}
	}

	function togglePublish(exercise: VideoExerciseResponse) {
		const nextStatus = exercise.status === "published" ? "draft" : "published";
		updateMutation.mutate(
			{
				id: exercise.id,
				payload: {
					categoryId: exercise.categoryId,
					title: exercise.title,
					description: exercise.description,
					videoUrl: exercise.videoUrl,
					youtubeUrl: "",
					thumbnailUrl: exercise.thumbnailUrl,
					cefrLevel: exercise.cefrLevel,
					duration: exercise.duration,
					status: nextStatus,
				},
			},
			{
				onSuccess: () =>
					toast.success(
						nextStatus === "published"
							? m["admin.toast.published"]()
							: m["admin.toast.unpublished"](),
					),
				onError: () => toast.error(m["admin.toast.error"]()),
			},
		);
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

	const columns: DataTableColumn<VideoExerciseResponse>[] = [
		{ id: "title", header: "Title", cell: (row) => row.title },
		{
			id: "category",
			header: "Category",
			cell: (row) => row.category?.name ?? "—",
		},
		{ id: "cefr", header: "CEFR", cell: (row) => row.cefrLevel ?? "—" },
		{
			id: "status",
			header: m["admin.table.status"](),
			cell: (row) => (
				<StatusBadge
					status={row.status === "published" ? "published" : "draft"}
				/>
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
						variant="outline"
						size="sm"
						onClick={() => togglePublish(row)}
					>
						{row.status === "published"
							? m["admin.actions.unpublish"]()
							: m["admin.actions.publish"]()}
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
					<FieldLabel htmlFor="exercise-title">Title</FieldLabel>
					<Input
						id="exercise-title"
						value={form.title}
						onChange={(event) =>
							setForm((current) => ({ ...current, title: event.target.value }))
						}
					/>
				</Field>
				<div className="grid gap-3 sm:grid-cols-2">
					<Field>
						<FieldLabel>Category</FieldLabel>
						<Select
							value={form.categoryId || "none"}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									categoryId: value === "none" ? "" : value,
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">—</SelectItem>
								{categories.map((category) => (
									<SelectItem key={category.id} value={category.id}>
										{category.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
					<Field>
						<FieldLabel>CEFR</FieldLabel>
						<Select
							value={form.cefrLevel || "none"}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									cefrLevel: value === "none" ? "" : (value as CEFR),
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">—</SelectItem>
								{CEFR_OPTIONS.map((option) => (
									<SelectItem key={option} value={option}>
										{option}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
				</div>
				<Field data-invalid={Boolean(error)}>
					<FieldLabel htmlFor="exercise-video-url">Video URL</FieldLabel>
					<Input
						id="exercise-video-url"
						value={form.videoUrl}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								videoUrl: event.target.value,
							}))
						}
					/>
				</Field>
				<div className="grid gap-3 sm:grid-cols-2">
					<Field>
						<FieldLabel htmlFor="exercise-thumbnail">Thumbnail URL</FieldLabel>
						<Input
							id="exercise-thumbnail"
							value={form.thumbnailUrl}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									thumbnailUrl: event.target.value,
								}))
							}
						/>
					</Field>
					<Field>
						<FieldLabel htmlFor="exercise-duration">Duration (s)</FieldLabel>
						<Input
							id="exercise-duration"
							type="number"
							value={form.duration}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									duration: event.target.value,
								}))
							}
						/>
					</Field>
				</div>
				<Field>
					<FieldLabel htmlFor="exercise-description">Description</FieldLabel>
					<Textarea
						id="exercise-description"
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
				{editing ? (
					<Field>
						<FieldLabel>{m["admin.table.status"]()}</FieldLabel>
						<Select
							value={form.status}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									status: value as "draft" | "published",
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="draft">
									{m["admin.status.draft"]()}
								</SelectItem>
								<SelectItem value="published">
									{m["admin.status.published"]()}
								</SelectItem>
							</SelectContent>
						</Select>
					</Field>
				) : null}
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
