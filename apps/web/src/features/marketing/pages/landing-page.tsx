import { APP_ROUTES } from "#/app/app-route";
import { MoMascot } from "#/components/common/mo-mascot";
import { SHOW_TESTIMONIALS } from "#/features/marketing/content-gates";
import { MarketingFooter } from "#/features/marketing/marketing-footer";
import { m } from "#/paraglide/messages";

/**
 * The landing page's own sections: hero, level path, the merged practice band, the
 * gated testimonials, and the closing call to action it hands to the footer band.
 *
 * Transcribed from the approved mockup
 * (`.superpowers/brainstorm/.../25-full-page-final.html`) with its own stylesheet
 * (marketing.css, `.landing`-scoped, `l-`-prefixed classes).
 *
 * Why not rebuilt from Tailwind utilities: a first port re-expressed every band
 * in utilities and drifted from the approved design — the closing band's crest
 * and the hero stage were measurably wrong. The page is static, so the mockup's
 * own markup and CSS are the source of truth; only copy (m["…"]) and links
 * (APP_ROUTES) are dynamic.
 *
 * Shared chrome lives outside this file now: `MarketingShell` owns `.landing` and
 * `.l-page` (and the palette variables they carry), `MarketingHeader` the top bar,
 * `MarketingFooter` the dark closing band. Regenerating from the mockup means
 * re-running that split — the header and the closing band must not be pasted back
 * in here.
 *
 * Learning content (lesson titles, scenario names, the sample exchange) stays in
 * English on purpose — it is material, not UI chrome.
 */
