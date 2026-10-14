import type { ReactNode } from "react";

/**
 * Shared page layout: content container + optional header (bare icon,
 * Quicksand title, description as peach bubble or muted subtitle) + body.
 * The icon is rendered as-is — pass a container-free `MoMascot`, never a
 * boxed tile.
 */
export function PageHero({
	icon,
	title,
	description,
}: {
	icon: ReactNode;
	title: string;
	description: ReactNode;
}) {
	return (
		<div className="flex items-center gap-4">
			<span className="inline-flex shrink-0">{icon}</span>
			<div className="min-w-0">
				<h1 className="text-[28px] leading-[36px] font-bold text-foreground md:text-[38px] md:leading-[46px]">
					{title}
				</h1>
				{description && (
					<p className="bubble bubble-own mt-2 w-fit text-[15px] font-bold">
						{description}
					</p>
				)}
			</div>
		</div>
	);
}

export interface PageHeroSpec {
	icon: ReactNode;
	title: string;
	description: ReactNode;
}

/**
 * The app's one two-pane split: main work area (7) + side panel (5), stacked
 * below `lg`. Every activity page pairs its material with the interaction it
 * drives through this component so the reading rhythm never shifts between
 * reading/listening/writing/speaking. Cells stretch to the row height so a
 * sticky aside (`lg:sticky`) can anchor beside taller main content.
 */
export function PageSplit({
	main,
	aside,
}: {
	main: ReactNode;
	aside: ReactNode;
}) {
	return (
		<div className="grid items-stretch gap-4 lg:grid-cols-12">
			<div className="lg:col-span-7">{main}</div>
			<div className="lg:col-span-5">{aside}</div>
		</div>
	);
}

export function PageLayout({
	hero,
	action,
	children,
}: {
	hero?: PageHeroSpec | null;
	/** Left-of-hero slot: back navigation or a page-level call to action.
	 *  The caller owns label + link (a `Button asChild` + `Link` block) —
	 *  this component only positions it. */
	action?: ReactNode;
	children?: ReactNode;
}) {
	return (
		<div className="container-content flex flex-col gap-6 py-8">
			{action ? <div className="flex items-center gap-2">{action}</div> : null}
			{hero ? (
				<PageHero
					icon={hero.icon}
					title={hero.title}
					description={hero.description}
				/>
			) : null}
			{children}
		</div>
	);
}
