import type { CEFR, Gender, Persona } from "@engflex/contracts";
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
	useAdminPersonas,
	useCreatePersona,
	useDeletePersona,
	useUpdatePersona,
} from "#/features/admin/scenarios/queries";
import { m } from "#/paraglide/messages";

const CEFR_OPTIONS: CEFR[] = ["A1", "A2", "B1", "B2", "C1", "C2"];

interface PersonaFormState {
	name: string;
	roleTitle: string;
	personality: string;
	style: string;
	objective: string;
	gender: Gender;
	defaultCefr: CEFR | "";
}

const EMPTY_FORM: PersonaFormState = {
	name: "",
	roleTitle: "",
	personality: "",
	style: "",
	objective: "",
	gender: "female",
	defaultCefr: "",
};

export function PersonasPanel() {
	const [page, setPage] = useState(1);
	const pageSize = 20;
	const query = useAdminPersonas(page, pageSize);
	const createMutation = useCreatePersona();
	const updateMutation = useUpdatePersona();
	const deleteMutation = useDeletePersona();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<Persona | null>(null);
	const [form, setForm] = useState<PersonaFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] = useState<Persona | null>(null);

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(persona: Persona) {
		setEditing(persona);
		setForm({
			name: persona.name,
			roleTitle: persona.roleTitle,
			personality: persona.personality,
			style: persona.style,
			objective: persona.objective,
			gender: persona.gender,
			defaultCefr: persona.defaultCefr ?? "",
		});
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.name.trim() || !form.roleTitle.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		const payload = {
			name: form.name.trim(),
			roleTitle: form.roleTitle.trim(),
			personality: form.personality || undefined,
			style: form.style || undefined,
			objective: form.objective || undefined,
			gender: form.gender,
			defaultCefr: form.defaultCefr || undefined,
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

	const columns: DataTableColumn<Persona>[] = [
		{ id: "name", header: "Name", cell: (row) => row.name },
		{ id: "roleTitle", header: "Role", cell: (row) => row.roleTitle },
		{ id: "gender", header: "Gender", cell: (row) => row.gender },
		{
			id: "cefr",
			header: "Default CEFR",
			cell: (row) => row.defaultCefr ?? "—",
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
				<div className="grid gap-3 sm:grid-cols-2">
					<Field data-invalid={Boolean(error)}>
						<FieldLabel htmlFor="persona-name">Name</FieldLabel>
						<Input
							id="persona-name"
							value={form.name}
							onChange={(event) =>
								setForm((current) => ({ ...current, name: event.target.value }))
							}
						/>
					</Field>
					<Field data-invalid={Boolean(error)}>
						<FieldLabel htmlFor="persona-role">Role title</FieldLabel>
						<Input
							id="persona-role"
							value={form.roleTitle}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									roleTitle: event.target.value,
								}))
							}
						/>
					</Field>
				</div>
				<div className="grid gap-3 sm:grid-cols-2">
					<Field>
						<FieldLabel>Gender</FieldLabel>
						<Select
							value={form.gender}
							onValueChange={(value) =>
								setForm((current) => ({ ...current, gender: value as Gender }))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="female">Female</SelectItem>
								<SelectItem value="male">Male</SelectItem>
							</SelectContent>
						</Select>
					</Field>
					<Field>
						<FieldLabel>Default CEFR</FieldLabel>
						<Select
							value={form.defaultCefr || "none"}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									defaultCefr: value === "none" ? "" : (value as CEFR),
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
				<Field>
					<FieldLabel htmlFor="persona-personality">Personality</FieldLabel>
					<Textarea
						id="persona-personality"
						rows={2}
						value={form.personality}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								personality: event.target.value,
							}))
						}
					/>
				</Field>
				<Field>
					<FieldLabel htmlFor="persona-style">Style</FieldLabel>
					<Textarea
						id="persona-style"
						rows={2}
						value={form.style}
						onChange={(event) =>
							setForm((current) => ({ ...current, style: event.target.value }))
						}
					/>
				</Field>
				<Field>
					<FieldLabel htmlFor="persona-objective">Objective</FieldLabel>
					<Textarea
						id="persona-objective"
						rows={2}
						value={form.objective}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								objective: event.target.value,
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
