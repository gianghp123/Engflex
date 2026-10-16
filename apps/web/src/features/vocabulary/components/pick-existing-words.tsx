import { MaxDeckWords } from "@engflex/contracts";
import { Plus, Search } from "lucide-react";
import { type FormEvent, type ReactNode, useState } from "react";
import { toast } from "sonner";
import { CardShell } from "#/components/common/card-shell";
import { Button } from "#/components/ui/button";
import { DialogDescription } from "#/components/ui/dialog";
import { Input } from "#/components/ui/input";
import { Skeleton } from "#/components/ui/skeleton";
import { posLabel } from "#/features/vocabulary/components/pos-label";
import {
	useAddExistingWordToDeck,
	useDeck,
	useLibrarySearch,
} from "#/features/vocabulary/queries";
import { ApiError } from "#/lib/api";
import { m } from "#/paraglide/messages";

/**
 * The searched term inside `text`, wrapped in <mark>. The API matches the term
 * and the definition, so both wear the highlight.
 */
function Highlight({ text, needle }: { text: string; needle: string }) {
	const haystack = text.toLowerCase();
	const query = needle.toLowerCase();
	const parts: ReactNode[] = [];
	let cursor = 0;
	let at = haystack.indexOf(query);
	while (at !== -1) {
		parts.push(text.slice(cursor, at));
		parts.push(
			<mark
				key={at}
				className="rounded-xs bg-secondary text-secondary-foreground"
			>
				{text.slice(at, at + needle.length)}
			</mark>,
		);
		cursor = at + needle.length;
		at = haystack.indexOf(query, cursor);
	}
	parts.push(text.slice(cursor));
	return <>{parts}</>;
}

/**
 * The "from my words" mode: pick words the reader already has and put them in
 * this deck.
 *
 * Search is explicit: Enter or the search button runs the API's `%term%`
 * search (term or definition) — typing alone fetches nothing. The deck is read
 * here rather than passed in, because the landing's deck objects carry a
 * bounded *preview* — the exclusion set has to be the real membership.
 */
export function PickExistingWords({ deckId }: { deckId: string }) {
	const addToDeck = useAddExistingWordToDeck(deckId);
	const deck = useDeck(deckId);
	const [draft, setDraft] = useState("");
	const [submitted, setSubmitted] = useState<string | null>(null);
	const results = useLibrarySearch(submitted);

	const inDeck = new Set((deck.data?.items ?? []).map((row) => row.itemId));
	const matches = results.data?.items ?? [];
	const candidates = matches.filter((item) => !inDeck.has(item.id));
	const pending = submitted !== null && (results.isPending || deck.isPending);
	const failed = results.isError || deck.isError;
	const needle = submitted ?? "";

	function search(event: FormEvent) {
		event.preventDefault();
		const term = draft.trim();
		if (!term) return;
		setSubmitted(term);
	}

	function add(itemId: string) {
		addToDeck.mutate(itemId, {
			onSuccess: () => toast.success(m["vocabulary.deck.added"]()),
			onError: (err) => {
				// 422 is the server's per-deck cap on a new membership.
				if (err instanceof ApiError && err.status === 422) {
					toast.error(m["vocabulary.deck.limitReached"]({ max: MaxDeckWords }));
					return;
				}
				toast.error(m["vocabulary.deck.addFailed"]());
			},
		});
	}

	return (
		<>
			<DialogDescription className="font-medium">
				{m["vocabulary.deck.pickerDescription"]()}
			</DialogDescription>
			<form onSubmit={search} className="flex items-center gap-2">
				<Input
					value={draft}
					onChange={(event) => setDraft(event.target.value)}
					placeholder={m["vocabulary.deck.pickerSearch"]()}
				/>
				<Button
					type="submit"
					variant="outline"
					size="icon"
					disabled={!draft.trim()}
					aria-label={m["vocabulary.deck.pickerSearch"]()}
				>
					<Search />
				</Button>
			</form>
			<div className="flex max-h-80 flex-col gap-2 overflow-y-auto">
				{submitted === null ? (
					<p className="text-sm font-medium text-muted-foreground">
						{m["vocabulary.deck.pickerHint"]()}
					</p>
				) : pending ? (
					<CardShell
						pending
						className="p-3"
						skeleton={<Skeleton className="h-10 w-full" />}
					/>
				) : failed ? (
					<p className="text-sm font-medium text-muted-foreground">
						{m["vocabulary.deck.pickerFailed"]()}
					</p>
				) : candidates.length === 0 ? (
					<p className="text-sm font-medium text-muted-foreground">
						{matches.length === 0
							? m["vocabulary.deck.pickerNoResults"]({ term: needle })
							: m["vocabulary.deck.pickerEmpty"]()}
					</p>
				) : (
					candidates.map((item) => (
						<Button
							key={item.id}
							variant="outline"
							onClick={() => add(item.id)}
							className="flex items-center justify-between gap-3 text-left hover:border-primary"
						>
							<span className="min-w-0">
								<span className="block truncate font-bold text-foreground">
									<Highlight text={item.term} needle={needle} />
								</span>
								<span className="block truncate text-xs font-medium text-muted-foreground">
									{posLabel(item.partOfSpeech)} ·{" "}
									<Highlight text={item.definition} needle={needle} />
								</span>
							</span>
							<Plus className="size-4 shrink-0" />
						</Button>
					))
				)}
			</div>
		</>
	);
}
