import { createFileRoute } from "@tanstack/react-router";
import { WordDetail } from "#/features/vocabulary/components/word-detail";
import { vocabularyItemQueryOptions } from "#/features/vocabulary/queries";
import { loadOr404 } from "#/lib/router/route-loader";

export const Route = createFileRoute("/_app/vocabulary/words/$itemId")({
	loader: ({ context: { queryClient }, params: { itemId } }) =>
		loadOr404(queryClient, vocabularyItemQueryOptions(itemId)),
	component: WordDetailRoute,
});

function WordDetailRoute() {
	const { itemId } = Route.useParams();
	return <WordDetail itemId={itemId} />;
}
