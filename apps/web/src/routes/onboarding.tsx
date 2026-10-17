import { RedirectToSignIn, Show } from "@clerk/tanstack-react-start";
import type { Goal, InterestTopic, Level } from "@engflex/contracts";
import {
	GOAL_VALUES,
	INTEREST_TOPIC_VALUES,
	LEVEL_VALUES,
} from "@engflex/contracts";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute, redirect, useNavigate } from "@tanstack/react-router";
import {
	AudioWaveform,
	BadgeCheck,
	Brain,
	Briefcase,
	Coffee,
	GraduationCap,
	type LucideIcon,
	Plane,
	Smile,
} from "lucide-react";
import { useEffect, useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { SubmitButton } from "#/components/common/submit-button";
import { LocaleSwitcher } from "#/components/layout/locale-switcher";
import { ThemeToggle } from "#/components/ThemeToggle";
import { Button } from "#/components/ui/button";
import { FieldError } from "#/components/ui/field";
import { RadioGroup } from "#/components/ui/radio-group";
import { Separator } from "#/components/ui/separator";
import { ToggleGroup } from "#/components/ui/toggle-group";
import { FlowHeader } from "#/features/onboarding/components/flow-header";
import { GoalChip } from "#/features/onboarding/components/goal-chip";
import { LevelRadioCard } from "#/features/onboarding/components/level-radio-card";
import { OnboardingSummaryCard } from "#/features/onboarding/components/onboarding-summary-card";
import {
	type Slide,
	SlideTrack,
} from "#/features/onboarding/components/slide-track";
import { TimeCommitmentCard } from "#/features/onboarding/components/time-commitment-card";
import { TopicChip } from "#/features/onboarding/components/topic-chip";
import {
	profileKeys,
	profileQueryOptions,
	useUpdateProfile,
} from "#/features/profiles/queries";
import { ApiError } from "#/lib/api/api";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/onboarding")({
	beforeLoad: async ({ context: { queryClient }, preload }) => {
		// See _app/route.tsx: a hover must not read anything.
		if (preload) return;

		let onboardingCompleted = false;
		try {
			const profile = await queryClient.query(profileQueryOptions());
			onboardingCompleted = profile.onboardingCompletedAt != null;
		} catch (err) {
			if (err instanceof ApiError && err.status === 401) {
				throw redirect({ href: APP_ROUTES.AUTH.SIGN_IN });
			}
			if (!(err instanceof ApiError && err.status === 404)) {
				throw err;
			}
		}
		if (onboardingCompleted) {
			throw redirect({ to: APP_ROUTES.HOME });
		}
	},
	component: OnboardingPage,
});

/** Local onboarding draft: level/commitment stay unset until the user picks. */
type Draft = {
	level: Level | null;
	goals: Goal[];
	dailyCommitmentMin: number;
	topics: InterestTopic[];
};

/** Key for the in-progress answers, so a locale switch (which reloads the page)
 *  does not lose them. Cleared once onboarding succeeds. */
const ONBOARDING_PROGRESS_KEY = "onboarding:progress";

const EMPTY_DRAFT: Draft = {
	level: null,
	goals: [],
	dailyCommitmentMin: 0,
	topics: [],
};

const LEVEL_COPY: Record<
	Level,
	{
		range: string;
		title: () => string;
		description: () => string;
		hint: () => string;
		hintIcon: LucideIcon;
	}
> = {
	beginner: {
		range: "A1–A2",
		title: () => m["onboarding.level.option.beginner.title"](),
		description: () => m["onboarding.level.option.beginner.description"](),
		hint: () => m["onboarding.level.option.beginner.hint"](),
		hintIcon: AudioWaveform,
	},
	intermediate: {
		range: "B1–B2",
		title: () => m["onboarding.level.option.intermediate.title"](),
		description: () => m["onboarding.level.option.intermediate.description"](),
		hint: () => m["onboarding.level.option.intermediate.hint"](),
		hintIcon: BadgeCheck,
	},
	advanced: {
		range: "C1–C2",
		title: () => m["onboarding.level.option.advanced.title"](),
		description: () => m["onboarding.level.option.advanced.description"](),
		hint: () => m["onboarding.level.option.advanced.hint"](),
		hintIcon: Brain,
	},
};
const LEVELS = LEVEL_VALUES.map((value) => ({ value, ...LEVEL_COPY[value] }));

const GOAL_COPY: Record<Goal, { label: () => string; icon: LucideIcon }> = {
	travel: { label: () => m["onboarding.goal.option.travel"](), icon: Plane },
	work: { label: () => m["onboarding.goal.option.work"](), icon: Briefcase },
	exams: {
		label: () => m["onboarding.goal.option.exams"](),
		icon: GraduationCap,
	},
	daily: { label: () => m["onboarding.goal.option.daily"](), icon: Coffee },
	fun: { label: () => m["onboarding.goal.option.fun"](), icon: Smile },
};
const GOALS = GOAL_VALUES.map((value) => ({ value, ...GOAL_COPY[value] }));

