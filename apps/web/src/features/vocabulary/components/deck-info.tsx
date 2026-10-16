import type { VocabularyDeckDetail } from "@engflex/contracts";
import { useNavigate } from "@tanstack/react-router";
import { Pencil, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
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
import { Textarea } from "#/components/ui/textarea";
import { useDeleteDeck, useUpdateDeck } from "#/features/vocabulary/queries";
import { m } from "#/paraglide/messages";

/** The deck's own information, with edit and delete. */
export function DeckInfo({ deck }: { deck: VocabularyDeckDetail }) {
	const navigate = useNavigate();
	const updateDeck = useUpdateDeck(deck.id);
	const deleteDeck = useDeleteDeck();
	const [editing, setEditing] = useState(false);
	const [confirming, setConfirming] = useState(false);
	const [name, setName] = useState(deck.name);
	const [description, setDescription] = useState(deck.description);
	const [error, setError] = useState<string | null>(null);

	function save() {
		const trimmed = name.trim();
		if (!trimmed) {
			setError(m["vocabulary.decks.required"]());
			return;
		}
		setError(null);
		updateDeck.mutate(
			{ name: trimmed, description: description.trim() },
			{
				onSuccess: () => {
					toast.success(m["vocabulary.deck.updated"]());
					setEditing(false);
				},
				onError: () => toast.error(m["vocabulary.deck.updateFailed"]()),
			},
		);
	}

	function remove() {
		deleteDeck.mutate(deck.id, {
			onSuccess: () => {
				toast.success(m["vocabulary.deck.deleted"]());
				void navigate({ to: APP_ROUTES.VOCABULARY.DECKS.LIST });
			},
			onError: () => toast.error(m["vocabulary.deck.deleteFailed"]()),
		});
	}

	return (
		<div className="surface-card flex flex-col gap-3 p-5">
			<div className="flex flex-wrap items-start justify-between gap-3">
				<div className="min-w-0">
					<h1 className="truncate text-2xl font-bold text-foreground">
						{deck.name}
					</h1>
					{deck.description ? (
						<p className="mt-1 text-sm font-medium text-muted-foreground">
							{deck.description}
						</p>
					) : null}
				</div>
				<div className="flex shrink-0 items-center gap-2">
					<Button
						type="button"
						variant="outline"
						size="sm"
						onClick={() => setEditing(true)}
					>
						<Pencil data-icon="inline-start" />
						{m["vocabulary.deck.edit"]()}
					</Button>
					<Button
						type="button"
						variant="ghost"
						size="sm"
						onClick={() => setConfirming(true)}
					>
						<Trash2 data-icon="inline-start" />
						{m["vocabulary.deck.delete"]()}
					</Button>
				</div>
			</div>
			<p className="text-xs font-bold text-muted-foreground">
				{m["vocabulary.deck.itemCount"]({ count: (deck.items ?? []).length })}
			</p>

			<Dialog open={editing} onOpenChange={setEditing}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{m["vocabulary.deck.editTitle"]()}</DialogTitle>
						<DialogDescription>
							{m["vocabulary.deck.editDescription"]()}
						</DialogDescription>
					</DialogHeader>
					<FieldGroup>
						<Field data-invalid={Boolean(error)}>
							<FieldLabel htmlFor="deck-edit-name">
								{m["vocabulary.decks.nameLabel"]()}
							</FieldLabel>
							<Input
								id="deck-edit-name"
								value={name}
								onChange={(event) => setName(event.target.value)}
							/>
						</Field>
						<Field>
							<FieldLabel htmlFor="deck-edit-description">
								{m["vocabulary.decks.descriptionLabel"]()}
							</FieldLabel>
							<Textarea
								id="deck-edit-description"
								rows={2}
								value={description}
								onChange={(event) => setDescription(event.target.value)}
							/>
						</Field>
						{error ? <FieldError>{error}</FieldError> : null}
					</FieldGroup>
					<DialogFooter>
						<Button
							type="button"
							variant="ghost"
							onClick={() => setEditing(false)}
						>
							{m["common.actions.cancel"]()}
						</Button>
						<SubmitButton
							type="button"
							onClick={save}
							pending={updateDeck.isPending}
						>
							{m["common.actions.save"]()}
						</SubmitButton>
					</DialogFooter>
				</DialogContent>
			</Dialog>

			<Dialog open={confirming} onOpenChange={setConfirming}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{m["vocabulary.deck.deleteTitle"]()}</DialogTitle>
						<DialogDescription>
							{m["vocabulary.deck.deleteBody"]()}
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button
							type="button"
							variant="ghost"
							onClick={() => setConfirming(false)}
						>
							{m["common.actions.cancel"]()}
						</Button>
						<SubmitButton
							type="button"
							variant="destructive"
							onClick={remove}
							pending={deleteDeck.isPending}
						>
							{m["vocabulary.deck.delete"]()}
						</SubmitButton>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</div>
	);
}
