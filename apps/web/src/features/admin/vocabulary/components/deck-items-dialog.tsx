import type {
	VocabularyDeckDetail,
	VocabularyDeckItemResponse,
} from "@engflex/contracts";
import { useState } from "react";
import { toast } from "sonner";
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
	useCreateVocabularyDeckItem,
	useDeleteVocabularyDeckItem,
	useUpdateVocabularyDeckItem,
} from "#/features/admin/vocabulary/queries";
import { m } from "#/paraglide/messages";

interface CardFormState {
	phrase: string;
	meaning: string;
	exampleSentence: string;
	note: string;
}

const EMPTY_CARD: CardFormState = {
	phrase: "",
	meaning: "",
	exampleSentence: "",
	note: "",
};

/** Manages one deck's flashcard rows — opened from the Decks table's "Cards"
 *  action. `deck` null means closed; keeping it as the single source of
 *  truth (instead of a separate boolean) means the dialog always knows
 *  which deck it's editing. */
export function DeckItemsDialog({
	deck,
	onOpenChange,
}: {
	deck: VocabularyDeckDetail | null;
	onOpenChange: (open: boolean) => void;
}) {
	const deckId = deck?.id ?? "";
	const query = useAdminVocabularyDeckItems(deckId);
	const createMutation = useCreateVocabularyDeckItem(deckId);
	const updateMutation = useUpdateVocabularyDeckItem(deckId);
	const deleteMutation = useDeleteVocabularyDeckItem(deckId);

	const [editing, setEditing] = useState<VocabularyDeckItemResponse | null>(
		null,
	);
	const [form, setForm] = useState<CardFormState>(EMPTY_CARD);
	const [deleteTarget, setDeleteTarget] =
		useState<VocabularyDeckItemResponse | null>(null);

	function resetForm() {
		setEditing(null);
		setForm(EMPTY_CARD);
	}

	function submit() {
		if (!form.phrase.trim() || !form.meaning.trim()) return;
		if (editing) {
			updateMutation.mutate(
				{ id: editing.id, payload: form },
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
				{ deckId, ...form },
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
			open={Boolean(deck)}
			onOpenChange={(open) => {
				if (!open) resetForm();
				onOpenChange(open);
			}}
		>
			<DialogContent className="max-h-[85vh] max-w-2xl overflow-y-auto">
				<DialogHeader>
					<DialogTitle>Cards — {deck?.name}</DialogTitle>
				</DialogHeader>
				<form
					onSubmit={(event) => {
						event.preventDefault();
						submit();
					}}
				>
					<FieldGroup>
						<div className="grid gap-3 sm:grid-cols-2">
							<Field>
								<FieldLabel htmlFor="card-phrase">Phrase</FieldLabel>
								<Input
									id="card-phrase"
									value={form.phrase}
									onChange={(event) =>
										setForm((current) => ({
											...current,
											phrase: event.target.value,
										}))
									}
								/>
							</Field>
							<Field>
								<FieldLabel htmlFor="card-meaning">Meaning</FieldLabel>
								<Input
									id="card-meaning"
									value={form.meaning}
									onChange={(event) =>
										setForm((current) => ({
											...current,
											meaning: event.target.value,
										}))
									}
								/>
							</Field>
						</div>
						<Field>
							<FieldLabel htmlFor="card-example">Example</FieldLabel>
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
							<FieldLabel htmlFor="card-note">Note</FieldLabel>
							<Input
								id="card-note"
								value={form.note}
								onChange={(event) =>
									setForm((current) => ({
										...current,
										note: event.target.value,
									}))
								}
							/>
						</Field>
					</FieldGroup>
					<div className="mt-3 flex justify-end gap-2">
						{editing ? (
							<Button type="button" variant="ghost" onClick={resetForm}>
								{m["admin.actions.cancel"]()}
							</Button>
						) : null}
						<SubmitButton
							type="submit"
							size="sm"
							pending={createMutation.isPending || updateMutation.isPending}
						>
							{editing
								? m["admin.actions.save"]()
								: m["admin.actions.create"]()}
						</SubmitButton>
					</div>
				</form>
				<Table>
					<TableHeader>
						<TableRow>
							<TableHead>Phrase</TableHead>
							<TableHead>Meaning</TableHead>
							<TableHead>{m["admin.table.actions"]()}</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{(query.data ?? []).map((item) => (
							<TableRow key={item.id}>
								<TableCell>{item.phrase}</TableCell>
								<TableCell>{item.meaning}</TableCell>
								<TableCell>
									<div className="flex items-center gap-2">
										<Button
											type="button"
											variant="outline"
											size="sm"
											onClick={() => {
												setEditing(item);
												setForm({
													phrase: item.phrase,
													meaning: item.meaning,
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
