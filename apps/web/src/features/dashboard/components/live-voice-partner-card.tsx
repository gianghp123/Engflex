import { useNavigate } from "@tanstack/react-router";
import { Mic } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { CardShell } from "#/components/common/card-shell";
import { SubmitButton } from "#/components/common/submit-button";
import { Skeleton } from "#/components/ui/skeleton";
import { VOICE_PARTNER_SCENARIO_IDS } from "#/features/dashboard/fixtures";
import { useCreateConversation, useScenarios } from "#/features/voice/queries";
import { m } from "#/paraglide/messages";
import { ScenarioSelectorRow } from "./scenario-selector-row";

export function LiveVoicePartnerCard() {
	const navigate = useNavigate();
	const createConversation = useCreateConversation();
	const [selectedId, setSelectedId] = useState<string>(
		VOICE_PARTNER_SCENARIO_IDS[0],
	);
	const partnerScenarios = useScenarios();
	const rows = (partnerScenarios.data?.items ?? []).filter((scenario) =>
		(VOICE_PARTNER_SCENARIO_IDS as readonly string[]).includes(scenario.id),
	);

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
		<CardShell
			pending={partnerScenarios.isPending}
			className="flex h-full flex-col gap-4 p-5"
			skeleton={
				<>
					<div className="flex flex-col gap-2">
						<Skeleton className="h-5 w-2/3" />
						<Skeleton className="h-4 w-full" />
					</div>
					<div className="flex flex-col gap-2">
						<Skeleton className="h-12 w-full rounded-xl" />
						<Skeleton className="h-12 w-full rounded-xl" />
					</div>
					<Skeleton className="mt-auto h-10 w-full rounded-lg" />
				</>
			}
		>
			<div className="flex items-start justify-between gap-3">
				<div>
					<h3 className="text-base font-bold text-foreground">
						{m["dashboard.voicePartner.title"]()}
					</h3>
					<p className="mt-1 text-[15px] font-medium text-muted-foreground">
						{m["dashboard.voicePartner.description"]()}
					</p>
				</div>
			</div>
			<div className="flex flex-col gap-2">
				{rows.map((scenario) => (
					<ScenarioSelectorRow
						key={scenario.id}
						title={scenario.title}
						status={
							selectedId === scenario.id
								? m["dashboard.voicePartner.selected"]()
								: m["dashboard.voicePartner.ready"]()
						}
						selected={selectedId === scenario.id}
						onSelect={() => setSelectedId(scenario.id)}
					/>
				))}
			</div>
			<SubmitButton
				className="btn btn-primary mt-auto w-full"
				pending={createConversation.isPending}
				onClick={() => {
					void startFreeTalk();
				}}
			>
				<Mic data-icon="inline-start" />
				{m["dashboard.voicePartner.start"]()}
			</SubmitButton>
		</CardShell>
	);
}
