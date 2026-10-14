import type { Turn } from "@engflex/contracts";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageSplit } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { SurveyCard } from "#/features/voice/components/room/survey-card";
import { PersistedTranscript } from "#/features/voice/components/transcript/persisted-transcript";
import type { AnalyzeControl } from "#/features/voice/components/transcript/turn-feedback";
import { useConversation, useScenario } from "#/features/voice/queries";
import { m } from "#/paraglide/messages";

/** Scene summary: what this session just was. The eyebrow frames it so a
 * bare title + objective never reads as a stray card. */
function SceneCard({ title, body }: { title: string; body: string }) {
	return (
		<section className="surface-card flex flex-col gap-1 p-5">
			<div className="flex items-center gap-3">
				<MoMascot variant="cheer" size={40} />

				<div className="flex flex-col gap-1">
					<p className="text-xs font-bold tracking-wide text-muted-foreground uppercase">
						{m["voice.room.scene.label"]()}
					</p>
					<h2 className="text-base font-bold tracking-tight text-foreground">
						{title}
					</h2>
				</div>
			</div>
			<p className="text-sm text-muted-foreground">{body}</p>
		</section>
	);
}

/**
 * A finished conversation renders its persisted transcript (review state)
 * with per-turn Analyze — no live session, no modal, no redirect. Two panes:
 * the transcript scrolls in the main column while a sticky recap rail
 * anchors the aside: scene card (scenario title + objective, or the
 * free-talk line), honest session stats + the way back, then the
 * post-session survey. Empty sessions skip the split: one centered card with the way
 * back. The room passes `conversationId` (query-backed); the DEV preview
 * passes `turns` fixtures instead.
 */
export function EndedReview({
	conversationId,
	turns,
	scene: sceneOverride,
	durationSec: durationSecOverride,
	analyze,
	onBack,
}: {
	conversationId?: string;
	turns?: Turn[];
	/** Preview-only scene override (fixtures carry no scenarioId). */
	scene?: { title: string; body: string };
	/** Preview-only duration (fixtures carry no durationSec). */
	durationSec?: number;
	analyze?: AnalyzeControl;
	onBack: () => void;
}) {
	const useFixtures = turns !== undefined;
	// Same query key as PersistedTranscript — shared cache, no extra request.
	// Also tells the aside whether the session produced anything worth a survey.
	const conversation = useConversation(conversationId ?? "", {
		enabled: !useFixtures && Boolean(conversationId),
	});
	const sessionTurns = turns ?? conversation.data?.turns ?? [];
	const scenarioId = useFixtures ? undefined : conversation.data?.scenarioId;
	const scenario = useScenario(scenarioId);
	const scene =
		sceneOverride ??
		(scenario.data
			? { title: scenario.data.title, body: scenario.data.objective }
			: undefined);
	const showFreeTalk =
		!scene && !useFixtures && conversation.data?.mode === "free_talk";

	if (sessionTurns.length === 0) {
		return (
			<div className="flex min-h-0 flex-1 flex-col items-center p-4">
				<div className="surface-card flex w-full max-w-xl flex-col items-center gap-4 p-6 text-center">
					<MoMascot variant="confused" size={72} />
					<div className="flex flex-col gap-1">
						<h2 className="text-lg font-bold text-foreground">
							{m["voice.room.review.emptyTitle"]()}
						</h2>
						<p className="text-sm text-muted-foreground">
							{m["voice.room.review.emptyBody"]()}
						</p>
					</div>
					<Button
						type="button"
						className="btn btn-primary w-fit"
						onClick={onBack}
					>
						{m["voice.room.chooseMode"]()}
					</Button>
				</div>
			</div>
		);
	}

	const userTurns = sessionTurns.filter((turn) => turn.role === "user");
	const corrections = userTurns.filter(
		(turn) => (turn.feedback?.payload.spans?.length ?? 0) > 0,
	).length;
	const durationSec = durationSecOverride ?? conversation.data?.durationSec;
	const minutes =
		durationSec == null ? null : Math.max(1, Math.round(durationSec / 60));

	return (
		<div className="flex min-h-0 flex-1 flex-col gap-4 p-4">
			<PageSplit
				main={
					<PersistedTranscript
						conversationId={conversationId}
						turns={turns}
						analyze={analyze}
					/>
				}
				aside={
					<div className="flex flex-col gap-4 lg:sticky lg:top-4">
						{scene ? <SceneCard title={scene.title} body={scene.body} /> : null}
						{!scene && showFreeTalk ? (
							<SceneCard
								title={m["voice.room.scene.freeTitle"]()}
								body={m["voice.room.scene.freeBody"]()}
							/>
						) : null}
						<section className="surface-card flex flex-col gap-4 p-5">
							<div className="flex items-center gap-3">
								<MoMascot variant="cheer" size={40} />
								<p className="text-lg font-bold text-foreground">
									{m["voice.room.recap.title"]()}
								</p>
							</div>
							<div
								className={
									minutes === null
										? "grid grid-cols-2 gap-2"
										: "grid grid-cols-3 gap-2"
								}
							>
								<div className="flex flex-col items-center gap-0.5 rounded-xl border-2 border-border bg-muted px-2 py-3">
									<span className="text-2xl leading-none font-bold text-foreground">
										{userTurns.length}
									</span>
									<span className="text-xs font-bold text-muted-foreground">
										{m["voice.room.recap.exchanges"]()}
									</span>
								</div>
								{minutes === null ? null : (
									<div className="flex flex-col items-center gap-0.5 rounded-xl border-2 border-border bg-muted px-2 py-3">
										<span className="text-2xl leading-none font-bold text-foreground">
											{minutes}
										</span>
										<span className="text-xs font-bold text-muted-foreground">
											{m["voice.room.recap.minutes"]()}
										</span>
									</div>
								)}
								<div className="flex flex-col items-center gap-0.5 rounded-xl border-2 border-border bg-muted px-2 py-3">
									<span className="text-2xl leading-none font-bold text-foreground">
										{corrections}
									</span>
									<span className="text-xs font-bold text-muted-foreground">
										{m["voice.room.recap.corrections"]()}
									</span>
								</div>
							</div>
							<Button
								type="button"
								className="btn btn-primary w-full"
								onClick={onBack}
							>
								{m["voice.room.chooseMode"]()}
							</Button>
						</section>
						<SurveyCard />
					</div>
				}
			/>
		</div>
	);
}
