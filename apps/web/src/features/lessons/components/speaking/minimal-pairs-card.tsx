import type { RespeakResponse } from "@engflex/contracts";
import { Mic, Volume2 } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { useRecorder } from "#/features/lessons/hooks/use-recorder";
import { useRespeak } from "#/features/lessons/queries";
import { m } from "#/paraglide/messages";
import { PronunciationPanel } from "./pronunciation-panel";

/**
 * Static contrast set, deliberately mirrored from the Go list
 * (internal/voice/pronunciation/minimal_pairs.go). It is content, not wire
 * data: no endpoint, no persistence, so the card works before any attempt.
 */
interface MinimalPair {
	a: string;
	b: string;
	ipaA: string;
	ipaB: string;
	contrastId: string;
}

const MINIMAL_PAIRS: MinimalPair[] = [
	{ a: "think", b: "tink", ipaA: "/θɪŋk/", ipaB: "/tɪŋk/", contrastId: "th–t" },
	{ a: "thick", b: "tick", ipaA: "/θɪk/", ipaB: "/tɪk/", contrastId: "th–t" },
	{ a: "three", b: "tree", ipaA: "/θriː/", ipaB: "/triː/", contrastId: "th–t" },
	{ a: "they", b: "day", ipaA: "/ðeɪ/", ipaB: "/deɪ/", contrastId: "th–d" },
	{ a: "then", b: "den", ipaA: "/ðen/", ipaB: "/den/", contrastId: "th–d" },
	{
		a: "breathe",
		b: "breed",
		ipaA: "/briːð/",
		ipaB: "/briːd/",
		contrastId: "th–d",
	},
	{ a: "she", b: "see", ipaA: "/ʃiː/", ipaB: "/siː/", contrastId: "sh–s" },
	{ a: "ship", b: "sip", ipaA: "/ʃɪp/", ipaB: "/sɪp/", contrastId: "sh–s" },
	{ a: "shoe", b: "Sue", ipaA: "/ʃuː/", ipaB: "/suː/", contrastId: "sh–s" },
	{ a: "chew", b: "true", ipaA: "/tʃuː/", ipaB: "/truː/", contrastId: "ch–tr" },
	{
		a: "chain",
		b: "train",
		ipaA: "/tʃeɪn/",
		ipaB: "/treɪn/",
		contrastId: "ch–tr",
	},
	{ a: "light", b: "night", ipaA: "/laɪt/", ipaB: "/naɪt/", contrastId: "l–n" },
	{ a: "lap", b: "nap", ipaA: "/læp/", ipaB: "/næp/", contrastId: "l–n" },
	{ a: "low", b: "no", ipaA: "/ləʊ/", ipaB: "/nəʊ/", contrastId: "l–n" },
	{ a: "sing", b: "sin", ipaA: "/sɪŋ/", ipaB: "/sɪn/", contrastId: "ng–n" },
	{ a: "thing", b: "thin", ipaA: "/θɪŋ/", ipaB: "/θɪn/", contrastId: "ng–n" },
	{ a: "vest", b: "west", ipaA: "/vest/", ipaB: "/west/", contrastId: "v–w" },
	{ a: "vine", b: "wine", ipaA: "/vaɪn/", ipaB: "/waɪn/", contrastId: "v–w" },
	{ a: "zip", b: "sip", ipaA: "/zɪp/", ipaB: "/sɪp/", contrastId: "z–s" },
	{ a: "big", b: "pig", ipaA: "/bɪɡ/", ipaB: "/pɪɡ/", contrastId: "b–p" },
	{ a: "buy", b: "pie", ipaA: "/baɪ/", ipaB: "/paɪ/", contrastId: "b–p" },
	{ a: "fan", b: "pan", ipaA: "/fæn/", ipaB: "/pæn/", contrastId: "f–p" },
	{
		a: "bat",
		b: "back",
		ipaA: "/bæt/",
		ipaB: "/bæk/",
		contrastId: "final t–k",
	},
	{ a: "cat", b: "cap", ipaA: "/kæt/", ipaB: "/kæp/", contrastId: "final t–p" },
	{
		a: "walked",
		b: "walk",
		ipaA: "/wɔːkt/",
		ipaB: "/wɔːk/",
		contrastId: "final -ed",
	},
	{
		a: "missed",
		b: "miss",
		ipaA: "/mɪst/",
		ipaB: "/mɪs/",
		contrastId: "final -ed",
	},
];

const CONTRASTS = [...new Set(MINIMAL_PAIRS.map((pair) => pair.contrastId))];

/** Browser speech keeps the card self-contained; no model audio is stored. */
function speak(text: string) {
	if (typeof window === "undefined" || !("speechSynthesis" in window)) return;
	window.speechSynthesis.cancel();
	const utterance = new SpeechSynthesisUtterance(text);
	utterance.lang = "en-US";
	utterance.rate = 0.85;
	window.speechSynthesis.speak(utterance);
}