const TOPIC_COPY: Record<InterestTopic, () => string> = {
	travel: () => m["onboarding.topics.option.travel"](),
	business: () => m["onboarding.topics.option.business"](),
	movies_tv: () => m["onboarding.topics.option.moviesTv"](),
	technology: () => m["onboarding.topics.option.technology"](),
	food: () => m["onboarding.topics.option.food"](),
	music: () => m["onboarding.topics.option.music"](),
	sports: () => m["onboarding.topics.option.sports"](),
	news: () => m["onboarding.topics.option.news"](),
};
const TOPICS = INTEREST_TOPIC_VALUES.map((value) => ({
	value,
	label: TOPIC_COPY[value],
}));

const COMMITMENTS = [
	{
		minutes: 5,
		label: () => m["onboarding.time.option.c5.label"](),
		sub: () => m["onboarding.time.option.c5.sub"](),
	},
	{
		minutes: 10,
		label: () => m["onboarding.time.option.c10.label"](),
		sub: () => m["onboarding.time.option.c10.sub"](),
	},
	{
		minutes: 15,
		label: () => m["onboarding.time.option.c15.label"](),
		sub: () => m["onboarding.time.option.c15.sub"](),
		recommended: true,
	},
	{
		minutes: 30,
		label: () => m["onboarding.time.option.c30.label"](),
		sub: () => m["onboarding.time.option.c30.sub"](),
	},
];

