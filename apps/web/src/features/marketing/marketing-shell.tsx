import type { ReactNode } from "react";
import { MarketingHeader } from "#/features/marketing/marketing-header";
import "#/features/marketing/marketing.css";

/**
 * The marketing shell: everything a page needs to render the site chrome, in one
 * element.
 *
 * `.landing` scopes the stylesheet, and `.l-page` is not decoration — it is the
 * element that carries the palette variables (`--purple`, `--inkpad`, `--line`…)
 * that every rule in marketing.css reads. A page rendered without it silently
 * loses every colour on the page.
 *
 * The scope class is still `landing`: it is the name the mockup's stylesheet was
 * scoped with when this was one page, and marketing.css is keyed off it in 1181
 * lines of selectors (plus the `l-` class prefix). Renaming the folder and the
 * components was worth it; renaming the scope would be a mechanical rewrite of
 * every rule and every marketing markup file for no behavioural gain.
 *
 * A page renders this once and lays out its own sections and footer inside:
 *
 *     <MarketingShell>
 *       …sections…
 *       <MarketingFooter>…closing call to action…</MarketingFooter>
 *     </MarketingShell>
 */
export function MarketingShell({ children }: { children: ReactNode }) {
	return (
		<div className="landing">
			<div className="l-page">
				<MarketingHeader />
				{children}
			</div>
		</div>
	);
}
