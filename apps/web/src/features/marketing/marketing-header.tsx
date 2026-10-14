import { Link } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { LocaleSwitcher } from "#/components/layout/locale-switcher";
import { m } from "#/paraglide/messages";

/**
 * The marketing top bar: wordmark, page nav, locale, sign in and get started.
 * Rendered by `MarketingShell`, so every marketing page gets the same chrome and
 * none has to remember to include it.
 *
 * The nav lists pages, not sections — one entry per marketing page — and the one
 * you are on is marked with `aria-current`, so the highlight is announced rather
 * than only coloured. The router decides that; this component computes nothing.
 */
export function MarketingHeader() {
	return (
		<div className="l-ptop">
			<a className="l-brand" href={APP_ROUTES.LANDING}>
				<svg
					aria-hidden="true"
					width="36"
					height="35"
					viewBox="0 0 120 116"
					fill="none"
				>
					<path
						d="M30 8h60a22 22 0 0 1 22 22v36a22 22 0 0 1-22 22H52l-20 16v-16h-2A22 22 0 0 1 8 66V30A22 22 0 0 1 30 8z"
						transform="translate(0 4)"
						fill="#432F96"
					/>
					<path
						d="M30 8h60a22 22 0 0 1 22 22v36a22 22 0 0 1-22 22H52l-20 16v-16h-2A22 22 0 0 1 8 66V30A22 22 0 0 1 30 8z"
						fill="#6C4DF0"
					/>
					<g transform="translate(16 6) scale(.88)">
						<ellipse cx="50" cy="50" rx="42" ry="38" fill="#966F47" />
						<ellipse cx="50" cy="52" rx="33" ry="28" fill="#E8D7B8" />
						<ellipse
							cx="36"
							cy="48"
							rx="10"
							ry="14"
							transform="rotate(-12 36 48)"
							fill="#583A1F"
						/>
						<ellipse
							cx="64"
							cy="48"
							rx="10"
							ry="14"
							transform="rotate(12 64 48)"
							fill="#583A1F"
						/>
						<circle cx="36" cy="48" r="4.5" fill="#1D1342" />
						<circle cx="34.5" cy="46" r="1.5" fill="#fff" />
						<circle cx="64" cy="48" r="4.5" fill="#1D1342" />
						<circle cx="62.5" cy="46" r="1.5" fill="#fff" />
						<ellipse cx="50" cy="58" rx="5" ry="3.5" fill="#1D1342" />
						<path
							d="M44 65C47 68 53 68 56 65"
							fill="none"
							stroke="#1D1342"
							strokeWidth="2.5"
							strokeLinecap="round"
						/>
					</g>
				</svg>
				<span className="l-wordmark">
					eng<span style={{ color: "#FF9E3D" }}>flex</span>
				</span>
			</a>
			<nav className="l-nav">
				{NAV.map((item) => (
					<Link
						key={item.to}
						to={item.to}
						activeOptions={{ exact: item.exact }}
						activeProps={{ "aria-current": "page" }}
					>
						{item.label()}
					</Link>
				))}
			</nav>
			<div className="l-hact">
				<LocaleSwitcher />
				<a className="l-hsignin" href={APP_ROUTES.AUTH.SIGN_IN}>
					{m["landing.nav.signIn"]()}
				</a>
				<a className="l-hget" href={APP_ROUTES.AUTH.SIGN_UP}>
					{m["landing.nav.getStarted"]()}
				</a>
			</div>
		</div>
	);
}

/**
 * One entry per marketing page, in the order they appear. Adding a page means
 * adding a line here — the highlight comes from the router, so nothing else
 * needs to know which entries exist.
 */
const NAV = [
	// `exact` matters on the home link only: every path begins with it, so
	// without this the router would call Home current on every page.
	{ to: APP_ROUTES.LANDING, label: m["landing.nav.home"], exact: true },
	{ to: APP_ROUTES.PRICING, label: m["landing.nav.pricing"], exact: false },
	{ to: APP_ROUTES.ABOUT, label: m["landing.nav.about"], exact: false },
] as const;
