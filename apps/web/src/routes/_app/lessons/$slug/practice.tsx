import { useAuth } from "@clerk/tanstack-react-start";
import type {
	Activity,
	ActivityType,
	AttemptProgress,
	CheckedQuestion,
	CheckResult,
	LessonDetail,
	WritingScoreResponse,
} from "@engflex/contracts";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute, notFound, useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { LanguagePicker } from "#/components/common/language-picker";
import { PageSplit } from "#/components/common/page-layout";
import { SubmitButton } from "#/components/common/submit-button";
import { CheckQuestionPanel } from "#/features/lessons/components/check-question-panel";
import { ListeningBriefCard } from "#/features/lessons/components/listening/listening-brief-card";
import { PartShell } from "#/features/lessons/components/part-shell";
import { PassagePanel } from "#/features/lessons/components/reading/passage-panel";
import { SpeakingActivity } from "#/features/lessons/components/speaking/read-aloud-card";
import { ModelAnswerCard } from "#/features/lessons/components/writing/model-answer-card";
import { ScoreCard } from "#/features/lessons/components/writing/score-card";
import { WritingEditorCard } from "#/features/lessons/components/writing/writing-editor-card";
import { WritingPromptBanner } from "#/features/lessons/components/writing/writing-prompt-banner";
import { PART_META } from "#/features/lessons/parts";
import {
	lessonDetailBySlugQueryOptions,
	openAttemptQueryOptions,
	useCheckAnswer,
	useLessonDetailBySlug,
	useScoreWriting,
	useStartLessonAttempt,
} from "#/features/lessons/queries";
import {
	completePart,
	getFirstIncompleteActivity,
} from "#/features/lessons/store";
import { DockGuide } from "#/features/vocabulary/components/dock-guide";
import { TextActionDock } from "#/features/vocabulary/components/text-action-dock";
import { useExplanationLanguage } from "#/hooks/use-explanation-language";
import { loadOr404 } from "#/lib/route-loader";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/lessons/$slug/practice")({
	loader: async ({ context: { queryClient }, params: { slug } }) => {
		const detail = await loadOr404(
			queryClient,
			lessonDetailBySlugQueryOptions(slug),
		);
		if (detail.activities.length === 0) throw notFound();
	},
	component: PracticePage,
});

function PracticePage() {
	const { slug } = Route.useParams();
	const detailQuery = useLessonDetailBySlug(slug);

	if (detailQuery.isPending) return null;
	// A failed fetch is an error, not a missing lesson (see loader above:
	// genuine 404s and empty lessons never reach the component).
	if (detailQuery.isError) throw detailQuery.error;
	const detail = detailQuery.data;
	if (!detail) return null;

	return <PracticeInner key={detail.id} slug={slug} detail={detail} />;
}

/** Rehydrated per-question state for one checkable activity. */
interface QuestionSeed {
	answers: Record<number, string>;
	results: Record<number, CheckResult>;
}

function toQuestionSeed(list: CheckedQuestion[]): QuestionSeed {
	const answers: Record<number, string> = {};
	const results: Record<number, CheckResult> = {};
	for (const item of list) {
		answers[item.index] = item.key;
		results[item.index] = {
			correct: item.correct,
			correctKey: item.correctKey,
			explanation: item.explanation,
		};
	}
	return { answers, results };
}

function buildQuestionSeeds(
	progress: AttemptProgress,
): Record<string, QuestionSeed> {
	const seeds: Record<string, QuestionSeed> = {};
	for (const [activityId, list] of Object.entries(progress.checks)) {
		seeds[activityId] = toQuestionSeed(list);
	}
	return seeds;
}

/** Every part with a recorded server entry counts complete — the server's
 *  own finalize rule (any attached part covers it), so writing and speaking
 *  rehydrate too. Mid-activity answers still restore via `checks`. */
function markCompleteParts(
	lessonId: string,
	activities: Activity[],
	progress: AttemptProgress,
) {
	const completed = new Set(progress.completedParts ?? []);
	for (const activity of activities) {
		if (completed.has(activity.id)) {
			completePart(lessonId, activity.partNumber, activities.length);
		}
	}
}

