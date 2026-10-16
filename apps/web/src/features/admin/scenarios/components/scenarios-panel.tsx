import type { Scenario, ScenarioDifficulty } from "@engflex/contracts";
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
	useAdminPersonas,
	useAdminScenarios,
	useAdminScenarioTopics,
	useCreateScenario,
	useDeleteScenario,
	useUpdateScenario,
} from "#/features/admin/scenarios/queries";
import { m } from "#/paraglide/messages";

const CEFR_OPTIONS: ScenarioDifficulty[] = ["B1+", "B2", "C1"];

interface ScenarioFormState {
	topicId: string;
	personaId: string;
	title: string;
	objective: string;
	cefrLevel: ScenarioDifficulty;
	maxDuration: string;
	role: string;
	interlocutorRole: string;
	goalFormat: string;
	context: string;
	openingLabel: string;
	openingText: string;
	vocab: string;
	status: "draft" | "published";
}

const EMPTY_FORM: ScenarioFormState = {
	topicId: "",
	personaId: "",
	title: "",
	objective: "",
	cefrLevel: "B1+",
	maxDuration: "15",
	role: "",
	interlocutorRole: "",
	goalFormat: "",
	context: "",
	openingLabel: "",
	openingText: "",
	vocab: "",
	status: "draft",
};

const linesOf = (text: string) =>
	text
		.split("\n")
		.map((line) => line.trim())
		.filter(Boolean);

export function ScenariosPanel() {
	const [page, setPage] = useState(1);
	const pageSize = 20;
	const query = useAdminScenarios(page, pageSize);
	const topicsQuery = useAdminScenarioTopics();
	const personasQuery = useAdminPersonas(1, 100);
	const createMutation = useCreateScenario();
	const updateMutation = useUpdateScenario();
	const deleteMutation = useDeleteScenario();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<Scenario | null>(null);
	const [form, setForm] = useState<ScenarioFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] = useState<Scenario | null>(null);

	const topics = topicsQuery.data ?? [];
	const personas = personasQuery.data?.items ?? [];

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(scenario: Scenario) {
		setEditing(scenario);
		setForm({
			topicId: scenario.topic?.id ?? "",
			personaId: scenario.persona?.id ?? "",
			title: scenario.title,
			objective: scenario.objective,
			cefrLevel: scenario.cefrLevel,
			maxDuration: String(scenario.maxDuration),
			role: scenario.details.role ?? "",
			interlocutorRole: scenario.details.interlocutorRole ?? "",
			goalFormat: scenario.details.goalFormat ?? "",
			context: (scenario.details.context ?? []).join("\n"),
			openingLabel: scenario.details.opening?.label ?? "",
			openingText: scenario.details.opening?.text ?? "",
			vocab: (scenario.details.vocab ?? []).join("\n"),
			status: scenario.status === "published" ? "published" : "draft",
		});
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.topicId || !form.title.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		const details = {
			role: form.role,
			interlocutorRole: form.interlocutorRole,
			goalFormat: form.goalFormat,
			context: linesOf(form.context),
			opening: { label: form.openingLabel, text: form.openingText },
			vocab: linesOf(form.vocab),
		};
		const base = {
			topicId: form.topicId,
			personaId: form.personaId || undefined,
			title: form.title.trim(),
			objective: form.objective,
			cefrLevel: form.cefrLevel,
			maxDuration: Number(form.maxDuration) || 1,
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

	function togglePublish(scenario: Scenario) {
		const nextStatus = scenario.status === "published" ? "draft" : "published";
		updateMutation.mutate(
			{ id: scenario.id, payload: { status: nextStatus } },
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

	const columns: DataTableColumn<Scenario>[] = [
		{ id: "title", header: "Title", cell: (row) => row.title },
		{ id: "topic", header: "Topic", cell: (row) => row.topic?.name ?? "—" },
		{ id: "cefr", header: "CEFR", cell: (row) => row.cefrLevel },
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
					<FieldLabel htmlFor="scenario-title">Title</FieldLabel>
					<Input
						id="scenario-title"
						value={form.title}
						onChange={(event) =>
							setForm((current) => ({ ...current, title: event.target.value }))
						}
					/>
				</Field>
				<div className="grid gap-3 sm:grid-cols-3">
					<Field data-invalid={Boolean(error)}>
						<FieldLabel>Topic</FieldLabel>
						<Select
							value={form.topicId || "none"}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									topicId: value === "none" ? "" : value,
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">—</SelectItem>
								{topics.map((topic) => (
									<SelectItem key={topic.id} value={topic.id}>
										{topic.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
					<Field>
						<FieldLabel>Persona</FieldLabel>
						<Select
							value={form.personaId || "none"}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									personaId: value === "none" ? "" : value,
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">—</SelectItem>
								{personas.map((persona) => (
									<SelectItem key={persona.id} value={persona.id}>
										{persona.name}
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
								setForm((current) => ({
									...current,
									cefrLevel: value as ScenarioDifficulty,
								}))
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
				<div className="grid gap-3 sm:grid-cols-2">
					<Field>
						<FieldLabel htmlFor="scenario-objective">Objective</FieldLabel>
						<Input
							id="scenario-objective"
							value={form.objective}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									objective: event.target.value,
								}))
							}
						/>
					</Field>
					<Field>
						<FieldLabel htmlFor="scenario-duration">
							Max duration (min)
						</FieldLabel>
						<Input
							id="scenario-duration"
							type="number"
							value={form.maxDuration}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									maxDuration: event.target.value,
								}))
							}
						/>
					</Field>
				</div>
				<div className="grid gap-3 sm:grid-cols-2">
					<Field>
						<FieldLabel htmlFor="scenario-role">Role</FieldLabel>
						<Input
							id="scenario-role"
							value={form.role}
							onChange={(event) =>
								setForm((current) => ({ ...current, role: event.target.value }))
							}
						/>
					</Field>
					<Field>
						<FieldLabel htmlFor="scenario-interlocutor">
							Interlocutor role
						</FieldLabel>
						<Input
							id="scenario-interlocutor"
							value={form.interlocutorRole}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									interlocutorRole: event.target.value,
								}))
							}
						/>
					</Field>
				</div>
				<Field>
					<FieldLabel htmlFor="scenario-goal-format">Goal format</FieldLabel>
					<Input
						id="scenario-goal-format"
						value={form.goalFormat}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								goalFormat: event.target.value,
							}))
						}
					/>
				</Field>
				<Field>
					<FieldLabel htmlFor="scenario-context">
						Context (one per line)
					</FieldLabel>
					<Textarea
						id="scenario-context"
						rows={3}
						value={form.context}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								context: event.target.value,
							}))
						}
					/>
				</Field>
				<div className="grid gap-3 sm:grid-cols-2">
					<Field>
						<FieldLabel htmlFor="scenario-opening-label">
							Opening label
						</FieldLabel>
						<Input
							id="scenario-opening-label"
							value={form.openingLabel}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									openingLabel: event.target.value,
								}))
							}
						/>
					</Field>
					<Field>
						<FieldLabel htmlFor="scenario-opening-text">
							Opening text
						</FieldLabel>
						<Input
							id="scenario-opening-text"
							value={form.openingText}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									openingText: event.target.value,
								}))
							}
						/>
					</Field>
				</div>
				<Field>
					<FieldLabel htmlFor="scenario-vocab">Vocab (one per line)</FieldLabel>
					<Textarea
						id="scenario-vocab"
						rows={3}
						value={form.vocab}
						onChange={(event) =>
							setForm((current) => ({ ...current, vocab: event.target.value }))
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
