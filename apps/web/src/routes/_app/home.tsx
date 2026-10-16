import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { ProgressRing, StatCard } from "#/components/common/stat-card";
import { Button } from "#/components/ui/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "#/components/ui/empty";
import { Skeleton } from "#/components/ui/skeleton";
import { useProgress } from "#/features/attempts/queries";
import { ActiveLessonCard } from "#/features/dashboard/components/active-lesson-card";
import { LiveVoicePartnerCard } from "#/features/dashboard/components/live-voice-partner-card";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
import { useUser } from "@clerk/tanstack-react-start";
import { createFileRoute, Link } from "@tanstack/react-router";
import { BookOpenText, Flame } from "lucide-react";

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
  // Computed here, not inside the card: the ring is the caller's widget now.
  const voiceGoalPct =
    progress.data && progress.data.dailyGoalMinutes > 0
      ? (progress.data.voiceMinutesToday / progress.data.dailyGoalMinutes) * 100
      : 0;
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
            title={m["dashboard.stats.streak"]()}
            value={progress.data?.streakDays ?? 0}
            aside={
              <span className="tile bg-secondary text-primary">
                <Flame className="size-5" />
              </span>
            }
            pending={progress.isPending}
          />
          <StatCard
            title={m["dashboard.stats.voice"]()}
            value={progress.data?.voiceMinutesToday ?? 0}
            aside={<ProgressRing value={voiceGoalPct} />}
            // The ring is 64px where the default placeholder is a 36px
            // tile, so this card states its own loading shape.
            skeleton={
              <>
                <div className="flex min-w-0 flex-1 flex-col gap-2 self-stretch">
                  <Skeleton className="h-3 w-20" />
                  <Skeleton className="h-7 w-16" />
                  <Skeleton className="h-3 w-24" />
                </div>
                <Skeleton className="size-16 rounded-full" />
              </>
            }
            pending={progress.isPending}
          >
            {progress.data ? (
              <span className="text-xs text-muted-foreground">
                {m["dashboard.stats.voiceGoal"]({
                  done: progress.data.voiceMinutesToday,
                  goal: progress.data.dailyGoalMinutes,
                })}
              </span>
            ) : null}
          </StatCard>
          <StatCard
            title={m["dashboard.stats.lessons"]()}
            value={progress.data?.lessonsCompleted ?? 0}
            aside={
              <span className="tile bg-secondary text-primary">
                <BookOpenText className="size-5" />
              </span>
            }
            pending={progress.isPending}
          />
        </div>
      ) : null}

      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-bold tracking-tight text-foreground">
          {m["dashboard.sequence.title"]()}
        </h2>
        <Empty className="surface-card items-center text-center">
          <MoMascot variant="nice" size={72} />
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