function PracticeInner({
	slug,
	detail,
}: {
	slug: string;
	detail: LessonDetail;
}) {
	// lessonId below is the unit id: the store stays id-keyed, the slug is URL-only.
	const lessonId = detail.id;
	useBreadcrumbs([
		{ label: m["nav.item.lessons"](), to: APP_ROUTES.LESSONS.LIST },
		{ label: detail.title },
	]);
	const fallback = detail.activities[0];
	if (!fallback) throw new Error("practice: lesson has no activities");
	const start =
		getFirstIncompleteActivity(lessonId, detail.activities) ?? fallback;
	const [activePart, setActivePart] = useState<ActivityType>(start.type);

	// Always-on refresh guard: browsers show their own generic prompt text,
	// so no message key exists for it. Cleanup detaches it on unmount so the
	// prompt never leaks onto other pages.
	useEffect(() => {
		const handler = (event: BeforeUnloadEvent) => {
			event.preventDefault();
		};
		window.addEventListener("beforeunload", handler);
		return () => window.removeEventListener("beforeunload", handler);
	}, []);

	const { isLoaded } = useAuth();
	// Read-only mount: the run is opened by the entry click (LessonHeader /
	// ActivityRow) or the fallback card below. Remounts are pure reads by
	// construction — no once-guards, no writes.
	// `enabled: isLoaded` prevents doomed fail-fast fetches: inside Clerk's
	// loading window getToken() yields null, so the request is a guaranteed
	// 401 and retry can only burn time on it — including seconds of SSR burn
	// on hard loads. In-flight token dips are covered by the
	// retry-except-404 in the query options instead.

	const { asking, confirm, dismiss, reopen } =
		useExplanationLanguage("writing");
	const openQuery = useQuery({
		...openAttemptQueryOptions(lessonId),
		enabled: isLoaded,
	});
	const [startedId, setStartedId] = useState<string | null>(null);
	const progress = openQuery.data ?? null;
	const attemptId = startedId ?? progress?.attempt.id ?? null;
	const seeds = progress ? buildQuestionSeeds(progress) : {};
	useEffect(() => {
		if (progress) markCompleteParts(lessonId, detail.activities, progress);
	}, [progress, lessonId, detail]);

	const activity = detail.activities.find((item) => item.type === activePart);
	if (!activity)
		throw new Error(`practice: no activity for part ${activePart}`);
	const currentPart: ActivityType = activity.type;

	if (attemptId === null && !openQuery.isPending) {
		return (
			<PartShell
				lessonId={lessonId}
				part={currentPart}
				onSelect={setActivePart}
				headerExtra={<DockGuide />}
			>
				<NoOpenRunCard
					lessonId={lessonId}
					onStarted={(id) => {
						setStartedId(id);
						void openQuery.refetch();
					}}
				/>
			</PartShell>
		);
	}

	return (
		<TextActionDock source={{ type: "lesson", id: lessonId }}>
			<PartShell
				lessonId={lessonId}
				part={currentPart}
				onSelect={setActivePart}
				headerExtra={
					<div className="flex gap-2">
						<LanguagePicker
							open={asking}
							onOpenChange={(next) => {
								if (next) reopen();
								else dismiss();
							}}
							onConfirm={confirm}
						/>
						<DockGuide />
					</div>
				}
			>
				{currentPart === "reading" &&
				activity?.reading &&
				attemptId !== null ? (
					<ReadingActivity
						key={`${activity.id}:${attemptId}`}
						lessonId={lessonId}
						slug={slug}
						activity={activity}
						activities={detail.activities}
						onAdvance={setActivePart}
						attemptId={attemptId}
						seed={seeds[activity.id] ?? { answers: {}, results: {} }}
					/>
				) : null}
				{currentPart === "listening" &&
				activity?.listening &&
				attemptId !== null ? (
					<ListeningActivity
						key={`${activity.id}:${attemptId}`}
						lessonId={lessonId}
						slug={slug}
						activity={activity}
						activities={detail.activities}
						onAdvance={setActivePart}
						attemptId={attemptId}
						seed={seeds[activity.id] ?? { answers: {}, results: {} }}
					/>
				) : null}
				{currentPart === "writing" && activity?.writing ? (
					<WritingActivity
						lessonId={lessonId}
						slug={slug}
						activity={activity}
						activities={detail.activities}
						onAdvance={setActivePart}
						attemptId={attemptId}
					/>
				) : null}
				{currentPart === "speaking" && activity?.speaking ? (
					<SpeakingBranch
						lessonId={lessonId}
						slug={slug}
						activity={activity}
						activities={detail.activities}
						onAdvance={setActivePart}
						attemptId={attemptId}
					/>
				) : null}
			</PartShell>
		</TextActionDock>
	);
}

/** Shared per-question check state: answers + locked results by index. The
 *  activity mounts only once the run (and therefore the seed) exists, so
 *  plain `useState` initializers do the whole job — no sync effect. */
