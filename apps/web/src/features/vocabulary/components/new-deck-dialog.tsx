import { MaxUserDecks } from "@engflex/contracts";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
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
import { useCreateDeck, useDecks } from "#/features/vocabulary/queries";
import { ApiError } from "#/lib/api/api";
import { m } from "#/paraglide/messages";

export function NewDeckDialog({
	open: controlledOpen,
	onOpenChange,
}: {
	/** Controlled mode: the caller owns `open`; no trigger is rendered. */
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
}) {
	const createDeck = useCreateDeck();
	const [selfOpen, setSelfOpen] = useState(false);
	const controlled = controlledOpen !== undefined;
	const open = controlledOpen ?? selfOpen;
	const setOpen = onOpenChange ?? setSelfOpen;
	const [name, setName] = useState("");
	const [description, setDescription] = useState("");
	const [error, setError] = useState<string | null>(null);
	const decks = useDecks(0);
	const typedName = name.trim().toLowerCase();
	const nameTaken =
		typedName.length > 0 &&
		(decks.data?.items ?? []).some(
			(deck) => deck.name.trim().toLowerCase() === typedName,
		);
	// Curated decks (no userId) are shared and never count toward the cap.
	const ownedCount = (decks.data?.items ?? []).filter(
		(deck) => deck.userId,
	).length;
	const atDeckLimit = ownedCount >= MaxUserDecks;

	function submit() {
		const trimmed = name.trim();
		if (!trimmed) {
			setError(m["vocabulary.decks.required"]());
			return;
		}
		setError(null);
		createDeck.mutate(
			{ name: trimmed, description: description.trim() },
			{
				onSuccess: () => {
					toast.success(m["vocabulary.decks.created"]());
					setOpen(false);
					setName("");
					setDescription("");
				},
				onError: (err) => {
					if (err instanceof ApiError && err.status === 409) {
						setError(m["vocabulary.decks.nameTaken"]());
						return;
					}
					// 422 is the server's deck cap (the client precheck usually
					// beats it; a race lands here).
					if (err instanceof ApiError && err.status === 422) {
						setError(m["vocabulary.decks.limitReached"]({ max: MaxUserDecks }));
						return;
					}
					toast.error(m["vocabulary.decks.createFailed"]());
				},
			},
		);
	}

	return (
		<>
			{controlled ? null : (
				<Button
					type="button"
					className="btn btn-accent"
					onClick={() => setOpen(true)}
				>
					<Plus data-icon="inline-start" />
					{m["vocabulary.decks.new"]()}
				</Button>
			)}
			<Dialog open={open} onOpenChange={setOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{m["vocabulary.decks.createTitle"]()}</DialogTitle>
						<DialogDescription>
							{m["vocabulary.decks.createDescription"]()}
						</DialogDescription>
					</DialogHeader>
					<FieldGroup>
						{atDeckLimit ? (
							<p className="text-xs font-bold text-muted-foreground">
								{m["vocabulary.decks.limitReached"]({ max: MaxUserDecks })}
							</p>
						) : null}
						<Field data-invalid={Boolean(error)}>
							<FieldLabel htmlFor="deck-name">
								{m["vocabulary.decks.nameLabel"]()}
							</FieldLabel>
							<Input
								id="deck-name"
								value={name}
								onChange={(event) => setName(event.target.value)}
							/>
							{nameTaken ? (
								<p className="text-xs font-bold text-muted-foreground">
									{m["vocabulary.decks.nameTaken"]()}
								</p>
							) : null}
						</Field>
						<Field>
							<FieldLabel htmlFor="deck-description">
								{m["vocabulary.decks.descriptionLabel"]()}
							</FieldLabel>
							<Textarea
								id="deck-description"
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
							onClick={() => setOpen(false)}
						>
							{m["common.actions.cancel"]()}
						</Button>
						<SubmitButton
							type="button"
							onClick={submit}
							pending={createDeck.isPending}
							disabled={atDeckLimit}
						>
							{m["vocabulary.decks.create"]()}
						</SubmitButton>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	);
}
