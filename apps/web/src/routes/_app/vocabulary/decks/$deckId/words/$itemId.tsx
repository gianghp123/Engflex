import { createFileRoute } from "@tanstack/react-router";
import { WordDetail } from "#/features/vocabulary/components/word-detail";
import {
	deckQueryOptions,
	useDeck,
	vocabularyItemQueryOptions,
} from "#/features/vocabulary/queries";
import { loadOr404 } from "#/lib/router/route-loader";

export const Route = createFileRoute(
	"/_app/vocabulary/decks/$deckId/words/$itemId",
)({
	loader: ({ context: { queryClient }, params: { deckId, itemId } }) =>
		Promise.all([
			// The deck is loaded for its name in the breadcrumb trail; it is
			// already cached when the reader came through the deck page.
			loadOr404(queryClient, deckQueryOptions(deckId)),
			loadOr404(queryClient, vocabularyItemQueryOptions(itemId)),
		]),
	component: DeckWordDetailRoute,
});

function DeckWordDetailRoute() {
	const { deckId, itemId } = Route.useParams();
	const deck = useDeck(deckId);
	return (
		<WordDetail itemId={itemId} deckId={deckId} deckName={deck.data?.name} />
	);
}