function useCheckedQuestions(
	activity: Activity,
	attemptId: string | null,
	seed: QuestionSeed,
) {
	const [index, setIndex] = useState(() => firstUnanswered(seed, activity));
	const [answers, setAnswers] = useState(seed.answers);
	const [results, setResults] = useState(seed.results);
	const check = useCheckAnswer();

	function firstUnanswered(seed: QuestionSeed, activity: Activity): number {
		const total =
			activity.reading?.questions.length ??
			activity.listening?.questions.length ??
			0;
		let target = 0;
		while (target < total && seed.results[target] !== undefined) target++;
		return Math.max(0, Math.min(target, total - 1));
	}

	function handleCheck() {
		if (!attemptId) return;
		const key = answers[index];
		if (key === undefined || results[index] !== undefined || check.isPending)
			return;
		check.mutate(
			{ activityId: activity.id, attemptId, questionIndex: index, key },
			{
				onSuccess: (result) => {
					setResults((current) => ({ ...current, [index]: result }));
				},
				onError: () => {
					toast.error(m["lessons.check.failed"]());
				},
			},
		);
	}

	return {
		index,
		setIndex,
		answers,
		setAnswers,
		results,
		pending: check.isPending,
		handleCheck,
	};
}

function useContinue(
	lessonId: string,
	slug: string,
	activity: Activity,
	activities: Activity[],
	onAdvance: (part: ActivityType) => void,
) {
	const navigate = useNavigate();
	const index = activities.findIndex((item) => item.id === activity.id);
	const next = index >= 0 ? activities[index + 1] : undefined;
	const continueLabel = next
		? m["lessons.step.continueTo"]({ part: PART_META[next.type].label() })
		: m["lessons.step.finishLesson"]();

	function handleContinue() {
		completePart(lessonId, activity.partNumber, activities.length);
		if (next) {
			onAdvance(next.type);
		} else {
			navigate({ to: APP_ROUTES.LESSONS.DETAIL, params: { slug } });
		}
	}

	return { continueLabel, handleContinue };
}

function NoOpenRunCard({
	lessonId,
	onStarted,
}: {
	lessonId: string;
	onStarted: (id: string) => void;
}) {
	const start = useStartLessonAttempt();
	return (
		<div className="surface-card flex flex-col items-start gap-3 p-5">
			<h2 className="text-lg font-bold text-foreground">
				{m["lessons.practice.noRunTitle"]()}
			</h2>
			<p className="text-[15px] font-medium text-muted-foreground">
				{m["lessons.practice.noRunBody"]()}
			</p>
			<SubmitButton
				type="button"
				className="btn btn-primary"
				pending={start.isPending}
				onClick={() =>
					start.mutate(lessonId, {
						onSuccess: (attempt) => onStarted(attempt.id),
						onError: () => {
							toast.error(m["lessons.attempt.failed"]());
						},
					})
				}
			>
				{m["lessons.practice.noRunAction"]()}
			</SubmitButton>
		</div>
	);
}

function ReadingActivity({
	lessonId,
	slug,
	activity,
	activities,
	onAdvance,
	attemptId,
	seed,
}: {
	lessonId: string;
	slug: string;
	activity: Activity;
	activities: Activity[];
	onAdvance: (part: ActivityType) => void;
	attemptId: string | null;
	seed: QuestionSeed;
}) {
	const payload = activity.reading;
	const checked = useCheckedQuestions(activity, attemptId, seed);
	const { handleContinue } = useContinue(
		lessonId,
		slug,
		activity,
		activities,
		onAdvance,
	);

	if (!payload) throw new Error("practice: activity payload missing");
	const question = payload.questions[checked.index];
	if (!question)
		throw new Error(`practice: question index ${checked.index} out of range`);
	const isLast = checked.index === payload.questions.length - 1;

	return (
		<PageSplit
			main={
				<PassagePanel
					kicker={PART_META.reading.label()}
					title={activity.title}
					passage={payload.text}
				/>
			}
			aside={
				<CheckQuestionPanel
					question={question}
					questionNumber={checked.index + 1}
					totalQuestions={payload.questions.length}
					selectedKey={checked.answers[checked.index]}
					result={checked.results[checked.index]}
					onSelect={(key) =>
						checked.setAnswers((current) => ({
							...current,
							[checked.index]: key,
						}))
					}
					onCheck={() => checked.handleCheck()}
					onPrev={() => checked.setIndex((current) => Math.max(0, current - 1))}
					onNext={() =>
						checked.setIndex((current) =>
							Math.min(payload.questions.length - 1, current + 1),
						)
					}
					isLast={isLast}
					onContinue={handleContinue}
					pending={checked.pending || attemptId === null}
				/>
			}
		/>
	);
}

