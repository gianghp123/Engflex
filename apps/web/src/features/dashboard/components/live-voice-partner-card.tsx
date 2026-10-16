import { useNavigate } from "@tanstack/react-router";
import { Mic } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { CardShell } from "#/components/common/card-shell";
import { SubmitButton } from "#/components/common/submit-button";
import { Skeleton } from "#/components/ui/skeleton";
import { useCreateConversation, useScenarios } from "#/features/voice/queries";
import { m } from "#/paraglide/messages";
import { ScenarioSelectorRow } from "./scenario-selector-row";

export function LiveVoicePartnerCard() {
	const navigate = useNavigate();
	const createConversation = useCreateConversation();
	const partnerScenarios = useScenarios({ random: true, pageSize: 2 });
	const rows = partnerScenarios.data?.items ?? [];
	// Two random partners each fetch; the picked row stays selected until a
	// refetch replaces the pair, then the first row is the effective pick.
	const [pickedId, setPickedId] = useState<string | null>(null);
	const selectedId = rows.some((row) => row.id === pickedId)
		? pickedId
		: (rows[0]?.id ?? null);

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
						onSelect={() => setPickedId(scenario.id)}
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
