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
import { useCreateDeck } from "#/features/vocabulary/queries";
import { m } from "#/paraglide/messages";

export function NewDeckDialog() {
	const createDeck = useCreateDeck();
	const [open, setOpen] = useState(false);
	const [name, setName] = useState("");
	const [description, setDescription] = useState("");
	const [error, setError] = useState<string | null>(null);

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
				onError: () => toast.error(m["vocabulary.decks.createFailed"]()),
			},
		);
	}

	return (
		<>
			<Button
				type="button"
				className="btn btn-accent"
				onClick={() => setOpen(true)}
			>
				<Plus data-icon="inline-start" />
				{m["vocabulary.decks.new"]()}
			</Button>
			<Dialog open={open} onOpenChange={setOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{m["vocabulary.decks.createTitle"]()}</DialogTitle>
						<DialogDescription>
							{m["vocabulary.decks.createDescription"]()}
						</DialogDescription>
					</DialogHeader>
					<FieldGroup>
						<Field data-invalid={Boolean(error)}>
							<FieldLabel htmlFor="deck-name">
								{m["vocabulary.decks.nameLabel"]()}
							</FieldLabel>
							<Input
								id="deck-name"
								value={name}
								onChange={(event) => setName(event.target.value)}
							/>
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
						>
							{m["vocabulary.decks.create"]()}
						</SubmitButton>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	);
}
