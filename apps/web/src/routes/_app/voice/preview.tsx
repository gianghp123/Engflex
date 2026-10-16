import type { Turn } from "@engflex/contracts";
import { createFileRoute } from "@tanstack/react-router";
import { useRef, useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { InDevelopment } from "#/components/common/in-development";
import { MO_VARIANTS, MoMascot } from "#/components/common/mo-mascot";
import { Button } from "#/components/ui/button";
import {
	type RoomSession,
	VoiceRoom,
} from "#/features/voice/components/room/voice-room";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/voice/preview")({
	validateSearch: (search: Record<string, unknown>) => ({
		view: search.view === "prod" ? ("prod" as const) : undefined,
	}),
	component: VoicePreviewPage,
});

const now = new Date().toISOString();

/** Fixture turns so the real room can be tweaked without a session. The
analyzed learner turns cover every feedback case: incorrect + partial +
both alternatives, awkward, off-topic, and clean. */
const PREVIEW_TURNS: Turn[] = [
	{
		id: "preview-1",
		position: 1,
		role: "ai",
		text: "Hi! What did you do last weekend?",
		wasInterrupted: false,
		createdAt: now,
	},
	{
		id: "preview-2",
		position: 2,
		role: "user",
		text: "I didn't went to the beach yesterday. I really like seafood.",
		wasInterrupted: false,
		feedback: {
			payload: {
				corrected: "I didn't go to the beach yesterday. I really like seafood.",
				spans: [
					{
						text: "didn't went",
						occurrence: 1,
						status: "incorrect",
						correction: "didn't go",
						reason: "Use the base form after did or didn't.",
					},
				],
				relevance: {
					status: "partially_relevant",
					reason:
						"The first sentence answers the question, but the second shifts to a different topic.",
				},
				alternatives: {
					language: {
						text: "I spent yesterday at the beach.",
						reason: "More natural phrasing of the same meaning.",
					},
					contextual: {
						text: "I went to the beach yesterday and had some great seafood.",
						reason:
							"Answers the question, fixes the grammar, and connects the detail.",
					},
				},
				tip: "Use the base form of the verb after did or didn't.",
			},
		},
		createdAt: now,
	},
	{
		id: "preview-3",
		position: 3,
		role: "ai",
		text: "Sounds fun. How did the trip go overall?",
		wasInterrupted: false,
		createdAt: now,
	},
	{
		id: "preview-4",
		position: 4,
		role: "user",
		text: "We shipped on time, but I struggled to explain the delay in the API migration.",
		wasInterrupted: false,
		feedback: {
			payload: {
				corrected: "We shipped on time, but I struggled to explain the delay.",
				spans: [
					{
						text: "struggled to explain",
						occurrence: 1,
						status: "awkward",
						correction: "had trouble explaining",
						reason: "More natural verb choice.",
					},
				],
				relevance: { status: "relevant", reason: undefined },
				alternatives: {
					language: {
						text: "We shipped on time, but justifying the delay was difficult.",
						reason: "Leads with the outcome.",
					},
					contextual: undefined,
				},
				tip: "Lead with the impact, then the cause.",
			},
		},
		createdAt: now,
	},
	{
		id: "preview-5",
		position: 5,
		role: "ai",
		text: "Did you swim while you were there?",
		wasInterrupted: false,
		createdAt: now,
	},
	{
		id: "preview-6",
		position: 6,
		role: "user",
		text: "My favorite movie is Interstellar.",
		wasInterrupted: false,
		feedback: {
			payload: {
				corrected: "My favorite movie is Interstellar.",
				spans: [],
				relevance: {
					status: "off_topic",
					reason: "This does not answer the question about swimming.",
				},
				alternatives: {
					language: undefined,
					contextual: {
						text: "No, I just watched Interstellar at the hotel instead.",
						reason: "Answers the question while keeping your idea.",
					},
				},
				tip: "Answer the main question directly before adding unrelated information.",
			},
		},
		createdAt: now,
	},
	{
		id: "preview-7",
		position: 7,
		role: "ai",
		text: "Got it. Anything else from the weekend?",
		wasInterrupted: false,
		createdAt: now,
	},
	{
		id: "preview-8",
		position: 8,
		role: "user",
		text: "I went home and rested.",
		wasInterrupted: false,
		feedback: {
			payload: {
				corrected: "I went home and rested.",
				spans: [],
				relevance: { status: "relevant", reason: undefined },
				alternatives: { language: undefined, contextual: undefined },
				tip: "Clean and direct — keep it up.",
			},
		},
		createdAt: now,
	},
];

