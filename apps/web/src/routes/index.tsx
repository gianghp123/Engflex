import { createFileRoute } from "@tanstack/react-router";
import { MarketingShell } from "#/features/marketing/marketing-shell";
import { LandingPageMarkup } from "#/features/marketing/pages/landing-page";
import { m } from "#/paraglide/messages";

/**
 * The public front door. Lives at the route root — deliberately OUTSIDE `_app`,
 * which redirects signed-out visitors to sign-in. No AppShell, no Clerk session:
 * a visitor who has never signed up must be able to read this page.
 *
 * `MarketingShell` supplies the marketing chrome and the scoping the mockup's
 * stylesheet needs; the page supplies its own sections and closing band.
 */
export const Route = createFileRoute("/")({
	head: () => ({
		meta: [
			{ title: m["landing.meta.title"]() },
			{ name: "description", content: m["landing.meta.description"]() },
		],
	}),
	component: LandingPage,
});

function LandingPage() {
	return (
		<MarketingShell>
			<LandingPageMarkup />
		</MarketingShell>
	);
}