function SideButton({
	label,
	ipa,
	onPlay,
	side,
}: {
	label: string;
	ipa: string;
	onPlay: () => void;
	side: "A" | "B";
}) {
	return (
		<button
			type="button"
			onClick={onPlay}
			className="surface-card flex flex-1 items-center gap-3 p-3 text-left transition hover:border-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
		>
			<span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-secondary text-sm font-bold">
				{side}
			</span>
			<span className="min-w-0 flex-1">
				<span className="block text-base font-bold text-foreground">
					{label}
				</span>
				<span className="block font-mono text-xs text-muted-foreground">
					{ipa}
				</span>
			</span>
			<Volume2 aria-hidden="true" className="size-4 text-muted-foreground" />
			<span className="sr-only">{m["lessons.minimalPairs.listen"]()}</span>
		</button>
	);
}

/**
 * Minimal-pairs practice: pick a contrast, hear a hidden side, discriminate,
 * then repeat one word for a targeted re-speak score. Nothing persists.
 */
export function MinimalPairsCard() {
	const [contrast, setContrast] = useState(CONTRASTS[0] ?? "");
	const pairs = MINIMAL_PAIRS.filter((pair) => pair.contrastId === contrast);
	const [pairIndex, setPairIndex] = useState(0);
	const pair = pairs[pairIndex] ?? pairs[0] ?? MINIMAL_PAIRS[0];

	const [quiz, setQuiz] = useState<"a" | "b" | null>(null);
	const [choice, setChoice] = useState<"a" | "b" | null>(null);
	const [target, setTarget] = useState<"a" | "b">("a");
	const [result, setResult] = useState<RespeakResponse | null>(null);
	const [selfAudioUrl, setSelfAudioUrl] = useState<string | null>(null);
	const { state, start, stop, reset } = useRecorder();
	const respeak = useRespeak();
	const submittedRef = useRef(false);

	const recording = state.status === "recording";
	const readyBlob = state.status === "ready" ? state.blob : null;
	const targetWord = target === "a" ? pair.a : pair.b;
	const quizCorrect = choice !== null && quiz !== null ? choice === quiz : null;

	useEffect(() => {
		return () => {
			if (typeof window !== "undefined" && "speechSynthesis" in window) {
				window.speechSynthesis.cancel();
			}
		};
	}, []);

	useEffect(() => {
		if (state.status !== "ready" || submittedRef.current) return;
		submittedRef.current = true;
		respeak.mutate(
			{ expectedText: targetWord, audio: state.blob, mime: state.mime },
			{
				onSuccess: (data) => setResult(data),
				onError: () => {
					submittedRef.current = false;
					reset();
				},
			},
		);
	}, [state, targetWord, respeak, reset]);

	useEffect(() => {
		if (readyBlob === null || result === null) return;
		const url = URL.createObjectURL(readyBlob);
		setSelfAudioUrl(url);
		return () => {
			URL.revokeObjectURL(url);
			setSelfAudioUrl(null);
		};
	}, [readyBlob, result]);

	function resetRound() {
		submittedRef.current = false;
		respeak.reset();
		setResult(null);
		setQuiz(null);
		setChoice(null);
		reset();
	}

	function changeContrast(next: string) {
		setContrast(next);
		setPairIndex(0);
		resetRound();
	}

	function changePair(next: string) {
		setPairIndex(Number(next));
		resetRound();
	}

	function playRound() {
		const side = Math.random() < 0.5 ? "a" : "b";
		setQuiz(side);
		setChoice(null);
		speak(side === "a" ? pair.a : pair.b);
	}

	function repeat(side: "a" | "b") {
		setTarget(side);
		setResult(null);
		submittedRef.current = false;
		respeak.reset();
		reset();
		speak(side === "a" ? pair.a : pair.b);
	}

	return (
		<section className="surface-card flex flex-col gap-4 p-5">
			<div>
				<h2 className="text-lg font-bold text-foreground">
					{m["lessons.minimalPairs.title"]()}
				</h2>
				<p className="text-sm text-muted-foreground">
					{m["lessons.minimalPairs.instruction"]()}
				</p>
			</div>

			<div className="grid gap-3 sm:grid-cols-2">
				<div className="flex flex-col gap-1.5">
					<span className="text-xs font-bold text-muted-foreground">
						{m["lessons.minimalPairs.contrast"]()}
					</span>
					<Select value={contrast} onValueChange={changeContrast}>
						<SelectTrigger className="w-full">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							{CONTRASTS.map((id) => (
								<SelectItem key={id} value={id}>
									{id}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
				</div>
				<div className="flex flex-col gap-1.5">
					<span className="text-xs font-bold text-muted-foreground">
						{m["lessons.minimalPairs.pair"]()}
					</span>
					<Select value={String(pairIndex)} onValueChange={changePair}>
						<SelectTrigger className="w-full">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							{pairs.map((item, index) => (
								<SelectItem key={`${item.a}-${item.b}`} value={String(index)}>
									{item.a} / {item.b}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
				</div>
			</div>

			<div className="flex flex-col gap-2 sm:flex-row">
				<SideButton
					side="A"
					label={pair.a}
					ipa={pair.ipaA}
					onPlay={() => speak(pair.a)}
				/>
				<SideButton
					side="B"
					label={pair.b}
					ipa={pair.ipaB}
					onPlay={() => speak(pair.b)}
				/>
			</div>

			<div className="flex flex-col gap-2 rounded-xl bg-secondary/60 p-3">
				<div className="flex flex-wrap items-center gap-2">
					<Button type="button" size="sm" onClick={playRound}>
						<Volume2 data-icon="inline-start" />
						{m["lessons.minimalPairs.play"]()}
					</Button>
					<span className="text-sm font-medium text-muted-foreground">
						{quiz
							? m["lessons.minimalPairs.quizPrompt"]()
							: m["lessons.minimalPairs.playPrompt"]()}
					</span>
					{quizCorrect !== null ? (
						<span
							className={
								quizCorrect
									? "ml-auto text-sm font-bold text-accuracy"
									: "ml-auto text-sm font-bold text-ai-coral"
							}
						>
							{quizCorrect
								? m["lessons.minimalPairs.correct"]()
								: m["lessons.minimalPairs.incorrect"]()}
						</span>
					) : null}
				</div>
				<div className="flex gap-2">
					<Button
						type="button"
						variant={choice === "a" ? "default" : "outline"}
						size="sm"
						disabled={quiz === null}
						onClick={() => setChoice("a")}
					>
						{pair.a}
					</Button>
					<Button
						type="button"
						variant={choice === "b" ? "default" : "outline"}
						size="sm"
						disabled={quiz === null}
						onClick={() => setChoice("b")}
					>
						{pair.b}
					</Button>
				</div>
			</div>

			<div className="flex flex-col gap-2">
				<span className="text-xs font-bold text-muted-foreground">
					{m["lessons.minimalPairs.repeatPrompt"]()}
				</span>
				<div className="flex flex-wrap items-center gap-2">
					<Button
						type="button"
						variant={target === "a" ? "default" : "outline"}
						size="sm"
						onClick={() => repeat("a")}
						disabled={recording || respeak.isPending}
					>
						{m["lessons.minimalPairs.say"]({ word: pair.a })}
					</Button>
					<Button
						type="button"
						variant={target === "b" ? "default" : "outline"}
						size="sm"
						onClick={() => repeat("b")}
						disabled={recording || respeak.isPending}
					>
						{m["lessons.minimalPairs.say"]({ word: pair.b })}
					</Button>
					{!recording && result === null ? (
						<SubmitButton
							type="button"
							size="sm"
							className="btn btn-primary"
							disabled={state.status === "ready"}
							pending={respeak.isPending}
							onClick={() => void start()}
						>
							<Mic data-icon="inline-start" />
							{m["lessons.minimalPairs.record"]()}
						</SubmitButton>
					) : null}
					{recording ? (
						<Button
							type="button"
							size="sm"
							disabled={respeak.isPending}
							onClick={stop}
						>
							{m["lessons.minimalPairs.stop"]()}
						</Button>
					) : null}
					{result !== null ? (
						<Button
							type="button"
							variant="outline"
							size="sm"
							disabled={respeak.isPending}
							onClick={resetRound}
						>
							{m["lessons.minimalPairs.retry"]()}
						</Button>
					) : null}
				</div>
				{state.status === "error" ? (
					<p role="alert" className="text-sm font-medium text-ai-coral">
						{state.message}
					</p>
				) : null}
				{respeak.isError ? (
					<p role="alert" className="text-sm font-medium text-ai-coral">
						{m["lessons.minimalPairs.error"]()}
					</p>
				) : null}
			</div>

			{result ? (
				<PronunciationPanel
					result={result}
					referenceText={targetWord}
					selfAudioUrl={selfAudioUrl}
				/>
			) : (
				<div className="surface-card flex min-h-28 flex-col items-center justify-center gap-2 p-4 text-center">
					<Mic aria-hidden="true" className="size-5 text-muted-foreground" />
					<p className="text-sm font-medium text-muted-foreground">
						{m["lessons.minimalPairs.scoreEmpty"]()}
					</p>
				</div>
			)}
		</section>
	);
}
