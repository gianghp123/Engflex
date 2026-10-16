import type {
	MispronouncedWord,
	PronunciationResult,
	ReferencePhone,
} from "@engflex/contracts";
import { cn } from "cn";
import { Pause, Play } from "lucide-react";
import { Fragment, type ReactNode, useRef, useState } from "react";
import { MoMascot, type MoVariant } from "#/components/common/mo-mascot";
import { ProgressBar } from "#/components/common/progress-bar";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";

/** Go nil slices arrive as `null` — every list below null-guards before use. */
function errorEntries(result: PronunciationResult): MispronouncedWord[] {
	return result.errors ?? [];
}

/**
 * Coaching bubble from the attempt state: a positive default when nothing was
 * flagged (same `>= 70` boundary as the "Near model" band below), the rhythm
 * note when the score is low-but-clean, otherwise every flagged word, ranked
 * by confidence, highlighted in the conjuncted list. Structured content (not
 * a flat string) so the words keep a highlight; private-use sentinels never
 * reach the screen — formatToParts only maps elements back to error entries.
 */
const LIST_SENTINEL = "\uE000";
function coachingMessage(
	score: number,
	errors: MispronouncedWord[],
): ReactNode {
	const ranked = [...errors].sort((a, b) => b.confidence - a.confidence);
	if (ranked.length === 0) {
		return score >= 70
			? m["lessons.speaking.coachCorrect"]()
			: m["lessons.speaking.coachNone"]();
	}
	const parts = new Intl.ListFormat(getLocale(), {
		style: "long",
		type: "conjunction",
	}).formatToParts(
		ranked.map((_, index) => `${LIST_SENTINEL}${index}${LIST_SENTINEL}`),
	);
	return (
		<>
			{m["lessons.speaking.coachPrefix"]()}{" "}
			{parts.map((part, partIndex) => {
				if (part.type !== "element") {
					// biome-ignore lint/suspicious/noArrayIndexKey: parts never reorder within a message; the index disambiguates repeated literals (", ", " and ").
					return <Fragment key={partIndex}>{part.value}</Fragment>;
				}
				const error = ranked[Number(part.value.replaceAll(LIST_SENTINEL, ""))];
				if (!error) return null;
				return (
					// biome-ignore lint/suspicious/noArrayIndexKey: same stability argument as above.
					<Fragment key={partIndex}>
						<span className="rounded-sm bg-secondary px-1">{error.word}</span>
					</Fragment>
				);
			})}
		</>
	);
}

/**
 * Mo's face follows the coaching story: flagged words get the coach's
 * thinking pose (or the puzzled face when nothing was heard at all), a clean
 * attempt gets approval — cheer for a near-model score, a thumbs up above the
 * 70 band, a soft smile below it. No sad/tears: Mo is a coach, not a judge.
 */
function moVariant(score: number, errors: MispronouncedWord[]): MoVariant {
	if (errors.length > 0) {
		return errors.some((error) => error.heard.length > 0)
			? "thinking"
			: "confused";
	}
	if (score >= 85) return "cheer";
	return score >= 70 ? "thumbsup" : "nice";
}

function bandLabel(score: number): string {
	if (score >= 70) return m["lessons.speaking.bandNearModel"]();
	if (score >= 50) return m["lessons.speaking.bandGettingThere"]();
	if (score >= 30) return m["lessons.speaking.bandKeepPracticing"]();
	return m["lessons.speaking.bandHardToRecognize"]();
}

function clampPct(value: number): number {
	if (!Number.isFinite(value)) return 0;
	return Math.min(100, Math.max(0, Math.round(value)));
}

type ChipStatus = "good" | "error" | "missing";

interface WordChip {
	key: string;
	display: string;
	status: ChipStatus;
	/** Target IPA (`/…/`) — same source as Mo's message — or null when unknown. */
	subline: string | null;
	/** Flagged chips: the IPA the engine actually heard, or null. */
	heard: string | null;
}

