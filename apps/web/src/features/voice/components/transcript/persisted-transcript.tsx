import type { Turn } from "@engflex/contracts";
import { CardShell } from "#/components/common/card-shell";
import { LanguagePicker } from "#/components/common/language-picker";
import { MoMascot } from "#/components/common/mo-mascot";
import { Badge } from "#/components/ui/badge";
import {
	Empty,
	EmptyDescription,
	EmptyHeader,
	EmptyTitle,
} from "#/components/ui/empty";
import { ScrollArea } from "#/components/ui/scroll-area";
import { Skeleton } from "#/components/ui/skeleton";
import {
	type TranscriptRow,
	TranscriptRows,
} from "#/features/voice/components/transcript/transcript-panel";
import type { AnalyzeControl } from "#/features/voice/components/transcript/turn-feedback";
import { TurnFeedbackDisclosure } from "#/features/voice/components/transcript/turn-feedback-disclosure";
import { useAnalyzeTurn, useConversation } from "#/features/voice/queries";
import { useExplanationLanguage } from "#/hooks/use-explanation-language";
import { cn } from "#/lib/utils";
import { m } from "#/paraglide/messages";

/**
 * Finished-session review view: pure rendering over resolved turns. No
 * network, no provider — the container below injects the data, whether live
 * or fixtures. Tweak bubble/feedback layout here.
 */
function PersistedTranscriptView({
	turns,
	analyzePending,
	analyzeFailed,
	onAnalyze,
	pending = false,
	className,
}: {
	turns: Turn[];
	analyzePending: boolean;
	analyzeFailed: boolean;
	onAnalyze: (position: number) => void;
	pending?: boolean;
	className?: string;
}) {
	const { asking, withLanguage, confirm, dismiss, reopen } =
		useExplanationLanguage("analyze");
	const rows: TranscriptRow[] = turns.map((turn) => ({
		key: turn.id,
		isUser: turn.role === "user",
		text: turn.text,
		accessory:
			turn.role === "user" ? (
				<TurnFeedbackDisclosure
					feedback={turn.feedback?.payload}
					utterance={turn.text}
					pending={analyzePending}
					failed={analyzeFailed}
					onAnalyze={() => withLanguage(() => onAnalyze(turn.position))}
				/>
			) : undefined,
	}));

	return (
		<CardShell
			pending={pending}
			className={cn("flex h-full min-h-0 flex-col gap-6 py-6", className)}
			skeleton={
				<>
					<div className="px-6">
						<Skeleton className="h-5 w-32" />
					</div>
					<div className="flex flex-col gap-4 px-6 pb-6">
						<Skeleton className="h-14 w-3/4" />
						<Skeleton className="h-14 w-2/3 self-end" />
						<Skeleton className="h-14 w-3/4" />
					</div>
				</>
			}
		>
			<div className="flex items-center gap-2 px-6 text-base leading-none font-semibold">
				{m["voice.room.transcript.title"]()}
				<Badge variant="secondary">{rows.length}</Badge>
				{/* One control per transcript, not per turn: it belongs to the
				    surface that sends the analyze requests. */}
				<span className="ml-auto">
					<LanguagePicker
						open={asking}
						onOpenChange={(next) => {
							if (next) reopen();
							else dismiss();
						}}
						onConfirm={confirm}
					/>
				</span>
			</div>
			<div className="min-h-0 flex-1 px-6 pb-6">
				<ScrollArea className="h-full max-h-[55vh] min-h-0 overflow-x-clip pr-3 lg:max-h-none">
					{rows.length === 0 ? (
						<Empty className="items-center py-8 text-center">
							<MoMascot variant="confused" size={72} />
							<EmptyHeader>
								<EmptyTitle>{m["voice.room.review.emptyTitle"]()}</EmptyTitle>
								<EmptyDescription>
									{m["voice.room.review.emptyBody"]()}
								</EmptyDescription>
							</EmptyHeader>
						</Empty>
					) : (
						<TranscriptRows rows={rows} />
					)}
				</ScrollArea>
			</div>
		</CardShell>
	);
}

/**
 * Finished-session review: one component for the room and the DEV preview.
 * TanStack hooks only — no Pipecat provider needed, so it mounts outside
 * PipecatAppBase. The room passes `conversationId`; the preview passes
 * `turns` fixtures (fetch skipped) plus a noop analyze control.
 */
export function PersistedTranscript({
	conversationId,
	turns: turnsOverride,
	analyze: analyzeOverride,
	className,
}: {
	conversationId?: string;
	turns?: Turn[];
	analyze?: AnalyzeControl;
	className?: string;
}) {
	const useFixtures = turnsOverride !== undefined;
	const conversation = useConversation(conversationId ?? "", {
		enabled: !useFixtures,
	});
	const analyzeLive = useAnalyzeTurn(conversationId ?? "");
	const stored = turnsOverride ?? conversation.data?.turns ?? [];
	const turns = stored;

	return (
		<PersistedTranscriptView
			turns={turns}
			analyzePending={analyzeOverride?.pending ?? analyzeLive.isPending}
			analyzeFailed={analyzeOverride?.failed ?? analyzeLive.isError}
			onAnalyze={
				analyzeOverride?.onAnalyze ??
				((position) => analyzeLive.mutate(position))
			}
			pending={!useFixtures && conversation.isPending}
			className={className}
		/>
	);
}