function ListeningActivity({
	lessonId,
	slug,
	activity,
	activities,
	onAdvance,
	attemptId,
	seed,
}: {
	lessonId: string;
	slug: string;
	activity: Activity;
	activities: Activity[];
	onAdvance: (part: ActivityType) => void;
	attemptId: string | null;
	seed: QuestionSeed;
}) {
	const payload = activity.listening;
	const checked = useCheckedQuestions(activity, attemptId, seed);
	const { handleContinue } = useContinue(
		lessonId,
		slug,
		activity,
		activities,
		onAdvance,
	);

	if (!payload) throw new Error("practice: activity payload missing");
	const question = payload.questions[checked.index];
	if (!question)
		throw new Error(`practice: question index ${checked.index} out of range`);
	const isLast = checked.index === payload.questions.length - 1;

	return (
		<PageSplit
			main={
				<ListeningBriefCard
					title={activity.title}
					description={activity.description}
					audioUrl={payload.audioUrl}
					transcript={payload.transcript}
					revealed={checked.results[checked.index] !== undefined}
				/>
			}
			aside={
				<CheckQuestionPanel
					question={question}
					questionNumber={checked.index + 1}
					totalQuestions={payload.questions.length}
					selectedKey={checked.answers[checked.index]}
					result={checked.results[checked.index]}
					onSelect={(key) =>
						checked.setAnswers((current) => ({
							...current,
							[checked.index]: key,
						}))
					}
					onCheck={() => checked.handleCheck()}
					onPrev={() => checked.setIndex((current) => Math.max(0, current - 1))}
					onNext={() =>
						checked.setIndex((current) =>
							Math.min(payload.questions.length - 1, current + 1),
						)
					}
					isLast={isLast}
					onContinue={handleContinue}
					pending={checked.pending || attemptId === null}
				/>
			}
		/>
	);
}

function WritingActivity({
	lessonId,
	slug,
	activity,
	activities,
	onAdvance,
	attemptId,
}: {
	lessonId: string;
	slug: string;
	activity: Activity;
	activities: Activity[];
	onAdvance: (part: ActivityType) => void;
	attemptId: string | null;
}) {
	const payload = activity.writing;
	const [text, setText] = useState("");
	const score = useScoreWriting();
	const { withLanguage } = useExplanationLanguage("writing");
	const [result, setResult] = useState<WritingScoreResponse | null>(null);
	// The first Check writing on this browser picks the explanation language
	// first; later submits run straight through.
	function handleSubmit() {
		if (!attemptId || score.isPending) return;
		withLanguage(() =>
			score.mutate(
				{ activityId: activity.id, attemptId, text },
				{
					onSuccess: (value) => setResult(value),
					onError: () => {
						toast.error(m["lessons.writing.score.failed"]());
					},
				},
			),
		);
	}
	const { continueLabel, handleContinue } = useContinue(
		lessonId,
		slug,
		activity,
		activities,
		onAdvance,
	);
	const partCount = activities.length;

	if (!payload) throw new Error("practice: writing payload missing");

	return (
		<PageSplit
			main={
				<div className="flex flex-col gap-4">
					<WritingPromptBanner
						partNumber={activity.partNumber}
						partCount={partCount}
						task={payload.task}
						instructions={payload.instructions}
						stimulus={payload.stimulus}
						minWords={payload.minWords}
						maxWords={payload.maxWords}
					/>
					{result ? (
						<ModelAnswerCard
							modelAnswer={payload.modelAnswer}
							checklist={payload.checklist}
						/>
					) : null}
				</div>
			}
			aside={
				<div className="flex flex-col gap-4">
					<WritingEditorCard
						value={text}
						onChange={(next) => {
							setText(next);
							if (result !== null) setResult(null);
						}}
						maxWords={payload.maxWords}
						onClear={() => {
							setText("");
							setResult(null);
						}}
						onSubmit={handleSubmit}
						onRevise={() => setResult(null)}
						submitted={result !== null || score.isPending}
						pending={score.isPending}
					/>
					{result ? (
						<ScoreCard
							score={result}
							continueLabel={continueLabel}
							onContinue={handleContinue}
						/>
					) : null}
				</div>
			}
		/>
	);
}

function SpeakingBranch({
	lessonId,
	slug,
	activity,
	activities,
	onAdvance,
	attemptId,
}: {
	lessonId: string;
	slug: string;
	activity: Activity;
	activities: Activity[];
	onAdvance: (part: ActivityType) => void;
	attemptId: string | null;
}) {
	const { handleContinue } = useContinue(
		lessonId,
		slug,
		activity,
		activities,
		onAdvance,
	);

	function handleComplete() {
		handleContinue();
	}

	return (
		<SpeakingActivity
			activity={activity}
			attemptId={attemptId}
			onComplete={handleComplete}
		/>
	);
}
