import { createFileRoute } from "@tanstack/react-router";
import { MarketingShell } from "#/features/marketing/marketing-shell";
import { PricingPageMarkup } from "#/features/marketing/pages/pricing-page";
import { m } from "#/paraglide/messages";

/**
 * The public pricing page, outside `_app` like the landing and About pages.
 *
 * What it shows is mock data: no billing exists yet, so read the note at the top
 * of `PricingPageMarkup` before treating any figure here as real.
 */
export const Route = createFileRoute("/pricing")({
	head: () => ({
		meta: [
			{ title: m["pricing.meta.title"]() },
			{ name: "description", content: m["pricing.meta.description"]() },
		],
	}),
	component: PricingPage,
});

function PricingPage() {
	return (
		<MarketingShell>
			<PricingPageMarkup />
		</MarketingShell>
	);
}