function OnboardingPage() {
	const navigate = useNavigate();
	const queryClient = useQueryClient();
	const saveProfile = useUpdateProfile();
	const [slide, setSlide] = useState(0);
	const [draft, setDraft] = useState<Draft>(EMPTY_DRAFT);
	const [restored, setRestored] = useState(false);
	const [levelError, setLevelError] = useState<string | null>(null);
	const [goalError, setGoalError] = useState<string | null>(null);
	const [topicError, setTopicError] = useState<string | null>(null);
	const [timeError, setTimeError] = useState<string | null>(null);

	// Restore in-progress answers on mount (client-only), then persist on change.
	// The `restored` gate keeps the first persist from overwriting stored answers
	// with the empty draft before the restore render lands.
	useEffect(() => {
		try {
			const raw = window.localStorage.getItem(ONBOARDING_PROGRESS_KEY);
			if (raw) setDraft(JSON.parse(raw) as Draft);
		} catch {
			window.localStorage.removeItem(ONBOARDING_PROGRESS_KEY);
		}
		setRestored(true);
	}, []);
	useEffect(() => {
		if (!restored) return;
		window.localStorage.setItem(ONBOARDING_PROGRESS_KEY, JSON.stringify(draft));
	}, [draft, restored]);

	function next() {
		if (slide === 0 && !draft.level) {
			setLevelError(m["onboarding.level.error"]());
			return;
		}
		if (slide === 1 && draft.goals.length < 1) {
			setGoalError(m["onboarding.goal.error"]());
			return;
		}
		if (slide === 1 && draft.topics.length < 2) {
			setTopicError(m["onboarding.topics.error"]());
			return;
		}
		if (slide === 2 && draft.dailyCommitmentMin === 0) {
			setTimeError(m["onboarding.time.error"]());
			return;
		}
		setLevelError(null);
		setGoalError(null);
		setTopicError(null);
		setTimeError(null);
		setSlide((current) => Math.min(3, current + 1));
	}

	function back() {
		setSlide((current) => Math.max(0, current - 1));
	}

	function finish() {
		if (!draft.level || draft.dailyCommitmentMin === 0) return;
		saveProfile.mutate(
			{
				level: draft.level,
				goals: draft.goals,
				dailyCommitmentMin: draft.dailyCommitmentMin,
				topics: draft.topics,
			},
			{
				onSuccess: (profile) => {
					window.localStorage.removeItem(ONBOARDING_PROGRESS_KEY);
					// MUST precede the navigation: `_app`'s guard reads this
					// entry for the onboarding flag, so writing it after
					// navigating would bounce the user back here.
					queryClient.setQueryData(profileKeys.me, profile);
					navigate({ to: APP_ROUTES.HOME });
				},
			},
		);
	}

	const selectedLevel = LEVELS.find((level) => level.value === draft.level);
	const selectedGoals = GOALS.filter((goal) =>
		draft.goals.includes(goal.value),
	);
	const selectedTopics = TOPICS.filter((topic) =>
		draft.topics.includes(topic.value),
	);

	const slides: Slide[] = [
		{
			key: "level",
			node: (
				<section className="flex flex-col gap-4 p-5">
					<div>
						<h2 className="text-lg font-extrabold tracking-tight">
							{m["onboarding.level.title"]()}
						</h2>
						<p className="mt-1 text-xs text-muted-foreground">
							{m["onboarding.level.subtitle"]()}
						</p>
					</div>
					<RadioGroup
						value={draft.level ?? ""}
						onValueChange={(value) => {
							setDraft((current) => ({ ...current, level: value as Level }));
							setLevelError(null);
						}}
						className="flex flex-col gap-3"
					>
						{LEVELS.map((level) => (
							<LevelRadioCard
								key={level.value}
								value={level.value}
								range={level.range}
								title={level.title()}
								description={level.description()}
								hint={level.hint()}
								hintIcon={level.hintIcon}
								selected={draft.level === level.value}
							/>
						))}
					</RadioGroup>
					<div className="min-h-5">
						{levelError ? <FieldError>{levelError}</FieldError> : null}
					</div>
				</section>
			),
		},
		{
			key: "goal",
			node: (
				<section className="flex flex-col gap-5 p-5">
					<div>
						<h2 className="text-lg font-extrabold tracking-tight">
							{m["onboarding.goal.title"]()}
						</h2>
						<p className="mt-1 text-xs text-muted-foreground">
							{m["onboarding.goal.subtitle"]()}
						</p>
					</div>
					<ToggleGroup
						type="multiple"
						value={draft.goals}
						onValueChange={(value) => {
							setDraft((current) => ({
								...current,
								goals: value as Goal[],
							}));
							if (value.length >= 1) setGoalError(null);
						}}
						className="flex flex-wrap gap-2"
					>
						{GOALS.map((goal) => (
							<GoalChip
								key={goal.value}
								value={goal.value}
								label={goal.label()}
								icon={goal.icon}
							/>
						))}
					</ToggleGroup>
					<div className="min-h-5">
						{goalError ? <FieldError>{goalError}</FieldError> : null}
					</div>
					<div className="flex flex-col gap-2">
						<h3 className="text-sm font-semibold">
							{m["onboarding.topics.title"]()}
						</h3>
						<p className="text-xs text-muted-foreground">
							{m["onboarding.topics.subtitle"]()}
						</p>
						<ToggleGroup
							type="multiple"
							value={draft.topics}
							onValueChange={(value) => {
								setDraft((current) => ({
									...current,
									topics: value as InterestTopic[],
								}));
								if (value.length >= 2) setTopicError(null);
							}}
							className="flex flex-wrap justify-start gap-2"
						>
							{TOPICS.map((topic) => (
								<TopicChip
									key={topic.value}
									value={topic.value}
									label={topic.label()}
								/>
							))}
						</ToggleGroup>
						<div className="min-h-5">
							{topicError ? <FieldError>{topicError}</FieldError> : null}
						</div>
					</div>
				</section>
			),
		},
		{
			key: "time",
			node: (
				<section className="flex flex-col gap-4 p-5">
					<div>
						<h2 className="text-lg font-extrabold tracking-tight">
							{m["onboarding.time.title"]()}
						</h2>
						<p className="mt-1 text-xs text-muted-foreground">
							{m["onboarding.time.subtitle"]()}
						</p>
					</div>
					<div className="grid gap-3 sm:grid-cols-2">
						{COMMITMENTS.map((commitment) => (
							<TimeCommitmentCard
								key={commitment.minutes}
								minutes={commitment.minutes}
								label={commitment.label()}
								sub={commitment.sub()}
								recommended={commitment.recommended}
								selected={draft.dailyCommitmentMin === commitment.minutes}
								onSelect={(minutes) => {
									setDraft((current) => ({
										...current,
										dailyCommitmentMin: minutes,
									}));
									setTimeError(null);
								}}
							/>
						))}
					</div>
					<div className="min-h-5">
						{timeError ? <FieldError>{timeError}</FieldError> : null}
					</div>
				</section>
			),
		},
		{
			key: "result",
			node: (
				<OnboardingSummaryCard
					level={{
						title: selectedLevel?.title() ?? "",
						range: selectedLevel?.range ?? "",
					}}
					goals={selectedGoals.map((goal) => goal.label())}
					topics={selectedTopics.map((topic) => topic.label())}
					commitment={m["onboarding.time.minutesUnit"]({
						count: draft.dailyCommitmentMin,
					})}
				/>
			),
		},
	];

	return (
		<>
			<Show when="signed-in">
				<div className="app-canvas flex min-h-svh items-center justify-center p-4 sm:p-6">
					<div className="surface-card w-full max-w-2xl">
						<div className="flex items-center justify-between gap-3 px-5">
							<FlowHeader />
							<div className="flex items-center gap-3">
								<LocaleSwitcher />
								<ThemeToggle />
							</div>
						</div>
						<Separator />
						<SlideTrack index={slide} slides={slides} />
						<div className="flex items-center justify-between border-t p-4">
							<Button
								type="button"
								variant="ghost"
								onClick={back}
								disabled={slide === 0}
							>
								{m["common.actions.back"]()}
							</Button>
							<SubmitButton
								type="button"
								pending={saveProfile.isPending}
								onClick={slide === 3 ? finish : next}
							>
								{slide === 3
									? m["common.actions.finish"]()
									: m["common.actions.continue"]()}
							</SubmitButton>
						</div>
						{saveProfile.isError ? (
							<div className="px-4 pb-4">
								<FieldError>{m["onboarding.submitError"]()}</FieldError>
							</div>
						) : null}
					</div>
				</div>
			</Show>
			<Show when="signed-out">
				<RedirectToSignIn />
			</Show>
		</>
	);
}
