import type { CEFR, VocabularyItem } from "@engflex/contracts";
import { VOCABULARY_PART_OF_SPEECH_VALUES } from "@engflex/contracts";
import { Eye, EyeOff, Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "#/components/ui/badge";
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
import { Switch } from "#/components/ui/switch";
import { Textarea } from "#/components/ui/textarea";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#/components/ui/tooltip";
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
import { enumOptions } from "#/lib/enum-options";
import { m } from "#/paraglide/messages";

const CEFR_OPTIONS: CEFR[] = ["A1", "A2", "B1", "B2", "C1", "C2"];
const POS_OPTIONS = enumOptions(VOCABULARY_PART_OF_SPEECH_VALUES);

interface ItemFormState {
	term: string;
	definition: string;
	cefr: CEFR;
	partOfSpeech: (typeof VOCABULARY_PART_OF_SPEECH_VALUES)[number] | "";
	status: "draft" | "published";
	isPremium: boolean;
}

const EMPTY_FORM: ItemFormState = {
	term: "",
	definition: "",
	cefr: "B1",
	partOfSpeech: "",
	status: "draft",
	isPremium: false,
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
			partOfSpeech: item.partOfSpeech ?? "",
			status: item.status === "published" ? "published" : "draft",
			isPremium: item.isPremium,
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
						partOfSpeech: form.partOfSpeech || undefined,
						status: form.status,
						isPremium: form.isPremium,
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
					partOfSpeech: form.partOfSpeech || undefined,
					isPremium: form.isPremium,
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
		{
			id: "cefr",
			header: "CEFR",
			className: "w-20",
			cell: (row) => row.cefr,
		},
		{
			id: "partOfSpeech",
			header: "Part of speech",
			className: "w-32",
			cell: (row) => row.partOfSpeech ?? "—",
		},
		{
			id: "status",
			header: m["admin.table.status"](),
			className: "w-28",
			cell: (row) => (
				<div className="flex items-center gap-1.5">
					<StatusBadge
						status={row.status === "published" ? "published" : "draft"}
					/>
					{row.isPremium && (
						<Badge variant="default">{m["admin.content.premiumBadge"]()}</Badge>
					)}
				</div>
			),
		},
		{
			id: "actions",
			header: m["admin.table.actions"](),
			className: "w-36",
			cell: (row) => (
				<div className="flex items-center gap-1">
					<Tooltip>
						<TooltipTrigger asChild>
							<Button
								type="button"
								variant="outline"
								size="icon-sm"
								aria-label={m["admin.actions.edit"]()}
								onClick={() => openEdit(row)}
							>
								<Pencil />
							</Button>
						</TooltipTrigger>
						<TooltipContent>{m["admin.actions.edit"]()}</TooltipContent>
					</Tooltip>
					<Tooltip>
						<TooltipTrigger asChild>
							<Button
								type="button"
								variant="outline"
								size="icon-sm"
								aria-label={
									row.status === "published"
										? m["admin.actions.unpublish"]()
										: m["admin.actions.publish"]()
								}
								onClick={() => togglePublish(row)}
							>
								{row.status === "published" ? <EyeOff /> : <Eye />}
							</Button>
						</TooltipTrigger>
						<TooltipContent>
							{row.status === "published"
								? m["admin.actions.unpublish"]()
								: m["admin.actions.publish"]()}
						</TooltipContent>
					</Tooltip>
					<Tooltip>
						<TooltipTrigger asChild>
							<Button
								type="button"
								variant="ghost"
								size="icon-sm"
								aria-label={m["admin.actions.delete"]()}
								onClick={() => setDeleteTarget(row)}
							>
								<Trash2 />
							</Button>
						</TooltipTrigger>
						<TooltipContent>{m["admin.actions.delete"]()}</TooltipContent>
					</Tooltip>
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
						<FieldLabel>Part of speech</FieldLabel>
						<Select
							value={form.partOfSpeech || "none"}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									partOfSpeech:
										value === "none"
											? ""
											: (value as ItemFormState["partOfSpeech"]),
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">—</SelectItem>
								{POS_OPTIONS.map((option) => (
									<SelectItem key={option.value} value={option.value}>
										{option.label()}
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
				<Field className="flex-row items-center justify-between">
					<FieldLabel htmlFor="item-premium">
						{m["admin.content.isPremium"]()}
					</FieldLabel>
					<Switch
						id="item-premium"
						checked={form.isPremium}
						onCheckedChange={(checked) =>
							setForm((current) => ({ ...current, isPremium: checked }))
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
