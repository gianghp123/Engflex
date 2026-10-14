import { APP_ROUTES } from "#/app/app-route";
import { MoMascot } from "#/components/common/mo-mascot";
import { MarketingFooter } from "#/features/marketing/marketing-footer";
import { m } from "#/paraglide/messages";

/**
 * The pricing page.
 *
 * MOCK DATA, and the only page here that is. There is no plan, price or
 * subscription model anywhere in the repo and no checkout, so the two tiers
 * below describe the intended split rather than a commitment, and both CTAs lead
 * to sign-up. Move the line between the tiers in `PLANS`; before this page
 * becomes a promise it needs real prices and the policies that go with them —
 * refunds, cancellation, payment methods, billing period. None of those are
 * written here, because they are business decisions, not mock data.
 *
 * The split follows the landing page's own two halves — "you can work through a
 * lesson, or just talk" — so Free is the lesson path and Plus is the voice half.
 *
 * Every product claim is one the landing already makes (see the allowlist in
 * `features/marketing/pages/about-page.tsx`).
 */
export function PricingPageMarkup() {
	return (
		<>
			{/* HERO */}
			<div className="l-band l-white">
				<div className="l-wrap">
					<div className="l-hero-split">
						<div>
							<h1>{m["pricing.hero.title"]()}</h1>
							<p
								className="l-lede"
								style={{ margin: "14px 0 28px", maxWidth: "46ch" }}
							>
								{m["pricing.hero.body"]()}
							</p>
							<div
								style={{
									display: "flex",
									alignItems: "center",
									gap: "18px",
									flexWrap: "wrap",
								}}
							>
								<a className="l-btn" href={APP_ROUTES.AUTH.SIGN_UP}>
									{m["pricing.hero.cta"]()}
								</a>
								<a className="l-btn l-link" href={APP_ROUTES.AUTH.SIGN_IN}>
									{m["pricing.hero.ctaSecondary"]()}
								</a>
							</div>
						</div>
						{/* Same shape as About: Mo in the right column, below the copy
						    once the hero stacks. */}
						<div className="l-hero-mo">
							<MoMascot variant="surprised" size={240} />
						</div>
					</div>
				</div>
			</div>

			{/* THE PLANS */}
			<div className="l-band">
				<div className="l-wrap">
					<div className="l-plans">
						{PLANS.map((plan) => (
							<div
								className={plan.featured ? "l-plan l-featured" : "l-plan"}
								key={plan.name()}
							>
								<h2 className="l-plan-name">{plan.name()}</h2>
								<div className="l-plan-for">{plan.pitch()}</div>
								<div className="l-plan-price">
									<b>{plan.price()}</b>
									<span>{plan.period()}</span>
								</div>
								<ul className="l-plan-list">
									{plan.features.map((feature) => (
										<li key={feature()}>{feature()}</li>
									))}
								</ul>
								<a
									className={plan.featured ? "l-btn" : "l-btn l-ghost"}
									href={APP_ROUTES.AUTH.SIGN_UP}
								>
									{plan.cta()}
								</a>
							</div>
						))}
					</div>
				</div>
			</div>

			{/* THE SAME EITHER WAY */}
			<div className="l-band l-white">
				<div className="l-wrap">
					<h2>{m["pricing.included.title"]()}</h2>
					<p
						className="l-lede"
						style={{ margin: "10px 0 30px", maxWidth: "64ch" }}
					>
						{m["pricing.included.body"]()}
					</p>
					<div className="l-tiles l-three">
						{FACTS.map((fact) => (
							<span className="l-tile" key={fact.title()}>
								<span>
									<b>{fact.title()}</b>
									<small>{fact.body()}</small>
								</span>
							</span>
						))}
					</div>
				</div>
			</div>

			{/* CLOSING + FOOTER */}
			<MarketingFooter>
				<h1>{m["pricing.closing.title"]()}</h1>
				<p className="l-lede">{m["pricing.closing.body"]()}</p>
				<div
					style={{
						display: "flex",
						alignItems: "center",
						justifyContent: "center",
						gap: "18px",
						flexWrap: "wrap",
					}}
				>
					<a className="l-btn l-light" href={APP_ROUTES.AUTH.SIGN_UP}>
						{m["pricing.closing.cta"]()}
					</a>
					<a className="l-btn l-link l-onpurple" href={APP_ROUTES.AUTH.SIGN_IN}>
						{m["pricing.closing.ctaSecondary"]()}
					</a>
				</div>
			</MarketingFooter>
		</>
	);
}

/**
 * The two tiers. `featured` is presentation only — it outlines the paid card and
 * gives it the filled CTA, nothing more.
 *
 * Thunks, so the locale resolves at render rather than at module load.
 */
const PLANS = [
	{
		featured: false,
		name: m["pricing.plans.free.name"],
		pitch: m["pricing.plans.free.for"],
		price: m["pricing.plans.free.price"],
		period: m["pricing.plans.free.period"],
		features: [
			m["pricing.plans.free.feature1"],
			m["pricing.plans.free.feature2"],
			m["pricing.plans.free.feature3"],
			m["pricing.plans.free.feature4"],
		],
		cta: m["pricing.plans.free.cta"],
	},
	{
		featured: true,
		name: m["pricing.plans.plus.name"],
		pitch: m["pricing.plans.plus.for"],
		price: m["pricing.plans.plus.price"],
		period: m["pricing.plans.plus.period"],
		features: [
			m["pricing.plans.plus.feature1"],
			m["pricing.plans.plus.feature2"],
			m["pricing.plans.plus.feature3"],
			m["pricing.plans.plus.feature4"],
		],
		cta: m["pricing.plans.plus.cta"],
	},
] as const;

/** True of both tiers, so they sit below the cards rather than inside either. */
const FACTS = [
	{
		title: m["pricing.included.item1Title"],
		body: m["pricing.included.item1Body"],
	},
	{
		title: m["pricing.included.item2Title"],
		body: m["pricing.included.item2Body"],
	},
	{
		title: m["pricing.included.item3Title"],
		body: m["pricing.included.item3Body"],
	},
] as const;
