import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { PageLayout } from "#/components/common/page-layout";
import { UsagePage } from "#/features/payments/components/usage/usage-page";
import { parsePage } from "#/lib/router/search-params";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/payments/usage")({
	validateSearch: (search: Record<string, unknown>): { page?: number } => ({
		page: parsePage(search.page),
	}),
	component: UsageRoutePage,
});

function UsageRoutePage() {
	useBreadcrumbs([{ label: m["payments.usage.crumb"]() }]);
	const { page = 1 } = Route.useSearch();
	const navigate = Route.useNavigate();
	return (
		<PageLayout
			action={
				<BackButton
					to={APP_ROUTES.PAYMENTS.PRICING}
					label={m["payments.usage.back"]()}
				/>
			}
		>
			<UsagePage
				page={page}
				onPageChange={(next, opts) =>
					void navigate({ search: { page: next }, replace: opts?.replace })
				}
			/>
		</PageLayout>
	);
}
