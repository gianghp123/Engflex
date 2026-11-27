import type {
	AccessAction,
	PricingCatalog,
	PricingEntitlements,
	SubscriptionPlan,
	UsagePeriod,
} from "@engflex/contracts";
import { APP_ROUTES } from "#/app/app-route";
import { MoMascot } from "#/components/common/mo-mascot";
import { MarketingFooter } from "#/features/marketing/marketing-footer";
import { m } from "#/paraglide/messages";

function formatVnd(amount: number, currency: string): string {
	return new Intl.NumberFormat("vi-VN", { style: "currency", currency }).format(
		amount,
	);
}
function perMonth(amount: number, months: number): string {
	return formatVnd(Math.round(amount / months), "VND");
}

/** UsagePeriod is a wire enum — render the localized prose, never the raw code. */
function periodLabel(period: UsagePeriod): string {
	switch (period) {
		case "day":
			return m["pricing.period.day"]();
		case "month":
			return m["pricing.period.month"]();
		case "total":
			return m["pricing.period.total"]();
	}
}

function planName(plan: SubscriptionPlan): string {
	switch (plan) {
		case "1m":
			return m["pricing.plans.name.1m"]();
		case "3m":
			return m["pricing.plans.name.3m"]();
		case "12m":
			return m["pricing.plans.name.12m"]();
	}
}

/** The catalog only prices the three paid actions; anything else renders as
 *  "other" rather than leaking a wire code. Labels live with the usage page
 *  (payments.usage.actions.*) so both surfaces stay in step. */
function costActionLabel(action: AccessAction): string {
	switch (action) {
		case "voice_session_start":
			return m["payments.usage.actions.voice_session_start"]();
		case "exam_ai_review":
			return m["payments.usage.actions.exam_ai_review"]();
		case "study_plan":
			return m["payments.usage.actions.study_plan"]();
		case "writing_score":
			return m["payments.usage.actions.writing_score"]();
		default:
			return m["payments.usage.actions.other"]();
	}
}

/** The two live entitlement lines every plan card shows: AI credits and voice
 *  sessions, each with its own localized period. */
function EntitlementLines({
	entitlements,
}: {
	entitlements: PricingEntitlements;
}) {
	return (
		<>
			<li>
				{m["pricing.plans.entitlements.credits"]({
					credits: entitlements.credits,
					period: periodLabel(entitlements.creditsPeriod),
				})}
			</li>
			<li>
				{m["pricing.plans.entitlements.voice"]({
					sessions: entitlements.voiceSessions,
					period: periodLabel(entitlements.voicePeriod),
				})}
			</li>
		</>
	);
}

/**
 * The pricing page, rendered from the live catalog (GET /pricing): the Free
 * card reads catalog.free, every paid card reads one catalog plan, and the
 * trial callout and cost list read catalog.trial / catalog.costs. No price or
 * count on this page is written by hand — amounts, entitlements and costs all
 * come from the loader.
 *
 * The split follows the landing page's own two halves — "you can work through a
 * lesson, or just talk" — so Free is the lesson path and the paid cards are the
 * voice half. Both CTAs lead to sign-up.
 *
 * Every product claim is one the landing already makes. Do not add a price, a
 * count or a promise here without the evidence to back it.
 */
export function PricingPageMarkup({ catalog }: { catalog: PricingCatalog }) {
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
						<div className="l-plan" key="free">
							<h2 className="l-plan-name">{m["pricing.plans.free.name"]()}</h2>
							<div className="l-plan-for">{m["pricing.plans.free.for"]()}</div>
							<div className="l-plan-price">
								<b>{formatVnd(0, "VND")}</b>
							</div>
							<ul className="l-plan-list">
								<EntitlementLines entitlements={catalog.free} />
							</ul>
							<a className="l-btn l-ghost" href={APP_ROUTES.AUTH.SIGN_UP}>
								{m["pricing.plans.free.cta"]()}
							</a>
						</div>
						{catalog.plans.map((plan, index) => (
							<div
								className={index === 0 ? "l-plan l-featured" : "l-plan"}
								key={plan.plan}
							>
								<h2 className="l-plan-name">{planName(plan.plan)}</h2>
								<div className="l-plan-for">
									{m["pricing.plans.paid.for"]()}
								</div>
								<div className="l-plan-price">
									<b>{formatVnd(plan.amount, plan.currency)}</b>
									<span>
										{m["pricing.plans.perMonth"]({
											amount: perMonth(plan.amount, plan.months),
										})}
									</span>
								</div>
								<ul className="l-plan-list">
									{/* Plan credits/voice are monthly limits by contract
									    (Get composes each plan with PlanMonthlyLimits) —
									    only the counts ride the wire, so "month" is
									    the period here, not a guess. */}
									<EntitlementLines
										entitlements={{
											credits: plan.credits,
											voiceSessions: plan.voiceSessions,
											creditsPeriod: "month",
											voicePeriod: "month",
										}}
									/>
								</ul>
								<a
									className={index === 0 ? "l-btn" : "l-btn l-ghost"}
									href={APP_ROUTES.AUTH.SIGN_UP}
								>
									{m["pricing.plans.paid.cta"]()}
								</a>
							</div>
						))}
					</div>
				</div>
			</div>

			{/* TRIAL */}
			<div className="l-band l-white">
				<div className="l-wrap">
					<h2>{m["pricing.trial.title"]()}</h2>
					<p
						className="l-lede"
						style={{ margin: "10px 0 0", maxWidth: "64ch" }}
					>
						{m["pricing.trial.body"]({
							days: catalog.trial.days,
							credits: catalog.trial.credits,
							sessions: catalog.trial.voiceSessions,
						})}
					</p>
				</div>
			</div>

			{/* WHAT EACH AI ACTION COSTS */}
			<div className="l-band">
				<div className="l-wrap">
					<h2>{m["pricing.costs.title"]()}</h2>
					<ul className="l-plan-list" style={{ marginTop: "18px" }}>
						{catalog.costs.map((cost) => (
							<li key={cost.action}>
								{costActionLabel(cost.action)}:{" "}
								{cost.credits === 0 ? m["pricing.costs.free"]() : cost.credits}
							</li>
						))}
					</ul>
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
