import { createFileRoute, Link } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot, type MoVariant } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";

interface ReturnSearch {
	vnp_ResponseCode?: string;
}

export const Route = createFileRoute("/_app/payments/vnpay/return")({
	validateSearch: (search: Record<string, unknown>): ReturnSearch => ({
		vnp_ResponseCode:
			typeof search.vnp_ResponseCode === "string"
				? search.vnp_ResponseCode
				: undefined,
	}),
	component: ReturnPage,
});

function ReturnPage() {
	useBreadcrumbs([{ label: m["payments.return.crumb"]() }]);
	const { vnp_ResponseCode: responseCode } = Route.useSearch();

	// The response code on this page is UX-only: the browser redirect can be
	// tampered with, so it only decides what to *show*, never what access the
	// account actually has — that's settled server-side by the VNPay IPN call.
	const succeeded = responseCode === "00";
	const unknown = responseCode === undefined;

	const variant: MoVariant = succeeded ? "cheer" : unknown ? "thinking" : "sad";
	const title = succeeded
		? m["payments.return.successTitle"]()
		: unknown
			? m["payments.return.pendingTitle"]()
			: m["payments.return.failedTitle"]();
	const body = succeeded
		? m["payments.return.successBody"]()
		: unknown
			? m["payments.return.pendingBody"]()
			: m["payments.return.failedBody"]();

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant={variant} size={84} />,
				title,
				description: body,
			}}
		>
			<div className="flex gap-2">
				<Button asChild>
					<Link to={APP_ROUTES.HOME}>{m["payments.return.backHome"]()}</Link>
				</Button>
				{!succeeded && !unknown && (
					<Button asChild variant="outline">
						<Link to={APP_ROUTES.PAYMENTS.PRICING}>
							{m["payments.return.tryAgain"]()}
						</Link>
					</Button>
				)}
			</div>
		</PageLayout>
	);
}
