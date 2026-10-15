import type { Goal, InterestTopic, Level } from "@engflex/contracts";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { useProfile } from "#/features/profiles/queries";
import { m } from "#/paraglide/messages";

// Value → label maps for the stored enums. Explicit (never a template key) so
// a Paraglide rename breaks the build rather than silently resolving undefined.
const LEVEL_LABELS: Record<Level, () => string> = {
	beginner: () => m["onboarding.level.option.beginner.title"](),
	intermediate: () => m["onboarding.level.option.intermediate.title"](),
	advanced: () => m["onboarding.level.option.advanced.title"](),
};

const GOAL_LABELS: Record<Goal, () => string> = {
	travel: () => m["onboarding.goal.option.travel"](),
	work: () => m["onboarding.goal.option.work"](),
	exams: () => m["onboarding.goal.option.exams"](),
	daily: () => m["onboarding.goal.option.daily"](),
	fun: () => m["onboarding.goal.option.fun"](),
};

const TOPIC_LABELS: Record<InterestTopic, () => string> = {
	travel: () => m["onboarding.topics.option.travel"](),
	business: () => m["onboarding.topics.option.business"](),
	movies_tv: () => m["onboarding.topics.option.moviesTv"](),
	technology: () => m["onboarding.topics.option.technology"](),
	food: () => m["onboarding.topics.option.food"](),
	music: () => m["onboarding.topics.option.music"](),
	sports: () => m["onboarding.topics.option.sports"](),
	news: () => m["onboarding.topics.option.news"](),
};

function Row({ label, value }: { label: string; value: string }) {
	return (
		<div className="rounded-xl border bg-card px-4 py-3">
			<dt className="text-[11px] font-semibold tracking-wide text-muted-foreground uppercase">
				{label}
			</dt>
			<dd className="mt-0.5 text-sm font-semibold text-foreground">{value}</dd>
		</div>
	);
}

/**
 * Our own profile modal, opened from the Clerk `UserButton` menu item — it
 * shows the saved `user_profiles` row (GET /profiles/me) rather than Clerk's
 * account UI. Read-only for now.
 */
export function ProfileDialog({
	open,
	onOpenChange,
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}) {
	const { data } = useProfile();
	const preferences = data?.preferences;

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>{m["profile.title"]()}</DialogTitle>
					<DialogDescription>{m["profile.subtitle"]()}</DialogDescription>
				</DialogHeader>
				{preferences ? (
					<dl className="grid gap-3">
						<Row
							label={m["profile.level"]()}
							value={LEVEL_LABELS[preferences.level]()}
						/>
						<Row
							label={m["profile.commitment"]()}
							value={m["onboarding.time.minutesUnit"]({
								count: preferences.dailyCommitmentMin,
							})}
						/>
						<Row
							label={m["profile.goals"]()}
							value={(preferences.goals ?? [])
								.map((goal) => GOAL_LABELS[goal]())
								.join(", ")}
						/>
						<Row
							label={m["profile.topics"]()}
							value={(preferences.topics ?? [])
								.map((topic) => TOPIC_LABELS[topic]())
								.join(", ")}
						/>
					</dl>
				) : (
					<p className="text-sm text-muted-foreground">
						{m["profile.loading"]()}
					</p>
				)}
			</DialogContent>
		</Dialog>
	);
}