const FIXTURE_CORRECTION_TEXT =
	"We shipped on time, but I struggled to explain the delay in the API migration.";

type Phase = "live" | "ended" | "loading" | "failed";

/**
 * Preview injector: mock session state in, the SAME VoiceRoom out. No
 * conversation, no engine, no network, no mic permission — the idle base
 * inside VoiceRoom only provides provider context. 404s in production via
 * beforeLoad above.
 */
function VoicePreviewPage() {
	// ?view=prod forces the production face while running dev.
	const { view } = Route.useSearch();
	if (!import.meta.env.DEV || view === "prod") {
		return (
			<div className="flex min-h-0 flex-1 flex-col gap-4 p-4">
				<InDevelopment />
			</div>
		);
	}
	return <PreviewRoom />;
}

function PreviewRoom() {
	useBreadcrumbs([
		{ label: m["nav.item.voice"](), to: APP_ROUTES.VOICE.LIST },
		{ label: m["voice.crumb.room"]() },
		{ label: "preview" },
	]);
	const [phase, setPhase] = useState<Phase>("live");
	const [source, setSource] = useState<"live" | "saved">("live");
	const [connectionLost, setConnectionLost] = useState(false);
	const [errorDetail, setErrorDetail] = useState<string | null>(null);
	const [endPending, setEndPending] = useState(false);
	const userDisconnecting = useRef(false);
	const noop = () => {};

	const session: RoomSession = {
		ready: phase !== "loading",
		query:
			phase === "loading" ? "pending" : phase === "failed" ? "error" : "ok",
		ended: phase === "ended",
		endPending,
		connectionLost,
		errorDetail,
		onExit: noop,
		onBack: noop,
		onEnd: noop,
		onConnectionLost: () => setConnectionLost(true),
		onRuntimeError: (text) => setErrorDetail(text),
		isUserDisconnectingRef: userDisconnecting,
	};

	return (
		<div className="flex min-h-0 flex-1 flex-col gap-4">
			<div className="flex flex-wrap items-center gap-3 px-4 pt-4">
				<p className="text-xs text-muted-foreground">
					DEV preview — the same VoiceRoom on mock state. The pencil opens the
					same correction modal on local state; Send runs the real
					corrected-bubble derivation.
				</p>
				<div className="surface-card flex flex-wrap items-end gap-4 p-4">
					{MO_VARIANTS.map((variant) => (
						<span key={variant} className="flex flex-col items-center gap-1">
							<MoMascot variant={variant} size={72} />
							<span className="chip px-2 py-0 text-[11px]">{variant}</span>
						</span>
					))}
				</div>
				<fieldset className="flex gap-1">
					<legend className="sr-only">Room phase</legend>
					{(["live", "ended", "loading", "failed"] as const).map((option) => (
						<Button
							key={option}
							type="button"
							variant={phase === option ? "default" : "outline"}
							size="sm"
							onClick={() => setPhase(option)}
						>
							{option}
						</Button>
					))}
				</fieldset>
				{phase === "live" && (
					<>
						<fieldset className="flex gap-1">
							<legend className="sr-only">Transcript source</legend>
							{(["live", "saved"] as const).map((option) => (
								<Button
									key={option}
									type="button"
									variant={source === option ? "default" : "outline"}
									size="sm"
									onClick={() => setSource(option)}
								>
									{option === "live" ? "Live transcript" : "Saved review"}
								</Button>
							))}
						</fieldset>
						<Button
							type="button"
							variant={connectionLost ? "default" : "outline"}
							size="sm"
							onClick={() => setConnectionLost((v) => !v)}
						>
							Simulate connection lost
						</Button>
						<Button
							type="button"
							variant={errorDetail ? "default" : "outline"}
							size="sm"
							onClick={() =>
								setErrorDetail((v) =>
									v ? null : "Preview engine error: pipeline stalled",
								)
							}
						>
							Simulate error detail
						</Button>
						<Button
							type="button"
							variant={endPending ? "default" : "outline"}
							size="sm"
							onClick={() => setEndPending((v) => !v)}
						>
							Ending spinner
						</Button>
					</>
				)}
			</div>
			<VoiceRoom
				session={session}
				connection={{ key: "preview", connect: false }}
				source={{
					kind: "fixture",
					turns: PREVIEW_TURNS,
					correctionText: FIXTURE_CORRECTION_TEXT,
					transcript: source,
					scene: {
						title: "Ordering coffee",
						body: "Order a drink and a snack, ask the price, then pay.",
					},
					durationSec: 360,
				}}
			/>
		</div>
	);
}
