import type { ActivityType, Lesson } from "@engflex/contracts";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import {
	Field,
	FieldError,
	FieldGroup,
	FieldLabel,
} from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "#/components/ui/table";
import { Textarea } from "#/components/ui/textarea";
import { ConfirmDeleteDialog } from "#/features/admin/components/confirm-delete-dialog";
import {
	useAdminLessonActivities,
	useAdminLessonActivity,
	useCreateLessonActivity,
	useDeleteLessonActivity,
	useUpdateLessonActivity,
} from "#/features/admin/lessons/queries";
import { m } from "#/paraglide/messages";

const ACTIVITY_TYPES: ActivityType[] = [
	"reading",
	"listening",
	"writing",
	"speaking",
];

/** Admin-authoring shapes — mirror apps/api/internal/modules/lessons/
 *  activityconfig/activity_config.go field for field (answer keys and
 *  voice ids included; the learner-facing payload strips those). */
const CONFIG_TEMPLATES: Record<ActivityType, unknown> = {
	reading: {
		text: "",
		questions: [
			{
				stem: "",
				instruction: "",
				options: [{ key: "a", text: "" }],
				answerKey: "a",
				explanation: "",
			},
		],
	},
	listening: {
		audioKey: "",
		script: [{ speaker: "", voice: "", text: "" }],
		questions: [
			{
				stem: "",
				instruction: "",
				options: [{ key: "a", text: "" }],
				answerKey: "a",
				explanation: "",
			},
		],
	},
	writing: {
		task: "email",
		instructions: "",
		stimulus: "",
		minWords: 50,
		maxWords: 150,
		modelAnswer: "",
		checklist: [""],
	},
	speaking: {
		items: [{ text: "", modelAudioKey: "", modelVoice: "" }],
	},
};

function templateFor(type: ActivityType): string {
	return JSON.stringify(CONFIG_TEMPLATES[type], null, 2);
}

interface ActivityFormState {
	partNumber: string;
	type: ActivityType;
	title: string;
	description: string;
	skillFocus: string;
	durationMin: string;
	configText: string;
}

const EMPTY_FORM: ActivityFormState = {
	partNumber: "1",
	type: "reading",
	title: "",
	description: "",
	skillFocus: "",
	durationMin: "5",
	configText: templateFor("reading"),
};

/** Manages one lesson's ordered activities — opened from the Lessons
 *  table's "Activities" action. `lesson` null means closed. Config is
 *  authored as raw JSON (with a per-type template to start from) rather
 *  than a fully dynamic field-by-field builder for every nested
 *  question/script-line/checklist array — the admin surface area here
 *  is small enough that a template + textarea is the pragmatic shape. */