/** Strip punctuation for matching only — the display token keeps it. */
function stripWord(token: string): string {
	return token.replace(/[^\p{L}\p{N}'’]+/gu, "");
}

/**
 * One chip per reference word (transcript order): joined to errors and to
 * `referencePhones` by lowercased word. Every chip carries the word's target
 * IPA (same source as Mo's message); a flagged chip additionally shows the
 * heard decode, labeled, so the learner sees what to fix at a glance. Pill
 * tint carries the status: red = error entry, gray = not heard at all,
 * blue = good.
 */
function buildChips(
	referenceText: string,
	errors: MispronouncedWord[],
	referencePhones: ReferencePhone[],
): WordChip[] {
	const byWord = new Map<string, MispronouncedWord>();
	for (const error of errors) {
		const key = error.word.toLowerCase();
		if (!byWord.has(key)) byWord.set(key, error);
	}
	const phonesByWord = new Map<string, string>();
	for (const entry of referencePhones) {
		const key = entry.word.toLowerCase();
		if (!phonesByWord.has(key)) phonesByWord.set(key, entry.phones);
	}
	return referenceText
		.split(/\s+/)
		.filter((token) => token.length > 0)
		.map((token, index) => {
			const match = stripWord(token).toLowerCase();
			// Duplicate words recur, so the index disambiguates identical keys.
			const base = {
				key: `${token}-${index}`,
				display: token,
			};
			if (match.length === 0) {
				return {
					...base,
					status: "good" as ChipStatus,
					subline: null,
					heard: null,
				};
			}
			const error = byWord.get(match) ?? null;
			const expected = phonesByWord.get(match) || error?.expected || "";
			const subline = expected.length > 0 ? `/${expected}/` : null;
			if (!error) {
				return {
					...base,
					status: "good" as ChipStatus,
					subline,
					heard: null,
				};
			}
			if (error.heard.length === 0) {
				return {
					...base,
					status: "missing" as ChipStatus,
					subline,
					heard: null,
				};
			}
			return {
				...base,
				status: "error" as ChipStatus,
				subline,
				heard: `/${error.heard}/`,
			};
		});
}

const CHIP_TONE: Record<ChipStatus, string> = {
	good: "bg-accuracy-tint",
	error: "bg-ai-coral-tint",
	missing: "bg-muted",
};

function ChipView({ chip }: { chip: WordChip }) {
	const pill = (
		<span
			className={cn(
				"rounded-lg px-2.5 py-1 text-[13px] font-bold text-foreground",
				CHIP_TONE[chip.status],
			)}
		>
			{chip.display}
		</span>
	);
	// The sub-line is the target pronunciation for every chip — same source
	// as Mo's message — so it never carries a status tint.
	const sub = chip.subline ? (
		<span className="font-mono text-[13px] font-medium text-muted-foreground">
			{chip.subline}
		</span>
	) : null;
	// Flagged chips only: what the engine heard, labeled so it can never be
	// mistaken for the target above it.
	const heard = chip.heard ? (
		<span className="max-w-36 break-all text-center font-mono text-[13px] font-medium text-ai-coral">
			{m["lessons.speaking.heardShort"]()} {chip.heard}
		</span>
	) : null;
	return (
		<span className="flex flex-col items-center justify-start gap-1">
			{pill}
			{sub}
			{heard}
		</span>
	);
}

function SelfPlayback({ url }: { url: string }) {
	const ref = useRef<HTMLAudioElement | null>(null);
	const [playing, setPlaying] = useState(false);
	return (
		<div>
			{/* biome-ignore lint/a11y/useMediaCaption: learner's own recording — no caption track exists; the reference sentence renders as transcript chips above. */}
			<audio
				ref={ref}
				src={url}
				preload="metadata"
				onEnded={() => setPlaying(false)}
			/>
			<Button
				type="button"
				size="sm"
				className="btn btn-primary"
				onClick={() => {
					const audio = ref.current;
					if (!audio) return;
					if (playing) {
						audio.pause();
						setPlaying(false);
					} else {
						audio.currentTime = 0;
						void audio.play().then(
							() => setPlaying(true),
							() => setPlaying(false),
						);
					}
				}}
			>
				{playing ? <Pause className="size-4" /> : <Play className="size-4" />}
				{m["lessons.speaking.listenSelf"]()}
			</Button>
		</div>
	);
}

/**
 * Mo coach panel: coaching bubble, score + band, recipe bars, one chip per
 * reference word, self-playback. Never shows acousticDistance, raw error
 * rates, confidence numbers, or the deferred curves/prosody/transcription.
 */
export function PronunciationPanel({
	result,
	referenceText,
	selfAudioUrl,
}: {
	result: PronunciationResult;
	referenceText: string;
	selfAudioUrl: string | null;
}) {
	const errors = errorEntries(result);
	const score = Number.isFinite(result.score) ? result.score : 0;
	const chips = buildChips(referenceText, errors, result.referencePhones ?? []);
	// Recipe bars only — the raw rates never render.
	const sounds = clampPct(100 * (1 - (result.phonemeErrorRate ?? 0)));
	const words = clampPct(100 * (1 - (result.wordErrorRate ?? 0)));

	return (
		<div className="surface-card flex flex-col gap-4 p-5">
			<div className="flex items-start gap-2.5">
				<MoMascot variant={moVariant(score, errors)} size={44} />
				<p className="bubble bubble-partner min-w-0 flex-1 text-[15px]">
					{coachingMessage(score, errors)}
				</p>
			</div>
			<div>
				<div className="flex flex-wrap items-baseline gap-x-2">
					<span className="text-sm font-bold text-muted-foreground">
						{m["lessons.speaking.scoreLabel"]()}
					</span>
					<span className="text-2xl font-bold text-foreground">
						{Math.round(score)} / 100
					</span>
					<span className="text-sm font-medium text-muted-foreground">
						· {bandLabel(score)}
					</span>
				</div>
				<div className="mt-2 flex flex-col gap-2">
					<div>
						<div className="flex items-baseline justify-between text-xs font-medium text-muted-foreground">
							<span>{m["lessons.speaking.soundsLabel"]()}</span>
							<span>{sounds}</span>
						</div>
						<ProgressBar value={sounds} tone="accuracy" />
					</div>
					<div>
						<div className="flex items-baseline justify-between text-xs font-medium text-muted-foreground">
							<span>{m["lessons.speaking.wordsLabel"]()}</span>
							<span>{words}</span>
						</div>
						<ProgressBar value={words} tone="primary" />
					</div>
				</div>
			</div>
			<div className="flex flex-wrap items-start gap-x-3 gap-y-2">
				{chips.map((chip) => (
					<ChipView key={chip.key} chip={chip} />
				))}
			</div>
			{selfAudioUrl ? <SelfPlayback url={selfAudioUrl} /> : null}
		</div>
	);
}
