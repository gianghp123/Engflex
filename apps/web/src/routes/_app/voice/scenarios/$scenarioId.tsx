import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { ScenarioDetailBody } from "#/features/voice/components/scenario-detail/scenario-detail-body";
import { ScenarioDetailHero } from "#/features/voice/components/scenario-detail/scenario-detail-hero";
import { ScenarioDetailSide } from "#/features/voice/components/scenario-detail/scenario-detail-side";
import {
	scenarioQueryOptions,
	useCreateConversation,
	useScenario,
} from "#/features/voice/queries";
import { loadOr404 } from "#/lib/route-loader";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/voice/scenarios/$scenarioId")({
	// Preload here: loader notFound() renders the 404 page; the component no
	// longer special-cases 404s (see below).
	loader: ({ context: { queryClient }, params: { scenarioId } }) =>
		loadOr404(queryClient, scenarioQueryOptions(scenarioId)),
	component: ScenarioDetailPage,
});

/** Server-driven detail: badges, partner, and content come from the API. */
function ScenarioDetailPage() {
	useBreadcrumbs([
		{ label: m["nav.item.voice"](), to: APP_ROUTES.VOICE.LIST },
		{ label: m["voice.crumb.scenarios"](), to: APP_ROUTES.VOICE.SCENARIOS },
		{ label: m["voice.crumb.detail"]() },
	]);
	const { scenarioId } = Route.useParams();
	const navigate = useNavigate();
	const scenarioQuery = useScenario(scenarioId);
	const createConversation = useCreateConversation();

	async function startRoleplay(id: string) {
		if (createConversation.isPending) return;
		try {
			const conversation = await createConversation.mutateAsync({
				mode: "roleplay",
				scenarioId: id,
			});
			await navigate({
				to: APP_ROUTES.VOICE.ROOM,
				params: { conversationId: conversation.id },
			});
		} catch {
			toast.error(m["voice.create.failed"]());
		}
	}

	if (scenarioQuery.isPending) {
		return (
			<PageLayout
				hero={{
					icon: <MoMascot variant="thinking" size={64} />,
					title: m["voice.detail.loading"](),
					description: "",
				}}
			>
				<div />
			</PageLayout>
		);
	}

	// A failed fetch is an error, not a missing scenario (see loader above:
	// genuine 404s never reach the component).
	if (scenarioQuery.isError) throw scenarioQuery.error;
	const scenario = scenarioQuery.data;
	if (!scenario) return null;
	return (
		<PageLayout
			action={
				<BackButton
					to={APP_ROUTES.VOICE.SCENARIOS}
					label={m["voice.detail.back"]()}
				/>
			}
		>
			<ScenarioDetailHero
				scenario={scenario}
				pending={createConversation.isPending}
				onStart={() => {
					void startRoleplay(scenario.id);
				}}
			/>
			<div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
				<div className="lg:col-span-7">
					<ScenarioDetailBody scenario={scenario} />
				</div>
				<div className="lg:col-span-5">
					<ScenarioDetailSide scenario={scenario} />
				</div>
			</div>
		</PageLayout>
	);
}
