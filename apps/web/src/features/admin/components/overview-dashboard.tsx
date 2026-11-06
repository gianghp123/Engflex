import { useAuth } from "@clerk/tanstack-react-start";
import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { StatCard } from "#/components/common/stat-card";
import { SubmitButton } from "#/components/common/submit-button";
import { listLessonHub } from "#/features/admin/lessons/service";
import { listAdminSubscriptions } from "#/features/admin/premium/service";
import { listUsers } from "#/features/admin/users/service";
import { listVideoExercises } from "#/features/admin/videos/service";
import { ApiError } from "#/lib/api";
import { m } from "#/paraglide/messages";

export function AdminOverview() {
	const { isLoaded } = useAuth();
	const options = {
		enabled: isLoaded,
		retry: (count: number, error: unknown) =>
			!(error instanceof ApiError && [401, 403, 404].includes(error.status)) &&
			count < 3,
	};
	const users = useQuery({
		queryKey: ["admin", "users", 1, 1, ""],
		queryFn: () => listUsers(1, 1, ""),
		...options,
	});
	const lessons = useQuery({
		queryKey: ["admin", "lessons"],
		queryFn: listLessonHub,
		...options,
	});
	const videos = useQuery({
		queryKey: ["admin", "video-exercises", 1, 1],
		queryFn: () => listVideoExercises(1, 1),
		...options,
	});
	const subscriptions = useQuery({
		queryKey: ["admin", "premium-subscriptions", 1, 1, "", ""],
		queryFn: () => listAdminSubscriptions(1, 1, "", ""),
		...options,
	});
	const metrics = [
		{
			id: "users",
			title: m["admin.nav.users"](),
			value: users.data?.pagination.total,
			query: users,
			to: APP_ROUTES.ADMIN.USERS,
		},
		{
			id: "lessons",
			title: m["admin.nav.lessons"](),
			value: lessons.data?.reduce(
				(sum, section) => sum + section.units.length,
				0,
			),
			query: lessons,
			to: APP_ROUTES.ADMIN.LESSONS,
		},
		{
			id: "videos",
			title: m["admin.nav.videos"](),
			value: videos.data?.pagination.total,
			query: videos,
			to: APP_ROUTES.ADMIN.VIDEOS,
		},
		{
			id: "subscriptions",
			title: m["admin.overview.metrics.subscriptions"](),
			value: subscriptions.data?.pagination.total,
			query: subscriptions,
			to: APP_ROUTES.ADMIN.PREMIUM,
		},
	];
	const actions = [
		{ to: APP_ROUTES.ADMIN.USERS, label: m["admin.nav.users"]() },
		{ to: APP_ROUTES.ADMIN.LESSONS, label: m["admin.nav.lessons"]() },
		{ to: APP_ROUTES.ADMIN.VIDEOS, label: m["admin.nav.videos"]() },
		{ to: APP_ROUTES.ADMIN.VOCABULARY, label: m["admin.nav.vocabulary"]() },
		{ to: APP_ROUTES.ADMIN.SCENARIOS, label: m["admin.nav.scenarios"]() },
		{ to: APP_ROUTES.ADMIN.PREMIUM, label: m["admin.nav.premium"]() },
		{ to: APP_ROUTES.ADMIN.EXAMS, label: m["admin.nav.exams"]() },
	];
	useBreadcrumbs([
		{ label: m["admin.nav.overview"](), to: APP_ROUTES.ADMIN.OVERVIEW },
	]);
	return (
		<div className="container-content flex flex-col gap-2 py-8">
			<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
				{m["admin.overview.title"]()}
			</h1>
			<p className="text-muted-foreground">
				{m["admin.overview.description"]()}
			</p>
			<div className="mt-6 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
				{metrics.map((metric) => (
					<StatCard
						key={metric.id}
						title={metric.title}
						value={metric.query.isError ? "—" : (metric.value ?? "—")}
						pending={metric.query.isPending}
					>
						{metric.query.isError ? (
							<div role="alert">
								<p className="text-sm text-destructive">
									{m["admin.toast.error"]()}
								</p>
								<SubmitButton
									pending={metric.query.isFetching}
									onClick={() => void metric.query.refetch()}
								>
									{m["admin.overview.metrics.retry"]()}
								</SubmitButton>
							</div>
						) : (
							<Link className="text-sm text-primary" to={metric.to}>
								{m["admin.overview.metrics.manage"]()}
							</Link>
						)}
					</StatCard>
				))}
			</div>
			<h2 className="mt-6 text-lg font-bold">
				{m["admin.overview.quickActions.title"]()}
			</h2>
			<div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
				{actions.map((action) => (
					<Link
						key={action.to}
						to={action.to}
						className="rounded-xl border bg-card p-4 font-medium hover:bg-muted"
					>
						{action.label}
					</Link>
				))}
			</div>
		</div>
	);
}
