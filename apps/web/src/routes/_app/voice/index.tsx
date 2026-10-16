import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { TipCard } from "#/components/common/tip-card";
import { HistoryPreview } from "#/features/voice/components/history/history-preview";
import {
	ModeCard,
	VOICE_MODES,
} from "#/features/voice/components/mode/mode-card";
import { useCreateConversation } from "#/features/voice/queries";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/voice/")({
	component: VoiceModesPage,
});

function VoiceModesPage() {
	useBreadcrumbs([{ label: m["nav.item.voice"]() }]);
	const navigate = useNavigate();
	const createConversation = useCreateConversation();

	async function startFreeTalk() {
		if (createConversation.isPending) return;
		try {
			const conversation = await createConversation.mutateAsync({
				mode: "free_talk",
			});
			await navigate({
				to: APP_ROUTES.VOICE.ROOM,
				params: { conversationId: conversation.id },
			});
		} catch {
			toast.error(m["voice.create.failed"]());
		}
	}

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="rest" size={84} />,
				title: m["voice.modes.title"](),
				description: m["voice.modes.subtitle"](),
			}}
		>
			<div className="grid items-start gap-4 lg:grid-cols-2">
				{VOICE_MODES.map((mode) => (
					<ModeCard
						key={mode}
						mode={mode}
						pending={
							mode === "spontaneous" ? createConversation.isPending : false
						}
						onStart={() => {
							if (mode === "spontaneous") {
								void startFreeTalk();
								return;
							}
							navigate({ to: APP_ROUTES.VOICE.SCENARIOS });
						}}
					/>
				))}
			</div>
			<HistoryPreview />
			<TipCard title={m["voice.tip.title"]()} body={m["voice.tip.body"]()} />
		</PageLayout>
	);
}
