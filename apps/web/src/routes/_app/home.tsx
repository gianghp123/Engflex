import { useUser } from "@clerk/tanstack-react-start";
import { createFileRoute, Link } from "@tanstack/react-router";
import { BookOpenText, Flame } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { StatCard } from "#/components/common/stat-card";
import { Button } from "#/components/ui/button";
import {
	Empty,
	EmptyContent,
	EmptyDescription,
	EmptyHeader,
	EmptyTitle,
} from "#/components/ui/empty";
import { useProgress } from "#/features/attempts/queries";
import { ActiveLessonCard } from "#/features/dashboard/components/active-lesson-card";
import { LiveVoicePartnerCard } from "#/features/dashboard/components/live-voice-partner-card";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";

export const Route = createFileRoute("/_app/home")({
	component: DashboardPage,
});

function DashboardPage() {
	useBreadcrumbs([
		{ label: m["nav.item.home"](), to: APP_ROUTES.HOME },
		{ label: m["nav.group.practice"]() },
	]);
	const { user } = useUser();
	const firstName = user?.firstName;
	const progress = useProgress();
	const today = new Intl.DateTimeFormat(getLocale(), {
		weekday: "long",
		day: "numeric",
		month: "long",
	}).format(new Date());

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="wave" size={84} />,
				title: firstName
					? m["dashboard.hero.greetingName"]({ name: firstName })
					: m["dashboard.hero.greeting"](),
				description: today,
			}}
		>
			<div className="grid gap-4 lg:grid-cols-12">
				<div className="lg:col-span-7">
					<ActiveLessonCard />
				</div>
				<div className="lg:col-span-5">
					<LiveVoicePartnerCard />
				</div>
			</div>

			{progress.isPending || progress.isSuccess ? (
				<div className="grid gap-4 sm:grid-cols-3">
					<StatCard
						label={m["dashboard.stats.streak"]()}
						value={progress.data?.streakDays ?? 0}
						icon={Flame}
						pending={progress.isPending}
					/>
					<StatCard
						label={m["dashboard.stats.voice"]()}
						value={progress.data?.voiceMinutesToday ?? 0}
						footer={
							progress.data
								? m["dashboard.stats.voiceGoal"]({
										done: progress.data.voiceMinutesToday,
										goal: progress.data.dailyGoalMinutes,
									})
								: undefined
						}
						ringValue={
							progress.data && progress.data.dailyGoalMinutes > 0
								? (progress.data.voiceMinutesToday /
										progress.data.dailyGoalMinutes) *
									100
								: 0
						}
						pending={progress.isPending}
					/>
					<StatCard
						label={m["dashboard.stats.lessons"]()}
						value={progress.data?.lessonsCompleted ?? 0}
						icon={BookOpenText}
						pending={progress.isPending}
					/>
				</div>
			) : null}

			<section className="flex flex-col gap-3">
				<h2 className="text-lg font-bold tracking-tight text-foreground">
					{m["dashboard.sequence.title"]()}
				</h2>
				<Empty className="surface-card items-center text-center">
					<MoMascot variant="wave" size={72} />
					<EmptyHeader>
						<EmptyTitle>{m["dashboard.zero.sequenceTitle"]()}</EmptyTitle>
						<EmptyDescription>
							{m["dashboard.zero.sequenceBody"]()}
						</EmptyDescription>
					</EmptyHeader>
					<EmptyContent>
						<Button asChild className="btn btn-primary">
							<Link to={APP_ROUTES.LESSONS.LIST}>
								{m["dashboard.zero.sequenceCta"]()}
							</Link>
						</Button>
					</EmptyContent>
				</Empty>
			</section>
		</PageLayout>
	);
}
