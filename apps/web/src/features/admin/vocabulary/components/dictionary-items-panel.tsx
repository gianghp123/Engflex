import type {
	CEFR,
	VocabularyDomain,
	VocabularyItem,
} from "@engflex/contracts";
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
	useAdminVocabularyItems,
	useCreateVocabularyItem,
	useDeleteVocabularyItem,
	useUpdateVocabularyItem,
} from "#/features/admin/vocabulary/queries";
import { m } from "#/paraglide/messages";

const CEFR_OPTIONS: CEFR[] = ["A1", "A2", "B1", "B2", "C1", "C2"];
const DOMAIN_OPTIONS: VocabularyDomain[] = [
	"backend_db",
	"distributed_systems",
	"devops_cloud",
	"frontend_ui",
	"ai_ml",
];

function domainLabel(domain: string): string {
	return domain.replace(/_/g, " ");
}

interface ItemFormState {
	term: string;
	definition: string;
	cefr: CEFR;
	domain: VocabularyDomain | "";
	status: "draft" | "published";
}

const EMPTY_FORM: ItemFormState = {
	term: "",
	definition: "",
	cefr: "B1",
	domain: "",
	status: "draft",
};

export function DictionaryItemsPanel() {
	const [page, setPage] = useState(1);
	const pageSize = 20;
	const query = useAdminVocabularyItems(page, pageSize);
	const createMutation = useCreateVocabularyItem();
	const updateMutation = useUpdateVocabularyItem();
	const deleteMutation = useDeleteVocabularyItem();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<VocabularyItem | null>(null);
	const [form, setForm] = useState<ItemFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] = useState<VocabularyItem | null>(null);

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(item: VocabularyItem) {
		setEditing(item);
		setForm({
			term: item.term,
			definition: item.definition,
			cefr: item.cefr,
			domain: item.domain ?? "",
			status: item.status === "published" ? "published" : "draft",
		});
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.term.trim() || !form.definition.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		if (editing) {
			updateMutation.mutate(
				{
					id: editing.id,
					payload: {
						term: form.term.trim(),
						definition: form.definition.trim(),
						cefr: form.cefr,
						domain: form.domain || undefined,
						status: form.status,
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
					term: form.term.trim(),
					definition: form.definition.trim(),
					cefr: form.cefr,
					domain: form.domain || undefined,
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

	function togglePublish(item: VocabularyItem) {
		const nextStatus = item.status === "published" ? "draft" : "published";
		updateMutation.mutate(
			{ id: item.id, payload: { status: nextStatus } },
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

	const columns: DataTableColumn<VocabularyItem>[] = [
		{ id: "term", header: "Term", cell: (row) => row.term },
		{ id: "cefr", header: "CEFR", cell: (row) => row.cefr },
		{
			id: "domain",
			header: "Domain",
			cell: (row) => (row.domain ? domainLabel(row.domain) : "—"),
		},
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
					<FieldLabel htmlFor="item-term">Term</FieldLabel>
					<Input
						id="item-term"
						value={form.term}
						onChange={(event) =>
							setForm((current) => ({ ...current, term: event.target.value }))
						}
					/>
				</Field>
				<div className="grid gap-3 sm:grid-cols-2">
					<Field>
						<FieldLabel>CEFR</FieldLabel>
						<Select
							value={form.cefr}
							onValueChange={(value) =>
								setForm((current) => ({ ...current, cefr: value as CEFR }))
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
						<FieldLabel>Domain</FieldLabel>
						<Select
							value={form.domain || "none"}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									domain: value === "none" ? "" : (value as VocabularyDomain),
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">—</SelectItem>
								{DOMAIN_OPTIONS.map((option) => (
									<SelectItem key={option} value={option}>
										{domainLabel(option)}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
				</div>
				<Field data-invalid={Boolean(error)}>
					<FieldLabel htmlFor="item-definition">Definition</FieldLabel>
					<Textarea
						id="item-definition"
						rows={3}
						value={form.definition}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								definition: event.target.value,
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
