import { Plus } from "lucide-react";
import type { ComponentProps } from "react";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "#/components/ui/tabs";
import { AddWordForm } from "#/features/vocabulary/components/add-word-form";
import { PickExistingWords } from "#/features/vocabulary/components/pick-existing-words";
import { m } from "#/paraglide/messages";

type ButtonProps = ComponentProps<typeof Button>;

/**
 * The one way to add a word, wherever the flow starts.
 *
 * With a deck in context it offers two modes: adding a custom word — the
 * default, because that is the case the deck has no equivalent of — and putting
 * words already in the library into this deck. Without a deck (the All words
 * page) the second mode is meaningless, so the form stands alone.
 *
 * `size`/`variant` reach the trigger: it is the page's primary action where it
 * is the only one (All words, a deck page), and a quiet control inside a deck
 * section's header, where the page's primary is the New-deck card.
 */
export function AddWordDialog({
	deckId,
	size = "default",
	variant = "default",
	className,
}: {
	deckId?: string;
	size?: ButtonProps["size"];
	variant?: ButtonProps["variant"];
	/** Styling for the visible control — the trigger button. */
	className?: string;
}) {
	const [open, setOpen] = useState(false);
	const close = () => setOpen(false);

	return (
		<>
			<Button
				type="button"
				size={size}
				variant={variant}
				className={className}
				onClick={() => setOpen(true)}
			>
				<Plus data-icon="inline-start" />
				{m["vocabulary.addWord.button"]()}
			</Button>
			<Dialog open={open} onOpenChange={setOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{m["vocabulary.addWord.title"]()}</DialogTitle>
					</DialogHeader>
					{deckId ? (
						<Tabs defaultValue="word">
							<TabsList className="w-full">
								<TabsTrigger value="word">
									{m["vocabulary.addWord.tabs.word"]()}
								</TabsTrigger>
								<TabsTrigger value="library">
									{m["vocabulary.deck.addFromLibrary"]()}
								</TabsTrigger>
							</TabsList>
							<TabsContent value="word" className="flex flex-col gap-4">
								<AddWordForm deckId={deckId} onClose={close} />
							</TabsContent>
							<TabsContent value="library" className="flex flex-col gap-4">
								<PickExistingWords deckId={deckId} />
							</TabsContent>
						</Tabs>
					) : (
						<AddWordForm onClose={close} />
					)}
				</DialogContent>
			</Dialog>
		</>
	);
}
