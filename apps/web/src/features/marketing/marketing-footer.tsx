import type { ReactNode } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { MoMascot } from "#/components/common/mo-mascot";
import { m } from "#/paraglide/messages";

/**
 * The dark closing band, footer included.
 *
 * The band has to travel with the footer because the crest, the ink background
 * and the single dot field that spans them are one material: a page that wanted
 * the footer alone would otherwise have to rebuild the crest and get the dot
 * alignment wrong (see the dot-field note in marketing.css).
 *
 * `children` is the page's closing call to action, laid out by the band. It is
 * optional, so a page can end on a bare footer.
 */
export function MarketingFooter({ children }: { children?: ReactNode }) {
	return (
		<div className="l-ctawrap">
			<div className="l-stage">
				<span
					className="l-debr"
					style={{ left: "6%", bottom: "66px", transform: "rotate(-7deg)" }}
				>
					<span
						className="l-ic l-w"
						style={{ width: "46px", height: "46px", borderRadius: "15px" }}
					>
						<svg aria-hidden="true" viewBox="0 0 24 24">
							<path d="M4 6.5h16v9H9.6L4 20V6.5Z" />
						</svg>
					</span>
				</span>
				<span
					className="l-debr"
					style={{
						left: "13%",
						bottom: "-6px",
						transform: "rotate(-13deg)",
					}}
				>
					<span
						className="l-ic l-a"
						style={{
							width: "54px",
							height: "54px",
							borderRadius: "17px",
							fontSize: "23px",
							fontWeight: "700",
							color: "#2A2150",
							display: "flex",
							alignItems: "center",
							justifyContent: "center",
						}}
					>
						A
					</span>
				</span>
				<span
					className="l-debr l-mid"
					style={{ left: "23%", bottom: "52px", transform: "rotate(9deg)" }}
				>
					<span
						className="l-ic l-s"
						style={{ width: "44px", height: "44px", borderRadius: "15px" }}
					>
						<svg aria-hidden="true" viewBox="0 0 24 24">
							<rect x="9" y="3" width="6" height="11" rx="3" />
							<path d="M5.5 12a6.5 6.5 0 0 0 13 0" />
							<path d="M12 18.5V21" />
						</svg>
					</span>
				</span>
				<span className="l-debr l-mid" style={{ left: "31%", bottom: "2px" }}>
					<span
						style={{
							display: "flex",
							alignItems: "flex-end",
							gap: "5px",
							height: "40px",
						}}
					>
						<span
							style={{
								width: "9px",
								height: "20px",
								borderRadius: "99px",
								background: "#C9BEFF",
								display: "block",
							}}
						></span>
						<span
							style={{
								width: "9px",
								height: "38px",
								borderRadius: "99px",
								background: "#FF9E3D",
								display: "block",
							}}
						></span>
						<span
							style={{
								width: "9px",
								height: "26px",
								borderRadius: "99px",
								background: "#C9BEFF",
								display: "block",
							}}
						></span>
					</span>
				</span>

				<span
					className="l-debr"
					style={{ right: "6%", bottom: "68px", transform: "rotate(7deg)" }}
				>
					<span
						className="l-ic l-a"
						style={{ width: "46px", height: "46px", borderRadius: "15px" }}
					>
						<svg aria-hidden="true" viewBox="0 0 24 24">
							<path d="M4 4.5h6a3 3 0 0 1 3 3v12a2.5 2.5 0 0 0-2.5-2.5H4v-12Z" />
							<path d="M20 4.5h-6a3 3 0 0 0-3 3v12a2.5 2.5 0 0 1 2.5-2.5H20v-12Z" />
						</svg>
					</span>
				</span>
				<span
					className="l-debr"
					style={{
						right: "13%",
						bottom: "-8px",
						transform: "rotate(12deg)",
					}}
				>
					<span
						className="l-ic l-w"
						style={{ width: "52px", height: "52px", borderRadius: "16px" }}
					>
						<svg aria-hidden="true" viewBox="0 0 24 24">
							<path d="M4 13v-1a3 3 0 0 1 3-3h10a3 3 0 0 1 3 3v1" />
							<path d="M4 13h3.4v5.5H6a2 2 0 0 1-2-2V13Zm16 0h-3.4v5.5H18a2 2 0 0 0 2-2V13Z" />
						</svg>
					</span>
				</span>
				<span
					className="l-debr l-mid"
					style={{
						right: "23%",
						bottom: "50px",
						transform: "rotate(-10deg)",
					}}
				>
					<span
						className="l-ic l-s"
						style={{ width: "44px", height: "44px", borderRadius: "15px" }}
					>
						<svg aria-hidden="true" viewBox="0 0 24 24">
							<path d="M5 12.5l4.5 4.5L19 7.5" />
						</svg>
					</span>
				</span>
				<span
					className="l-debr l-mid"
					style={{
						right: "30%",
						bottom: "4px",
						transform: "rotate(-14deg)",
					}}
				>
					<span
						className="l-chip"
						style={{
							borderColor: "#C9BEFF",
							background: "#fff",
							color: "var(--deep)",
							fontSize: "14px",
							boxShadow: "var(--sh-sm)",
						}}
					>
						B1
					</span>
				</span>

				<span
					className="l-debr"
					style={{
						left: "50%",
						bottom: "-46px",
						transform: "translateX(-50%)",
					}}
				>
					<span
						style={{
							filter: "drop-shadow(0 6px 0 rgba(0,0,0,.3))",
							overflow: "visible",
						}}
					>
						{/* The real mascot component, not the mockup's inlined copy of
						    the same artwork — same viewBox, so the position and scale
						    are unchanged, and `cheer` is the expression this band
						    wants. */}
						<MoMascot variant="cheer" size={152} />
					</span>
				</span>
			</div>

			<div className="l-hillcrest">
				<svg
					viewBox="0 0 1200 172"
					preserveAspectRatio="none"
					aria-hidden="true"
				>
					{/* One ink shape only. The dot field is a single CSS grid on the
			    closing block (see marketing.css): an SVG pattern here would be
			    stretched by preserveAspectRatio="none", so its pitch would
			    scale with the window and stop matching the body's grid. */}
					<path
						d="M0,172 C140,52 380,0 600,0 C820,0 1060,52 1200,172 Z"
						style={{ fill: "var(--inkpad)" }}
					/>
				</svg>
			</div>

			<div className="l-darkside">
				{children ? <div className="l-hillbody">{children}</div> : null}
				<div className="l-footgrid">
					<div className="l-fcol">
						<span className="l-fbrand">
							<svg
								aria-hidden="true"
								width="30"
								height="29"
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
							<span>
								eng<span style={{ color: "#FF9E3D" }}>flex</span>
							</span>
						</span>
						<p className="l-fblurb">{m["landing.footer.blurb"]()}</p>
					</div>

					<div className="l-fcol">
						<h4>{m["landing.footer.account.label"]()}</h4>
						<ul>
							<li>
								<a href={APP_ROUTES.ABOUT}>
									{m["landing.footer.account.about"]()}
								</a>
							</li>
							<li>
								<a href={APP_ROUTES.PRICING}>
									{m["landing.footer.account.pricing"]()}
								</a>
							</li>
							<li>
								<a href={APP_ROUTES.AUTH.SIGN_IN}>
									{m["landing.footer.account.signIn"]()}
								</a>
							</li>
							<li>
								<a href={APP_ROUTES.AUTH.SIGN_UP}>
									{m["landing.footer.account.signUp"]()}
								</a>
							</li>
						</ul>
						<h4 style={{ marginTop: "24px" }}>
							{m["landing.footer.help.label"]()}
						</h4>
						<ul>
							<li>
								<span className="l-pending" title={m["landing.nav.pending"]()}>
									{m["landing.footer.help.contact"]()}
								</span>
							</li>
							<li>
								<span className="l-pending" title={m["landing.nav.pending"]()}>
									{m["landing.footer.help.report"]()}
								</span>
							</li>
						</ul>
					</div>

					<div className="l-fcol">
						<h4>{m["landing.footer.legal.label"]()}</h4>
						<ul>
							<li>
								<span className="l-pending" title={m["landing.nav.pending"]()}>
									{m["landing.footer.legal.privacy"]()}
								</span>
							</li>
							<li>
								<span className="l-pending" title={m["landing.nav.pending"]()}>
									{m["landing.footer.legal.terms"]()}
								</span>
							</li>
							<li>
								<span className="l-pending" title={m["landing.nav.pending"]()}>
									{m["landing.footer.legal.cookies"]()}
								</span>
							</li>
						</ul>
					</div>
				</div>

				<div className="l-frow">
					<span>{m["landing.footer.copyright"]()}</span>
					<span>{m["landing.footer.tagline"]()}</span>
				</div>
			</div>
		</div>
	);
}
