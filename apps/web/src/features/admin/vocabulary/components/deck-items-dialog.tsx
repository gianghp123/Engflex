import type {
	VocabularyDeckDetail,
	VocabularyDeckItemResponse,
} from "@engflex/contracts";
import { useState } from "react";
import { toast } from "sonner";
import { CardShell } from "#/components/common/card-shell";
import { Pager } from "#/components/common/pager";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Field, FieldGroup, FieldLabel } from "#/components/ui/field";
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
	useAdminVocabularyDeckItems,
	useAdminVocabularyItems,
	useCreateVocabularyDeckItem,
	useDeleteVocabularyDeckItem,
	useUpdateVocabularyDeckItem,
} from "#/features/admin/vocabulary/queries";
import { m } from "#/paraglide/messages";

const EMPTY_CARD = { itemId: "", sequence: 0, exampleSentence: "", note: "" };

export function DeckItemsDialog({
	deck,
	onOpenChange,
}: {
	deck: VocabularyDeckDetail | null;
	onOpenChange: (open: boolean) => void;
}) {
	return (
		<Dialog open={Boolean(deck)} onOpenChange={onOpenChange}>
			<DialogContent className="max-h-[85vh] max-w-2xl overflow-y-auto">
				<DialogHeader>
					<DialogTitle>
						{m["admin.deckCards.title"]()} — {deck?.name}
					</DialogTitle>
				</DialogHeader>
				{deck && <DeckCards key={deck.id} deck={deck} />}
			</DialogContent>
		</Dialog>
	);
}

function DeckCards({ deck }: { deck: VocabularyDeckDetail }) {
	const query = useAdminVocabularyDeckItems(deck.id);
	const createMutation = useCreateVocabularyDeckItem(deck.id);
	const updateMutation = useUpdateVocabularyDeckItem(deck.id);
	const deleteMutation = useDeleteVocabularyDeckItem(deck.id);
	const [editing, setEditing] = useState<VocabularyDeckItemResponse | null>(
		null,
	);
	const [form, setForm] = useState(EMPTY_CARD);
	const [deleteTarget, setDeleteTarget] =
		useState<VocabularyDeckItemResponse | null>(null);
	const [page, setPage] = useState(1);
	const [search, setSearch] = useState("");
	const dictionary = useAdminVocabularyItems(page, 20, search);
	const words = dictionary.data?.items ?? [];
	function resetForm() {
		setEditing(null);
		setForm(EMPTY_CARD);
	}
	function submit() {
		if (!form.itemId) return;
		const options = {
			onSuccess: () => {
				toast.success(
					editing ? m["admin.toast.updated"]() : m["admin.toast.created"](),
				);
				resetForm();
			},
			onError: () => toast.error(m["admin.toast.error"]()),
		};
		if (editing)
			updateMutation.mutate({ id: editing.id, payload: form }, options);
		else createMutation.mutate({ deckId: deck.id, ...form }, options);
	}
	return (
		<>
			<form
				onSubmit={(event) => {
					event.preventDefault();
					submit();
				}}
			>
				<FieldGroup>
					<Field>
						<FieldLabel htmlFor="card-search">
							{m["admin.deckCards.search"]()}
						</FieldLabel>
						<Input
							id="card-search"
							value={search}
							onChange={(event) => {
								setSearch(event.target.value);
								setPage(1);
							}}
						/>
					</Field>
					<Field>
						<FieldLabel>{m["admin.deckCards.dictionary"]()}</FieldLabel>
						<Select
							value={form.itemId}
							onValueChange={(itemId) =>
								setForm((current) => ({ ...current, itemId }))
							}
						>
							<SelectTrigger>
								<SelectValue
									placeholder={
										editing?.item.term ?? m["admin.deckCards.dictionary"]()
									}
								/>
							</SelectTrigger>
							<SelectContent>
								{words.map((word) => (
									<SelectItem key={word.id} value={word.id}>
										{word.term}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
						{dictionary.isPending && <span>{m["admin.table.loading"]()}</span>}
						{dictionary.isError && (
							<span role="alert">{m["admin.toast.error"]()}</span>
						)}
					</Field>
					{words.length > 0 && (
						<Pager
							page={page}
							totalPages={dictionary.data?.pagination.totalPages ?? 1}
							onPageChange={setPage}
						/>
					)}
					<Field>
						<FieldLabel htmlFor="card-sequence">
							{m["admin.deckCards.sequence"]()}
						</FieldLabel>
						<Input
							id="card-sequence"
							type="number"
							min={0}
							value={form.sequence}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									sequence: Math.max(0, Number(event.target.value)),
								}))
							}
						/>
					</Field>
					<Field>
						<FieldLabel htmlFor="card-example">
							{m["admin.deckCards.example"]()}
						</FieldLabel>
						<Textarea
							id="card-example"
							rows={2}
							value={form.exampleSentence}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									exampleSentence: event.target.value,
								}))
							}
						/>
					</Field>
					<Field>
						<FieldLabel htmlFor="card-note">
							{m["admin.deckCards.note"]()}
						</FieldLabel>
						<Input
							id="card-note"
							value={form.note}
							onChange={(event) =>
								setForm((current) => ({ ...current, note: event.target.value }))
							}
						/>
					</Field>
				</FieldGroup>
				<div className="mt-3 flex justify-end gap-2">
					{editing && (
						<Button type="button" variant="ghost" onClick={resetForm}>
							{m["admin.actions.cancel"]()}
						</Button>
					)}
					<SubmitButton
						type="submit"
						size="sm"
						disabled={!form.itemId}
						pending={createMutation.isPending || updateMutation.isPending}
					>
						{editing ? m["admin.actions.save"]() : m["admin.actions.create"]()}
					</SubmitButton>
				</div>
			</form>
			<CardShell pending={query.isPending}>
				{query.isError ? (
					<p role="alert">{m["admin.toast.error"]()}</p>
				) : (
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>{m["admin.deckCards.term"]()}</TableHead>
								<TableHead>{m["admin.deckCards.definition"]()}</TableHead>
								<TableHead>{m["admin.table.actions"]()}</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{(query.data ?? []).map((item) => (
								<TableRow key={item.id}>
									<TableCell>{item.item.term}</TableCell>
									<TableCell>{item.item.definition}</TableCell>
									<TableCell>
										<div className="flex items-center gap-2">
											<Button
												type="button"
												variant="outline"
												size="sm"
												onClick={() => {
													setEditing(item);
													setForm({
														itemId: item.itemId,
														sequence: item.sequence,
														exampleSentence: item.exampleSentence,
														note: item.note,
													});
												}}
											>
												{m["admin.actions.edit"]()}
											</Button>
											<Button
												type="button"
												variant="ghost"
												size="sm"
												onClick={() => setDeleteTarget(item)}
											>
												{m["admin.actions.delete"]()}
											</Button>
										</div>
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				)}
			</CardShell>
			<ConfirmDeleteDialog
				open={Boolean(deleteTarget)}
				onOpenChange={(open) => {
					if (!open) setDeleteTarget(null);
				}}
				pending={deleteMutation.isPending}
				onConfirm={() => {
					if (!deleteTarget) return;
					deleteMutation.mutate(deleteTarget.id, {
						onSuccess: () => {
							toast.success(m["admin.toast.deleted"]());
							setDeleteTarget(null);
						},
						onError: () => toast.error(m["admin.toast.error"]()),
					});
				}}
			/>
		</>
	);
}
