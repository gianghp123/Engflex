import type { VocabularyItem } from "@engflex/contracts";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { CardShell } from "#/components/common/card-shell";
import { DialogDescription } from "#/components/ui/dialog";
import { Input } from "#/components/ui/input";
import { Skeleton } from "#/components/ui/skeleton";
import { posLabel } from "#/features/vocabulary/components/pos-label";
import {
	useAddExistingWordToDeck,
	useDeck,
	useMyVocabulary,
} from "#/features/vocabulary/queries";
import { m } from "#/paraglide/messages";

/**
 * The "from my words" mode: pick words the reader already has and put them in
 * this deck.
 *
 * The deck is read here rather than passed in, because the landing's deck
 * objects carry a bounded *preview* — the exclusion set has to be the real
 * membership. Mounting only when this mode is opened keeps that read on demand.
 */
export function PickExistingWords({ deckId }: { deckId: string }) {
	const addToDeck = useAddExistingWordToDeck(deckId);
	const library = useMyVocabulary();
	const deck = useDeck(deckId);
	const [search, setSearch] = useState("");

	const inDeck = new Set((deck.data?.items ?? []).map((row) => row.itemId));
	const candidates = (library.data?.items ?? []).filter(
		(item: VocabularyItem) => !inDeck.has(item.id),
	);
	const needle = search.trim().toLowerCase();
	const visible = needle
		? candidates.filter((item) => item.term.toLowerCase().includes(needle))
		: candidates;
	const pending = library.isPending || deck.isPending;

	function add(itemId: string) {
		addToDeck.mutate(itemId, {
			onSuccess: () => toast.success(m["vocabulary.deck.added"]()),
			onError: () => toast.error(m["vocabulary.deck.addFailed"]()),
		});
	}

	return (
		<>
			<DialogDescription className="font-medium">
				{m["vocabulary.deck.pickerDescription"]()}
			</DialogDescription>
			<Input
				value={search}
				onChange={(event) => setSearch(event.target.value)}
				placeholder={m["vocabulary.deck.pickerSearch"]()}
			/>
			<div className="flex max-h-80 flex-col gap-2 overflow-y-auto">
				{pending ? (
					<CardShell
						pending
						className="p-3"
						skeleton={<Skeleton className="h-10 w-full" />}
					/>
				) : library.isError ? (
					<p className="text-sm font-medium text-muted-foreground">
						{m["vocabulary.deck.pickerFailed"]()}
					</p>
				) : visible.length === 0 ? (
					<p className="text-sm font-medium text-muted-foreground">
						{m["vocabulary.deck.pickerEmpty"]()}
					</p>
				) : (
					visible.map((item) => (
						<button
							key={item.id}
							type="button"
							onClick={() => add(item.id)}
							className="surface-card flex items-center justify-between gap-3 p-3 text-left hover:border-primary"
						>
							<span className="min-w-0">
								<span className="block truncate font-bold text-foreground">
									{item.term}
								</span>
								<span className="block truncate text-xs font-medium text-muted-foreground">
									{posLabel(item.partOfSpeech)} · {item.definition}
								</span>
							</span>
							<Plus className="size-4 shrink-0" />
						</button>
					))
				)}
			</div>
		</>
	);
}
