import type { ActivityType, CEFR, Lesson } from "@engflex/contracts";
import { Plus } from "lucide-react";
import { useMemo, useState } from "react";
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
import { ActivitiesDialog } from "#/features/admin/lessons/components/activities-dialog";
import {
	useAdminLessonHub,
	useAdminLessonSections,
	useCreateLesson,
	useDeleteLesson,
	useUpdateLesson,
} from "#/features/admin/lessons/queries";
import { m } from "#/paraglide/messages";

const CEFR_OPTIONS: CEFR[] = ["A1", "A2", "B1", "B2", "C1", "C2"];
const SKILL_OPTIONS: ActivityType[] = [
	"reading",
	"listening",
	"writing",
	"speaking",
];

interface LessonRow {
	lesson: Lesson;
	sectionTitle: string;
}

interface LessonFormState {
	slug: string;
	title: string;
	sectionId: string;
	cefrLevel: CEFR;
	description: string;
	estimatedDurationMin: string;
	coverImageUrl: string;
	skill: ActivityType | "";
	status: "draft" | "published";
}

const EMPTY_FORM: LessonFormState = {
	slug: "",
	title: "",
	sectionId: "",
	cefrLevel: "A1",
	description: "",
	estimatedDurationMin: "",
	coverImageUrl: "",
	skill: "",
	status: "draft",
};

export function LessonsPanel() {
	const hubQuery = useAdminLessonHub();
	const sectionsQuery = useAdminLessonSections();
	const createMutation = useCreateLesson();
	const updateMutation = useUpdateLesson();
	const deleteMutation = useDeleteLesson();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<Lesson | null>(null);
	const [form, setForm] = useState<LessonFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] = useState<Lesson | null>(null);
	const [activitiesLesson, setActivitiesLesson] = useState<Lesson | null>(null);

	const sections = sectionsQuery.data ?? [];

	const rows = useMemo<LessonRow[]>(() => {
		const sectionsHub = hubQuery.data ?? [];
		return sectionsHub.flatMap((section) =>
			section.units.map((lesson) => ({ lesson, sectionTitle: section.title })),
		);
	}, [hubQuery.data]);

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(lesson: Lesson) {
		setEditing(lesson);
		setForm({
			slug: lesson.slug,
			title: lesson.title,
			sectionId: lesson.section?.id ?? "",
			cefrLevel: lesson.cefrLevel,
			description: lesson.description,
			estimatedDurationMin:
				lesson.details.estimatedDurationMin != null
					? String(lesson.details.estimatedDurationMin)
					: "",
			coverImageUrl: lesson.details.coverImageUrl,
			skill: lesson.details.skill ?? "",
			status: lesson.status === "published" ? "published" : "draft",
		});
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.slug.trim() || !form.title.trim() || !form.sectionId) {
			setError(m["admin.toast.error"]());
			return;
		}
		const details = {
			estimatedDurationMin: form.estimatedDurationMin
				? Number(form.estimatedDurationMin)
				: undefined,
			coverImageUrl: form.coverImageUrl,
			skill: form.skill || undefined,
		};
		const base = {
			slug: form.slug.trim(),
			title: form.title.trim(),
			sectionId: form.sectionId,
			cefrLevel: form.cefrLevel,
			description: form.description,
		};
		if (editing) {
			updateMutation.mutate(
				{ id: editing.id, payload: { ...base, details, status: form.status } },
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
				{ ...base, details },
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

	function togglePublish(lesson: Lesson) {
		const nextStatus = lesson.status === "published" ? "draft" : "published";
		updateMutation.mutate(
			{ id: lesson.id, payload: { status: nextStatus } },
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

	const columns: DataTableColumn<LessonRow>[] = [
		{ id: "title", header: "Title", cell: (row) => row.lesson.title },
		{ id: "section", header: "Section", cell: (row) => row.sectionTitle },
		{ id: "cefr", header: "CEFR", cell: (row) => row.lesson.cefrLevel },
		{
			id: "status",
			header: m["admin.table.status"](),
			cell: (row) => (
				<StatusBadge
					status={row.lesson.status === "published" ? "published" : "draft"}
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
						onClick={() => setActivitiesLesson(row.lesson)}
					>
						Activities
					</Button>
					<Button
						type="button"
						variant="outline"
						size="sm"
						onClick={() => openEdit(row.lesson)}
					>
						{m["admin.actions.edit"]()}
					</Button>
					<Button
						type="button"
						variant="outline"
						size="sm"
						onClick={() => togglePublish(row.lesson)}
					>
						{row.lesson.status === "published"
							? m["admin.actions.unpublish"]()
							: m["admin.actions.publish"]()}
					</Button>
					<Button
						type="button"
						variant="ghost"
						size="sm"
						onClick={() => setDeleteTarget(row.lesson)}
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
				rows={rows}
				rowKey={(row) => row.lesson.id}
				isLoading={hubQuery.isPending}
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
					<FieldLabel htmlFor="lesson-title">Title</FieldLabel>
					<Input
						id="lesson-title"
						value={form.title}
						onChange={(event) =>
							setForm((current) => ({ ...current, title: event.target.value }))
						}
					/>
				</Field>
				<div className="grid gap-3 sm:grid-cols-3">
					<Field data-invalid={Boolean(error)}>
						<FieldLabel htmlFor="lesson-slug">Slug</FieldLabel>
						<Input
							id="lesson-slug"
							value={form.slug}
							onChange={(event) =>
								setForm((current) => ({ ...current, slug: event.target.value }))
							}
						/>
					</Field>
					<Field data-invalid={Boolean(error)}>
						<FieldLabel>Section</FieldLabel>
						<Select
							value={form.sectionId || "none"}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									sectionId: value === "none" ? "" : value,
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">—</SelectItem>
								{sections.map((section) => (
									<SelectItem key={section.id} value={section.id}>
										{section.title}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
					<Field>
						<FieldLabel>CEFR</FieldLabel>
						<Select
							value={form.cefrLevel}
							onValueChange={(value) =>
								setForm((current) => ({ ...current, cefrLevel: value as CEFR }))
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
				</div>
				<div className="grid gap-3 sm:grid-cols-3">
					<Field>
						<FieldLabel>Primary skill</FieldLabel>
						<Select
							value={form.skill || "none"}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									skill: value === "none" ? "" : (value as ActivityType),
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">—</SelectItem>
								{SKILL_OPTIONS.map((option) => (
									<SelectItem key={option} value={option}>
										{option}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
					<Field>
						<FieldLabel htmlFor="lesson-duration">
							Est. duration (min)
						</FieldLabel>
						<Input
							id="lesson-duration"
							type="number"
							value={form.estimatedDurationMin}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									estimatedDurationMin: event.target.value,
								}))
							}
						/>
					</Field>
					<Field>
						<FieldLabel htmlFor="lesson-cover">Cover image URL</FieldLabel>
						<Input
							id="lesson-cover"
							value={form.coverImageUrl}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									coverImageUrl: event.target.value,
								}))
							}
						/>
					</Field>
				</div>
				<Field>
					<FieldLabel htmlFor="lesson-description">Description</FieldLabel>
					<Textarea
						id="lesson-description"
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
			<ActivitiesDialog
				lesson={activitiesLesson}
				onOpenChange={(open) => !open && setActivitiesLesson(null)}
			/>
		</div>
	);
}
