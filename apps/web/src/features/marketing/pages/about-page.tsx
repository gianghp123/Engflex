import { APP_ROUTES } from "#/app/app-route";
import { MoMascot } from "#/components/common/mo-mascot";
import { MarketingFooter } from "#/features/marketing/marketing-footer";
import { m } from "#/paraglide/messages";

/**
 * The About page: the product story, told the way the landing page tells it —
 * the marketing chrome (`MarketingShell`) and the mockup's stylesheet do the
 * styling, so this file is content plus the landing's own band idioms
 * (`l-band`, `l-wrap`, `l-eyebrow`, `l-lede`, `l-bens`/`l-ben`, `l-tiles`).
 *
 * Every claim here is one the landing page already makes, because the landing's
 * copy went through a claim audit and this page did not: 18 lessons A1–C2, three
 * at every level, four steps in each, 15 roleplays across seven topics up to B1,
 * 108 explained questions, writing scored against a model email, pronunciation
 * marked word by word, browser only. Do not add a statistic, a testimonial, a
 * price or a team member here without the evidence to back it.
 */
export function AboutPageMarkup() {
  return (
    <>
      {/* HERO */}
      <div className="l-band l-white">
        <div className="l-wrap">
          <div className="l-hero-split">
            <div>
              <h1>{m["about.hero.title"]()}</h1>
              <p
                className="l-lede"
                style={{ margin: "14px 0 28px", maxWidth: "46ch" }}
              >
                {m["about.hero.body"]()}
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
                  {m["about.hero.cta"]()}
                </a>
                <a className="l-btn l-link" href={APP_ROUTES.AUTH.SIGN_IN}>
                  {m["about.hero.ctaSecondary"]()}
                </a>
              </div>
            </div>
            {/* Mo takes the right column on a wide screen and drops below the
						    copy when the hero stacks. Decorative, so no label. */}
            <div className="l-hero-mo">
              <MoMascot variant="wave" size={240} />
            </div>
          </div>
        </div>
      </div>

      {/* WHY IT EXISTS */}
      <div className="l-band">
        <div className="l-wrap">
          <div className="l-eyebrow">{m["about.problem.eyebrow"]()}</div>
          <h2>{m["about.problem.title"]()}</h2>
          <p
            className="l-lede"
            style={{ margin: "10px 0 0", maxWidth: "64ch" }}
          >
            {m["about.problem.body"]()}
          </p>
        </div>
      </div>

      {/* THE PATH */}
      <div className="l-band l-white">
        <div className="l-wrap">
          <div className="l-eyebrow">{m["about.path.eyebrow"]()}</div>
          <h2>{m["about.path.title"]()}</h2>
          <p
            className="l-lede"
            style={{ margin: "10px 0 34px", maxWidth: "64ch" }}
          >
            {m["about.path.body"]()}
          </p>
          <div className="l-rail">
            <div className="l-nodes">
              {LEVELS.map((level) => (
                <div className="l-node" key={level.code}>
                  <span className={`l-medal ${level.tone}`}>{level.code}</span>
                  <span className="l-count">
                    {m["about.path.lessonsPerLevel"]()}
                  </span>
                </div>
              ))}
            </div>
          </div>
          <div className="l-mergediv" />
          <div className="l-subhead">
            <span>{m["about.path.stepsLabel"]()}</span>
            <b>{m["about.path.stepsMeta"]()}</b>
          </div>
          <div className="l-tiles">
            {STEPS.map((step, index) => (
              <span className="l-tile" key={step.title()}>
                <span className="l-ic l-a" style={{ fontSize: "17px" }}>
                  {index + 1}
                </span>
                <span>
                  <b>{step.title()}</b>
                  <small>{step.body()}</small>
                </span>
              </span>
            ))}
          </div>
        </div>
      </div>

      {/* TALKING, AND WHAT COMES BACK */}
      <div className="l-band">
        <div className="l-wrap">
          <div className="l-eyebrow">{m["about.talking.eyebrow"]()}</div>
          <h2>{m["about.talking.title"]()}</h2>
          <p
            className="l-lede"
            style={{ margin: "10px 0 30px", maxWidth: "64ch" }}
          >
            {m["about.talking.body"]()}
          </p>
          <div className="l-bens">
            {CARDS.map((card) => (
              <div className="l-ben" key={card.title()}>
                <h3>{card.title()}</h3>
                <p>{card.body()}</p>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* WHAT TO EXPECT */}
      <div className="l-band l-white">
        <div className="l-wrap">
          <div className="l-eyebrow">{m["about.expect.eyebrow"]()}</div>
          <h2>{m["about.expect.title"]()}</h2>
          <p
            className="l-lede"
            style={{ margin: "10px 0 30px", maxWidth: "64ch" }}
          >
            {m["about.expect.body"]()}
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

      {/* CLOSING + FOOTER — the landing's band, with this page's call to action. */}
      <MarketingFooter>
        <h1>{m["about.closing.title"]()}</h1>
        <p className="l-lede">{m["about.closing.body"]()}</p>
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
            {m["about.closing.cta"]()}
          </a>
          <a className="l-btn l-link l-onpurple" href={APP_ROUTES.AUTH.SIGN_IN}>
            {m["about.closing.ctaSecondary"]()}
          </a>
        </div>
      </MarketingFooter>
    </>
  );
}

/** Medallion tones follow the landing's path band: A apricot, B purple, C sky. */
const LEVELS = [
  { code: "A1", tone: "l-a" },
  { code: "A2", tone: "l-a" },
  { code: "B1", tone: "l-p" },
  { code: "B2", tone: "l-p" },
  { code: "C1", tone: "l-s" },
  { code: "C2", tone: "l-s" },
] as const;

/** The four steps every lesson runs. Thunks, so the locale resolves at render. */
const STEPS = [
  { title: m["about.path.stepReadTitle"], body: m["about.path.stepReadBody"] },
  { title: m["about.path.stepHearTitle"], body: m["about.path.stepHearBody"] },
  {
    title: m["about.path.stepWriteTitle"],
    body: m["about.path.stepWriteBody"],
  },
  { title: m["about.path.stepSayTitle"], body: m["about.path.stepSayBody"] },
] as const;

/** What the practice gives back. */
const CARDS = [
  {
    title: m["about.talking.cardRoleplayTitle"],
    body: m["about.talking.cardRoleplayBody"],
  },
  {
    title: m["about.talking.cardAnswersTitle"],
    body: m["about.talking.cardAnswersBody"],
  },
  {
    title: m["about.talking.cardPronunciationTitle"],
    body: m["about.talking.cardPronunciationBody"],
  },
] as const;

/** Three facts, including the limit of the voice partner today. */
const FACTS = [
  { title: m["about.expect.timeTitle"], body: m["about.expect.timeBody"] },
  { title: m["about.expect.levelsTitle"], body: m["about.expect.levelsBody"] },
  {
    title: m["about.expect.roleplayTitle"],
    body: m["about.expect.roleplayBody"],
  },
] as const;
