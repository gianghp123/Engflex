import { createFileRoute } from "@tanstack/react-router";
import { MarketingShell } from "#/features/marketing/marketing-shell";
import { AboutPageMarkup } from "#/features/marketing/pages/about-page";
import { m } from "#/paraglide/messages";

/**
 * The public About page. Deliberately outside `_app`, like the landing page: it
 * is marketing copy for people who have not signed up, so it must render without
 * a Clerk session and without the app's sidebar. It keeps the path it had inside
 * `_app`, so the sidebar's link to it keeps resolving.
 */
export const Route = createFileRoute("/about")({
	head: () => ({
		meta: [
			{ title: m["about.meta.title"]() },
			{ name: "description", content: m["about.meta.description"]() },
		],
	}),
	component: AboutPage,
});

function AboutPage() {
	return (
		<MarketingShell>
			<AboutPageMarkup />
		</MarketingShell>
	);
}