export function LandingPageMarkup() {
	return (
		<>
			{/* 1 HERO */}
			<div className="l-r4h">
				<div className="l-r4h-inner">
					<div className="l-r4h-copy">
						<div className="l-r4h-eyebrow">{m["landing.hero.eyebrow"]()}</div>
						<h1>{m["landing.hero.title"]()}</h1>
						<div className="l-r4h-badges">
							<span className="l-r4h-badge">
								<svg
									aria-hidden="true"
									style={{ flexShrink: "0" }}
									width="22"
									height="32"
									viewBox="0 0 24 34"
									fill="none"
								>
									<path
										d="M18 3C8 7 5 19 10 28"
										stroke="#E0A93B"
										strokeWidth="2.2"
										strokeLinecap="round"
									/>
									<ellipse
										cx="7.6"
										cy="9"
										rx="3.3"
										ry="1.9"
										transform="rotate(-38 7.6 9)"
										fill="#FFC857"
									/>
									<ellipse
										cx="5"
										cy="15.5"
										rx="3.3"
										ry="1.9"
										transform="rotate(-16 5 15.5)"
										fill="#FFC857"
									/>
									<ellipse
										cx="6.6"
										cy="22"
										rx="3.3"
										ry="1.9"
										transform="rotate(12 6.6 22)"
										fill="#FFC857"
									/>
								</svg>
								<span>
									<b>18</b>
									<small>
										{m["landing.hero.badgeLessons"]()}
										<br />
										{m["landing.hero.badgeLevels"]()}
									</small>
								</span>
								<svg
									aria-hidden="true"
									style={{ flexShrink: "0", transform: "scaleX(-1)" }}
									width="22"
									height="32"
									viewBox="0 0 24 34"
									fill="none"
								>
									<path
										d="M18 3C8 7 5 19 10 28"
										stroke="#E0A93B"
										strokeWidth="2.2"
										strokeLinecap="round"
									/>
									<ellipse
										cx="7.6"
										cy="9"
										rx="3.3"
										ry="1.9"
										transform="rotate(-38 7.6 9)"
										fill="#FFC857"
									/>
									<ellipse
										cx="5"
										cy="15.5"
										rx="3.3"
										ry="1.9"
										transform="rotate(-16 5 15.5)"
										fill="#FFC857"
									/>
									<ellipse
										cx="6.6"
										cy="22"
										rx="3.3"
										ry="1.9"
										transform="rotate(12 6.6 22)"
										fill="#FFC857"
									/>
								</svg>
							</span>
							<span className="l-r4h-badge">
								<svg
									aria-hidden="true"
									style={{ flexShrink: "0" }}
									width="22"
									height="32"
									viewBox="0 0 24 34"
									fill="none"
								>
									<path
										d="M18 3C8 7 5 19 10 28"
										stroke="#E0A93B"
										strokeWidth="2.2"
										strokeLinecap="round"
									/>
									<ellipse
										cx="7.6"
										cy="9"
										rx="3.3"
										ry="1.9"
										transform="rotate(-38 7.6 9)"
										fill="#FFC857"
									/>
									<ellipse
										cx="5"
										cy="15.5"
										rx="3.3"
										ry="1.9"
										transform="rotate(-16 5 15.5)"
										fill="#FFC857"
									/>
									<ellipse
										cx="6.6"
										cy="22"
										rx="3.3"
										ry="1.9"
										transform="rotate(12 6.6 22)"
										fill="#FFC857"
									/>
								</svg>
								<span>
									<b>108</b>
									<small>
										{m["landing.hero.badgeQuestions"]()}
										<br />
										{m["landing.hero.badgeExplained"]()}
									</small>
								</span>
								<svg
									aria-hidden="true"
									style={{ flexShrink: "0", transform: "scaleX(-1)" }}
									width="22"
									height="32"
									viewBox="0 0 24 34"
									fill="none"
								>
									<path
										d="M18 3C8 7 5 19 10 28"
										stroke="#E0A93B"
										strokeWidth="2.2"
										strokeLinecap="round"
									/>
									<ellipse
										cx="7.6"
										cy="9"
										rx="3.3"
										ry="1.9"
										transform="rotate(-38 7.6 9)"
										fill="#FFC857"
									/>
									<ellipse
										cx="5"
										cy="15.5"
										rx="3.3"
										ry="1.9"
										transform="rotate(-16 5 15.5)"
										fill="#FFC857"
									/>
									<ellipse
										cx="6.6"
										cy="22"
										rx="3.3"
										ry="1.9"
										transform="rotate(12 6.6 22)"
										fill="#FFC857"
									/>
								</svg>
							</span>
						</div>
						<div
							style={{
								display: "flex",
								alignItems: "center",
								gap: "18px",
								flexWrap: "wrap",
							}}
						>
							<a className="l-r4h-btn" href={APP_ROUTES.AUTH.SIGN_UP}>
								{m["landing.hero.cta"]()}
							</a>
							<a className="l-r4h-btn2" href={APP_ROUTES.AUTH.SIGN_IN}>
								{m["landing.hero.ctaSecondary"]()}
							</a>
						</div>
					</div>

					<div className="l-r4h-stage">
						<span
							className="l-r4h-ring"
							style={{
								right: "30px",
								bottom: "-40px",
								width: "266px",
								height: "266px",
							}}
						></span>
						<span
							className="l-r4h-ring"
							style={{
								right: "56px",
								bottom: "-14px",
								width: "214px",
								height: "214px",
								borderColor: "rgba(255,255,255,.13)",
							}}
						></span>

						<div className="l-r4h-dev">
							<div className="l-bar">
								<i style={{ background: "#FF6B6B" }}></i>
								<i style={{ background: "#FFC24D" }}></i>
								<i style={{ background: "#5BD28B" }}></i>
								<span className="l-url">engflex.app / voice</span>
							</div>
							<div
								style={{
									background: "#fff",
									borderRadius: "13px",
									overflow: "hidden",
								}}
							>
								<div
									style={{
										background: "#FFE2C5",
										padding: "9px 11px",
										fontSize: "11.5px",
										fontWeight: "700",
									}}
								>
									What can I get for you today?
								</div>
								<div
									style={{
										background: "#F2EBFF",
										padding: "9px 11px",
										fontSize: "11.5px",
										fontWeight: "700",
										textAlign: "right",
									}}
								>
									One coffee, please.
								</div>
								<div
									style={{
										padding: "9px 11px",
										background: "#FBF8FF",
										fontSize: "11px",
										fontWeight: "600",
										borderTop: "2px solid var(--line)",
									}}
								>
									I want &rarr; <b>I'd like</b>
								</div>
							</div>
						</div>

						<span className="l-r4h-mo">
							{/* Mo is the real component now. The three sparkles stay as their own layer:
					    they are decoration placed in the mascot's own viewBox, and the `rest`
					    variant does not carry them. They do not overlap the head, so the
					    painting order the inline copy had is irrelevant. */}
							<svg
								aria-hidden="true"
								width="196"
								height="196"
								viewBox="-8 -16 116 116"
								style={{ position: "absolute", inset: 0, overflow: "visible" }}
							>
								<path
									d="M50 -12Q50 -8 54 -8Q50 -8 50 -4Q50 -8 46 -8Q50 -8 50 -12Z"
									fill="#FFC857"
								/>
								<path
									d="M-2 30Q-2 35 3 35Q-2 35 -2 40Q-2 35 -7 35Q-2 35 -2 30Z"
									fill="#FFC857"
								/>
								<path
									d="M104 66Q104 71 109 71Q104 71 104 76Q104 71 99 71Q104 71 104 66Z"
									fill="#FFC857"
								/>
							</svg>
							<MoMascot variant="wave" size={196} />
						</span>
						<span className="l-r4h-bub">{m["landing.hero.scared"]()}</span>
						<span className="l-r4h-say">{m["landing.hero.reassure"]()}</span>
					</div>
				</div>

				<div className="l-r4h-medals">
					<span className="l-r4h-medal l-a">A1</span>
					<span className="l-r4h-medal l-a">A2</span>
					<span className="l-r4h-medal l-p">B1</span>
					<span className="l-r4h-medal l-p">B2</span>
					<span className="l-r4h-medal l-w">C1</span>
					<span className="l-r4h-medal l-w">C2</span>
				</div>
			</div>

			{/* 2 LEVEL PATH */}
			<div className="l-band l-white" id="levels">
				<div className="l-wrap">
					<div className="l-mo-peek">
						<MoMascot variant="heart" size={46} />
						<span className="l-say">{m["landing.path.say"]()}</span>
					</div>

					<h2>{m["landing.path.title"]()}</h2>
					<p className="l-lede" style={{ margin: "10px 0 34px" }}>
						{m["landing.path.body"]()}
					</p>

					<div className="l-rail">
						<div className="l-nodes">
							<div className="l-node">
								<span className="l-medal l-a">A1</span>
								<span className="l-count">
									{m["landing.path.lessonsPerLevel"]()}
								</span>
								<span className="l-lessons">
									<span>Introductions</span>
									<span>Numbers & time</span>
									<span>Places & directions</span>
								</span>
							</div>
							<div className="l-node">
								<span className="l-medal l-a">A2</span>
								<span className="l-count">
									{m["landing.path.lessonsPerLevel"]()}
								</span>
								<span className="l-lessons">
									<span>Food & dining</span>
									<span>Travel & hotels</span>
									<span>Shopping & services</span>
								</span>
							</div>
							<div className="l-node">
								<span className="l-medal l-p">B1</span>
								<span className="l-count">
									{m["landing.path.lessonsPerLevel"]()}
								</span>
								<span className="l-lessons">
									<span>Appointments</span>
									<span>Phone calls</span>
									<span>Office routines</span>
								</span>
							</div>
							<div className="l-node">
								<span className="l-medal l-p">B2</span>
								<span className="l-count">
									{m["landing.path.lessonsPerLevel"]()}
								</span>
								<span className="l-lessons">
									<span>Meetings</span>
									<span>Emails & reports</span>
									<span>Presentations</span>
								</span>
							</div>
							<div className="l-node">
								<span className="l-medal l-s">C1</span>
								<span className="l-count">
									{m["landing.path.lessonsPerLevel"]()}
								</span>
								<span className="l-lessons">
									<span>Leading talks</span>
									<span>Correspondence</span>
									<span>Persuasion</span>
								</span>
							</div>
							<div className="l-node">
								<span className="l-medal l-s">C2</span>
								<span className="l-count">
									{m["landing.path.lessonsPerLevel"]()}
								</span>
								<span className="l-lessons">
									<span>Idioms & register</span>
									<span>Argumentation</span>
									<span>Executive presence</span>
								</span>
							</div>
						</div>
					</div>
				</div>
			</div>

			{/* 3 · PRACTICE (merged: was voice topics + benefits) */}
			<div className="l-band" id="voice">
				<div className="l-wrap">
					<h2>{m["landing.practice.title"]()}</h2>
					<p
						className="l-lede"
						style={{ margin: "10px 0 0", maxWidth: "64ch" }}
					>
						{m["landing.practice.body"]()}
					</p>

					<div className="l-subhead" style={{ margin: "28px 0 16px" }}>
						<span>{m["landing.practice.roleplayLabel"]()}</span>
						<b>{m["landing.practice.roleplayMeta"]()}</b>
					</div>
					<div className="l-tiles">
						<span className="l-tile">
							<span className="l-ic l-a">
								<svg aria-hidden="true" viewBox="0 0 24 24">
									<path d="M4 9h11v5a5 5 0 0 1-5 5H9a5 5 0 0 1-5-5V9Z" />
									<path d="M15 10h2.2a2.8 2.8 0 0 1 0 5.6H15" />
									<path d="M7 5.5v1.2M10.5 4.6v2.1" />
								</svg>
							</span>
							<span>
								<b>Food & drink</b>
								<small>
									{m["landing.practice.roleplayCount"]({ count: 2 })}
								</small>
							</span>
						</span>
						<span className="l-tile">
							<span className="l-ic l-s">
								<svg aria-hidden="true" viewBox="0 0 24 24">
									<path d="M5 8h14l-1.1 11.2H6.1L5 8Z" />
									<path d="M8.6 8a3.4 3.4 0 0 1 6.8 0" />
								</svg>
							</span>
							<span>
								<b>Shopping & markets</b>
								<small>
									{m["landing.practice.roleplayCount"]({ count: 3 })}
								</small>
							</span>
						</span>
						<span className="l-tile">
							<span className="l-ic l-p">
								<svg aria-hidden="true" viewBox="0 0 24 24">
									<path d="M3.4 16.5l1.8-5.6A2.4 2.4 0 0 1 7.5 9.2h9a2.4 2.4 0 0 1 2.3 1.7l1.8 5.6v3.2h-2.7v-2H6.1v2H3.4v-3.2Z" />
									<circle cx="7.4" cy="16.2" r="1.1" />
									<circle cx="16.6" cy="16.2" r="1.1" />
								</svg>
							</span>
							<span>
								<b>Getting around town</b>
								<small>
									{m["landing.practice.roleplayCount"]({ count: 2 })}
								</small>
							</span>
						</span>
						<span className="l-tile">
							<span className="l-ic l-a">
								<svg aria-hidden="true" viewBox="0 0 24 24">
									<path d="M3 12.6l18-8.2-7.6 17.4-2.6-6.6L3 12.6Z" />
								</svg>
							</span>
							<span>
								<b>Travel & hotels</b>
								<small>
									{m["landing.practice.roleplayCount"]({ count: 3 })}
								</small>
							</span>
						</span>
						<span className="l-tile">
							<span className="l-ic l-s">
								<svg aria-hidden="true" viewBox="0 0 24 24">
									<path
										d="M8.4 5.4h7.2a4.3 4.3 0 0 1 0 8.6H8.4a4.3 4.3 0 0 1 0-8.6Z"
										transform="rotate(-40 12 9.7)"
									/>
									<path d="M9.6 7.9l4.8 4.8" transform="rotate(-40 12 9.7)" />
								</svg>
							</span>
							<span>
								<b>Health & pharmacy</b>
								<small>
									{m["landing.practice.roleplayCount"]({ count: 2 })}
								</small>
							</span>
						</span>
						<span className="l-tile">
							<span className="l-ic l-p">
								<svg aria-hidden="true" viewBox="0 0 24 24">
									<path d="M4 6.5h16v9H9.6L4 20V6.5Z" />
								</svg>
							</span>
							<span>
								<b>Social & small talk</b>
								<small>
									{m["landing.practice.roleplayCount"]({ count: 2 })}
								</small>
							</span>
						</span>
						<span className="l-tile">
							<span className="l-ic l-s">
								<svg aria-hidden="true" viewBox="0 0 24 24">
									<rect x="3" y="6" width="18" height="13" rx="3" />
									<path d="M3 10.4h18" />
								</svg>
							</span>
							<span>
								<b>Services & money</b>
								<small>{m["landing.practice.roleplayCountOne"]()}</small>
							</span>
						</span>
						<span className="l-tile l-soon">
							<span>{m["landing.practice.moreComing"]()}</span>
						</span>
					</div>

					<div className="l-mergediv"></div>

					<div className="l-subhead">
						<span>{m["landing.practice.lessonLabel"]()}</span>
						<b>{m["landing.practice.lessonMeta"]()}</b>
					</div>
					<div className="l-bens">
						<div className="l-ben">
							<div className="l-ill">
								<div style={{ display: "flex", gap: "10px" }}>
									<span className="l-ic l-p">
										<svg aria-hidden="true" viewBox="0 0 24 24">
											<path d="M4 4.5h6a3 3 0 0 1 3 3v12a2.5 2.5 0 0 0-2.5-2.5H4v-12Z" />
											<path d="M20 4.5h-6a3 3 0 0 0-3 3v12a2.5 2.5 0 0 1 2.5-2.5H20v-12Z" />
										</svg>
									</span>
									<span className="l-ic l-s">
										<svg aria-hidden="true" viewBox="0 0 24 24">
											<path d="M4 13v-1a3 3 0 0 1 3-3h10a3 3 0 0 1 3 3v1" />
											<path d="M4 13h3.4v5.5H6a2 2 0 0 1-2-2V13Zm16 0h-3.4v5.5H18a2 2 0 0 0 2-2V13Z" />
										</svg>
									</span>
									<span className="l-ic l-a">
										<svg aria-hidden="true" viewBox="0 0 24 24">
											<path d="M4 20h16" />
											<path d="M5 16.5l10.4-10.4a2 2 0 0 1 2.8 2.8L7.8 19.3 4.6 20l.4-3.5Z" />
										</svg>
									</span>
									<span className="l-ic l-p">
										<svg aria-hidden="true" viewBox="0 0 24 24">
											<rect x="9" y="3" width="6" height="11" rx="3" />
											<path d="M5.5 12a6.5 6.5 0 0 0 13 0" />
											<path d="M12 18.5V21" />
										</svg>
									</span>
								</div>
							</div>
							<h3>{m["landing.practice.skillsTitle"]()}</h3>
							<p>{m["landing.practice.skillsBody"]()}</p>
						</div>

						<div className="l-ben">
							<div className="l-ill">
								<div style={{ width: "100%" }}>
									<div
										style={{
											display: "flex",
											alignItems: "center",
											gap: "8px",
											fontSize: "14px",
											fontWeight: "700",
											marginBottom: "9px",
										}}
									>
										<span
											className="l-ic l-p"
											style={{
												width: "22px",
												height: "22px",
												borderRadius: "8px",
												boxShadow: "none",
											}}
										>
											<svg
												aria-hidden="true"
												viewBox="0 0 24 24"
												style={{ width: "13px", height: "13px" }}
											>
												<path d="M5 12.5l4.5 4.5L19 7.5" />
											</svg>
										</span>
										What is Kenji's job?
									</div>
									<div
										style={{
											fontSize: "13.5px",
											fontWeight: "600",
											color: "var(--muted)",
											lineHeight: "1.5",
											paddingLeft: "30px",
										}}
									>
										A sales assistant — the text says he sells to shops in
										Osaka.
									</div>
								</div>
							</div>
							<h3>{m["landing.practice.gradingTitle"]()}</h3>
							<p>{m["landing.practice.gradingBody"]()}</p>
						</div>

						<div className="l-ben">
							<div className="l-ill">
								<div style={{ textAlign: "center" }}>
									<div
										style={{
											fontSize: "14.5px",
											fontWeight: "700",
											marginBottom: "10px",
										}}
									>
										Good morning. My name is Kenji Sato.
									</div>
									<div
										style={{
											display: "flex",
											gap: "6px",
											justifyContent: "center",
											flexWrap: "wrap",
										}}
									>
										<span
											className="l-chip l-xs"
											style={{
												borderColor: "var(--sky)",
												color: "var(--skyink)",
											}}
										>
											good ✓
										</span>
										<span className="l-chip l-xs">morning ✓</span>
										<span className="l-chip l-xs">my name ✓</span>
										<span
											className="l-chip l-xs"
											style={{
												borderColor: "var(--apricot)",
												color: "var(--amber)",
											}}
										>
											Kenji ⚠
										</span>
									</div>
								</div>
							</div>
							<h3>{m["landing.practice.pronunciationTitle"]()}</h3>
							<p>{m["landing.practice.pronunciationBody"]()}</p>
						</div>
					</div>
				</div>
			</div>

			{/* Gated: renders nothing until real, attributable quotes exist. */}
			{SHOW_TESTIMONIALS && (
				<>
					{/* 5 TESTIMONIALS */}
					<div className="l-band l-white">
						<div className="l-wrap">
							<h2 style={{ marginBottom: "8px" }}>
								{m["landing.testimonials.title"]()}
							</h2>
							<p className="l-lede" style={{ marginBottom: "28px" }}>
								{m["landing.testimonials.body"]()}
							</p>
							<div className="l-quotes">
								<div className="l-q">
									<span className="l-glyph">“</span>
									<blockquote>
										{m["landing.testimonials.oneQuote"]()}
									</blockquote>
									<div className="l-who">
										<span
											className="l-av"
											style={{ background: "var(--purple)" }}
										>
											LN
										</span>
										<span>
											<b>{m["landing.testimonials.oneName"]()}</b>
											<small>{m["landing.testimonials.oneLevel"]()}</small>
										</span>
									</div>
								</div>
								<div className="l-q">
									<span className="l-glyph">“</span>
									<blockquote>
										{m["landing.testimonials.twoQuote"]()}
									</blockquote>
									<div className="l-who">
										<span
											className="l-av"
											style={{
												background: "var(--apricot)",
												color: "var(--ink)",
											}}
										>
											TP
										</span>
										<span>
											<b>{m["landing.testimonials.twoName"]()}</b>
											<small>{m["landing.testimonials.twoLevel"]()}</small>
										</span>
									</div>
								</div>
								<div className="l-q">
									<span className="l-glyph">“</span>
									<blockquote>
										{m["landing.testimonials.threeQuote"]()}
									</blockquote>
									<div className="l-who">
										<span className="l-av" style={{ background: "var(--sky)" }}>
											MĐ
										</span>
										<span>
											<b>{m["landing.testimonials.threeName"]()}</b>
											<small>{m["landing.testimonials.threeLevel"]()}</small>
										</span>
									</div>
								</div>
							</div>
						</div>
					</div>
				</>
			)}

			{/* 6 · CLOSING + FOOTER — one dark band. The footer owns the ink, the crest
			    and the dot field that spans them; the page supplies only this closing
			    call to action. */}
			<MarketingFooter>
				<h1>{m["landing.hero.title"]()}</h1>
				<p className="l-lede">{m["landing.closing.body"]()}</p>
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
						{m["landing.closing.cta"]()}
					</a>
					<a className="l-btn l-link l-onpurple" href={APP_ROUTES.AUTH.SIGN_IN}>
						{m["landing.closing.ctaSecondary"]()}
					</a>
				</div>
			</MarketingFooter>
		</>
	);
}
