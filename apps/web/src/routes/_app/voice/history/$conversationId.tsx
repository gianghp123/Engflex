import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { EndedReview } from "#/features/voice/components/room/ended-review";
import { conversationQueryOptions } from "#/features/voice/queries";
import { loadOr404 } from "#/lib/router/route-loader";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/voice/history/$conversationId")({
	loader: ({ context: { queryClient }, params: { conversationId } }) =>
		loadOr404(queryClient, conversationQueryOptions(conversationId)),
	component: HistorySessionPage,
});

/** A finished session's review, months later: same EndedReview the room
 *  shows at the end of a session, minus the survey, with history back. */
function HistorySessionPage() {
	useBreadcrumbs([
		{ label: m["nav.item.voice"](), to: APP_ROUTES.VOICE.LIST },
		{ label: m["voice.crumb.history"](), to: APP_ROUTES.VOICE.HISTORY.LIST },
		{ label: m["voice.crumb.session"]() },
	]);
	const { conversationId } = Route.useParams();
	const navigate = useNavigate();
	return (
		<EndedReview
			conversationId={conversationId}
			variant="history"
			backLabel={m["voice.history.backToList"]()}
			onBack={() => {
				void navigate({ to: APP_ROUTES.VOICE.HISTORY.LIST });
			}}
		/>
	);
}
