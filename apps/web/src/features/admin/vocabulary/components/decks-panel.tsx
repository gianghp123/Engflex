import type { VocabularyDeckDetail } from "@engflex/contracts";
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
import { DeckItemsDialog } from "#/features/admin/vocabulary/components/deck-items-dialog";
import {
	useAdminVocabularyCategories,
	useAdminVocabularyDecks,
	useCreateVocabularyDeck,
	useDeleteVocabularyDeck,
	useUpdateVocabularyDeck,
} from "#/features/admin/vocabulary/queries";
import { m } from "#/paraglide/messages";

interface DeckFormState {
	categoryId: string;
	name: string;
	description: string;
	thumbnailUrl: string;
	level: string;
	status: "draft" | "published";
}

const EMPTY_FORM: DeckFormState = {
	categoryId: "",
	name: "",
	description: "",
	thumbnailUrl: "",
	level: "",
	status: "draft",
};

export function DecksPanel() {
	const [page, setPage] = useState(1);
	const pageSize = 20;
	const query = useAdminVocabularyDecks(page, pageSize);
	const categoriesQuery = useAdminVocabularyCategories(1, 100);
	const createMutation = useCreateVocabularyDeck();
	const updateMutation = useUpdateVocabularyDeck();
	const deleteMutation = useDeleteVocabularyDeck();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<VocabularyDeckDetail | null>(null);
	const [form, setForm] = useState<DeckFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] = useState<VocabularyDeckDetail | null>(
		null,
	);
	const [itemsDeck, setItemsDeck] = useState<VocabularyDeckDetail | null>(null);

	const categories = categoriesQuery.data?.items ?? [];

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	function openEdit(deck: VocabularyDeckDetail) {
		setEditing(deck);
		setForm({
			categoryId: deck.categoryId ?? "",
			name: deck.name,
			description: deck.description,
			thumbnailUrl: deck.thumbnailUrl,
			level: deck.level,
			status: deck.status === "published" ? "published" : "draft",
		});
		setError(null);
		setDialogOpen(true);
	}

	function submit() {
		if (!form.name.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		if (editing) {
			updateMutation.mutate(
				{
					id: editing.id,
					payload: {
						categoryId: form.categoryId || undefined,
						name: form.name.trim(),
						description: form.description,
						thumbnailUrl: form.thumbnailUrl,
						level: form.level,
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
					categoryId: form.categoryId || undefined,
					name: form.name.trim(),
					description: form.description,
					thumbnailUrl: form.thumbnailUrl,
					level: form.level,
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

	function togglePublish(deck: VocabularyDeckDetail) {
		const nextStatus = deck.status === "published" ? "draft" : "published";
		updateMutation.mutate(
			{ id: deck.id, payload: { status: nextStatus } },
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

	const columns: DataTableColumn<VocabularyDeckDetail>[] = [
		{ id: "name", header: "Name", cell: (row) => row.name },
		{ id: "level", header: "Level", cell: (row) => row.level || "—" },
		{
			id: "category",
			header: "Category",
			cell: (row) => row.category?.name ?? "—",
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
						onClick={() => setItemsDeck(row)}
					>
						Cards
					</Button>
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
					<FieldLabel htmlFor="deck-name">Name</FieldLabel>
					<Input
						id="deck-name"
						value={form.name}
						onChange={(event) =>
							setForm((current) => ({ ...current, name: event.target.value }))
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
						<FieldLabel htmlFor="deck-level">Level</FieldLabel>
						<Input
							id="deck-level"
							value={form.level}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									level: event.target.value,
								}))
							}
						/>
					</Field>
				</div>
				<Field>
					<FieldLabel htmlFor="deck-description">Description</FieldLabel>
					<Textarea
						id="deck-description"
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
				<Field>
					<FieldLabel htmlFor="deck-thumbnail">Thumbnail URL</FieldLabel>
					<Input
						id="deck-thumbnail"
						value={form.thumbnailUrl}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								thumbnailUrl: event.target.value,
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
			<DeckItemsDialog
				deck={itemsDeck}
				onOpenChange={(open) => !open && setItemsDeck(null)}
			/>
		</div>
	);
}
