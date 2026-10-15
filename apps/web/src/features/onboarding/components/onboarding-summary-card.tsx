import { m } from "#/paraglide/messages";

/**
 * Decorative radar. The shape is fixed styling, not a measurement — onboarding
 * collects preferences, it does not assess speech — so the axes are labelled as
 * the dimensions a voice session will track, with no fabricated values shown.
 */
const AXES = [
	{
		key: "precision",
		label: () => m["onboarding.summary.axes.precision"](),
		value: 88,
	},
	{
		key: "lexical",
		label: () => m["onboarding.summary.axes.lexical"](),
		value: 76,
	},
	{
		key: "cadence",
		label: () => m["onboarding.summary.axes.cadence"](),
		value: 84,
	},
	{
		key: "stress",
		label: () => m["onboarding.summary.axes.stress"](),
		value: 82,
	},
	{
		key: "latency",
		label: () => m["onboarding.summary.axes.latency"](),
		value: 70,
	},
	{
		key: "clarity",
		label: () => m["onboarding.summary.axes.clarity"](),
		value: 92,
	},
] as const;

const CENTER = 110;
const RADIUS = 76;

function point(index: number, ratio: number): [number, number] {
	const angle = (Math.PI * 2 * index) / AXES.length - Math.PI / 2;
	return [
		CENTER + Math.cos(angle) * RADIUS * ratio,
		CENTER + Math.sin(angle) * RADIUS * ratio,
	];
}

const polygon = AXES.map((axis, index) => point(index, axis.value / 100))
	.map(([x, y]) => `${x.toFixed(1)},${y.toFixed(1)}`)
	.join(" ");

function SummaryRow({ label, value }: { label: string; value: string }) {
	return (
		<div className="rounded-xl border bg-card px-4 py-3">
			<dt className="text-[11px] font-semibold tracking-wide text-muted-foreground uppercase">
				{label}
			</dt>
			<dd className="mt-0.5 text-sm font-semibold text-foreground">{value}</dd>
		</div>
	);
}

export function OnboardingSummaryCard({
	level,
	goals,
	topics,
	commitment,
}: {
	level: { title: string; range: string };
	goals: string[];
	topics: string[];
	commitment: string;
}) {
	return (
		<div className="flex flex-col gap-5 p-5">
			<div>
				<h2 className="text-lg font-extrabold tracking-tight">
					{m["onboarding.summary.title"]()}
				</h2>
				<p className="mt-1 text-xs text-muted-foreground">
					{m["onboarding.summary.subtitle"]()}
				</p>
			</div>
			<dl className="grid gap-3 sm:grid-cols-2">
				<SummaryRow
					label={m["onboarding.summary.level"]()}
					value={`${level.title} · ${level.range}`}
				/>
				<SummaryRow
					label={m["onboarding.summary.commitment"]()}
					value={commitment}
				/>
				<SummaryRow
					label={m["onboarding.summary.goals"]()}
					value={goals.join(", ")}
				/>
				<SummaryRow
					label={m["onboarding.summary.topics"]()}
					value={topics.join(", ")}
				/>
			</dl>
			<div className="flex flex-col items-center gap-3">
				<svg
					viewBox="0 0 220 220"
					className="h-52 w-52 shrink-0"
					role="img"
					aria-label={m["onboarding.summary.chartLabel"]()}
				>
					{[0.33, 0.66, 1].map((ratio) => (
						<polygon
							key={ratio}
							points={AXES.map((_, index) =>
								point(index, ratio)
									.map((n) => n.toFixed(1))
									.join(","),
							).join(" ")}
							className="fill-none stroke-border"
							strokeWidth="1"
						/>
					))}
					{AXES.map((axis, index) => {
						const [x, y] = point(index, 1);
						return (
							<line
								key={axis.key}
								x1={CENTER}
								y1={CENTER}
								x2={x}
								y2={y}
								className="stroke-border"
								strokeWidth="1"
							/>
						);
					})}
					<polygon
						points={polygon}
						className="fill-primary/10 stroke-primary"
						strokeWidth="2"
					/>
					{AXES.map((axis, index) => {
						const [x, y] = point(index, 1.16);
						return (
							<text
								key={axis.key}
								x={x}
								y={y}
								textAnchor="middle"
								dominantBaseline="middle"
								className="fill-muted-foreground text-[9px] font-semibold tracking-wide"
							>
								{axis.label()}
							</text>
						);
					})}
				</svg>
				<p className="max-w-xs text-center text-[11px] text-muted-foreground">
					{m["onboarding.summary.previewCaption"]()}
				</p>
			</div>
		</div>
	);
}