export function ActivitiesDialog({
	lesson,
	onOpenChange,
}: {
	lesson: Lesson | null;
	onOpenChange: (open: boolean) => void;
}) {
	const lessonId = lesson?.id ?? "";
	const query = useAdminLessonActivities(lessonId);
	const createMutation = useCreateLessonActivity(lessonId);
	const updateMutation = useUpdateLessonActivity(lessonId);
	const deleteMutation = useDeleteLessonActivity(lessonId);

	const [editingId, setEditingId] = useState<string | null>(null);
	const editingQuery = useAdminLessonActivity(editingId);
	const [form, setForm] = useState<ActivityFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] = useState<{
		id: string;
		title: string;
	} | null>(null);

	useEffect(() => {
		if (editingId && editingQuery.data) {
			const activity = editingQuery.data;
			setForm({
				partNumber: String(activity.partNumber),
				type: activity.type,
				title: activity.title,
				description: activity.description,
				skillFocus: activity.skillFocus,
				durationMin: String(activity.durationMin),
				configText: JSON.stringify(activity.config, null, 2),
			});
		}
	}, [editingId, editingQuery.data]);

	function resetForm() {
		setEditingId(null);
		setForm(EMPTY_FORM);
		setError(null);
	}

	function submit() {
		let config: unknown;
		try {
			config = JSON.parse(form.configText);
		} catch {
			setError("Config must be valid JSON");
			return;
		}
		const base = {
			partNumber: Number(form.partNumber) || 1,
			type: form.type,
			title: form.title.trim(),
			description: form.description,
			skillFocus: form.skillFocus,
			durationMin: Number(form.durationMin) || 0,
			config,
		};
		if (!base.title) {
			setError(m["admin.toast.error"]());
			return;
		}
		if (editingId) {
			updateMutation.mutate(
				{ id: editingId, payload: base },
				{
					onSuccess: () => {
						toast.success(m["admin.toast.updated"]());
						resetForm();
					},
					onError: () => toast.error(m["admin.toast.error"]()),
				},
			);
		} else {
			createMutation.mutate(
				{ ...base, lessonId },
				{
					onSuccess: () => {
						toast.success(m["admin.toast.created"]());
						resetForm();
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

	return (
		<Dialog
			open={Boolean(lesson)}
			onOpenChange={(open) => {
				if (!open) resetForm();
				onOpenChange(open);
			}}
		>
			<DialogContent className="max-h-[85vh] max-w-3xl overflow-y-auto">
				<DialogHeader>
					<DialogTitle>Activities — {lesson?.title}</DialogTitle>
				</DialogHeader>
				<form
					onSubmit={(event) => {
						event.preventDefault();
						submit();
					}}
				>
					<FieldGroup>
						<div className="grid gap-3 sm:grid-cols-4">
							<Field data-invalid={Boolean(error)}>
								<FieldLabel htmlFor="activity-title">Title</FieldLabel>
								<Input
									id="activity-title"
									value={form.title}
									onChange={(event) =>
										setForm((current) => ({
											...current,
											title: event.target.value,
										}))
									}
								/>
							</Field>
							<Field>
								<FieldLabel>Type</FieldLabel>
								<Select
									value={form.type}
									onValueChange={(value) =>
										setForm((current) => ({
											...current,
											type: value as ActivityType,
											configText: templateFor(value as ActivityType),
										}))
									}
								>
									<SelectTrigger className="w-full">
										<SelectValue />
									</SelectTrigger>
									<SelectContent>
										{ACTIVITY_TYPES.map((type) => (
											<SelectItem key={type} value={type}>
												{type}
											</SelectItem>
										))}
									</SelectContent>
								</Select>
							</Field>
							<Field>
								<FieldLabel htmlFor="activity-part">Part #</FieldLabel>
								<Input
									id="activity-part"
									type="number"
									value={form.partNumber}
									onChange={(event) =>
										setForm((current) => ({
											...current,
											partNumber: event.target.value,
										}))
									}
								/>
							</Field>
							<Field>
								<FieldLabel htmlFor="activity-duration">
									Duration (min)
								</FieldLabel>
								<Input
									id="activity-duration"
									type="number"
									value={form.durationMin}
									onChange={(event) =>
										setForm((current) => ({
											...current,
											durationMin: event.target.value,
										}))
									}
								/>
							</Field>
						</div>
						<div className="grid gap-3 sm:grid-cols-2">
							<Field>
								<FieldLabel htmlFor="activity-description">
									Description
								</FieldLabel>
								<Input
									id="activity-description"
									value={form.description}
									onChange={(event) =>
										setForm((current) => ({
											...current,
											description: event.target.value,
										}))
									}
								/>
							</Field>
							<Field>
								<FieldLabel htmlFor="activity-skill">Skill focus</FieldLabel>
								<Input
									id="activity-skill"
									value={form.skillFocus}
									onChange={(event) =>
										setForm((current) => ({
											...current,
											skillFocus: event.target.value,
										}))
									}
								/>
							</Field>
						</div>
						<Field data-invalid={Boolean(error)}>
							<div className="flex items-center justify-between">
								<FieldLabel htmlFor="activity-config">Config (JSON)</FieldLabel>
								<Button
									type="button"
									variant="ghost"
									size="xs"
									onClick={() =>
										setForm((current) => ({
											...current,
											configText: templateFor(current.type),
										}))
									}
								>
									Reset to template
								</Button>
							</div>
							<Textarea
								id="activity-config"
								rows={10}
								className="font-mono text-xs"
								value={form.configText}
								onChange={(event) =>
									setForm((current) => ({
										...current,
										configText: event.target.value,
									}))
								}
							/>
						</Field>
						{error ? <FieldError>{error}</FieldError> : null}
					</FieldGroup>
					<div className="mt-3 flex justify-end gap-2">
						{editingId ? (
							<Button type="button" variant="ghost" onClick={resetForm}>
								{m["admin.actions.cancel"]()}
							</Button>
						) : null}
						<SubmitButton
							type="submit"
							size="sm"
							pending={createMutation.isPending || updateMutation.isPending}
						>
							{editingId
								? m["admin.actions.save"]()
								: m["admin.actions.create"]()}
						</SubmitButton>
					</div>
				</form>
				<Table>
					<TableHeader>
						<TableRow>
							<TableHead>#</TableHead>
							<TableHead>Title</TableHead>
							<TableHead>Type</TableHead>
							<TableHead>{m["admin.table.actions"]()}</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{(query.data ?? []).map((activity) => (
							<TableRow key={activity.id}>
								<TableCell>{activity.partNumber}</TableCell>
								<TableCell>{activity.title}</TableCell>
								<TableCell>{activity.type}</TableCell>
								<TableCell>
									<div className="flex items-center gap-2">
										<Button
											type="button"
											variant="outline"
											size="sm"
											onClick={() => setEditingId(activity.id)}
										>
											{m["admin.actions.edit"]()}
										</Button>
										<Button
											type="button"
											variant="ghost"
											size="sm"
											onClick={() =>
												setDeleteTarget({
													id: activity.id,
													title: activity.title,
												})
											}
										>
											{m["admin.actions.delete"]()}
										</Button>
									</div>
								</TableCell>
							</TableRow>
						))}
					</TableBody>
				</Table>
				<ConfirmDeleteDialog
					open={Boolean(deleteTarget)}
					onOpenChange={(open) => !open && setDeleteTarget(null)}
					onConfirm={confirmDelete}
					pending={deleteMutation.isPending}
				/>
			</DialogContent>
		</Dialog>
	);
}
