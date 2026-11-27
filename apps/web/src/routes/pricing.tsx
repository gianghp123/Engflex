import { createFileRoute } from "@tanstack/react-router";
import { MarketingShell } from "#/features/marketing/marketing-shell";
import { PricingPageMarkup } from "#/features/marketing/pages/pricing-page";
import { getPricingCatalog } from "#/features/payments/service";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/pricing")({
	loader: () => getPricingCatalog(),
	head: () => ({
		meta: [
			{ title: m["pricing.meta.title"]() },
			{ name: "description", content: m["pricing.meta.description"]() },
		],
	}),
	component: PricingPage,
	errorComponent: PricingUnavailable,
});

function PricingPage() {
	const catalog = Route.useLoaderData();
	return (
		<MarketingShell>
			<PricingPageMarkup catalog={catalog} />
		</MarketingShell>
	);
}

function PricingUnavailable() {
	return (
		<MarketingShell>
			<div className="container-content py-16">
				<h1 className="text-2xl font-bold">
					{m["pricing.unavailable.title"]()}
				</h1>
				<p className="mt-2 text-muted-foreground">
					{m["pricing.unavailable.body"]()}
				</p>
			</div>
		</MarketingShell>
	);
}
