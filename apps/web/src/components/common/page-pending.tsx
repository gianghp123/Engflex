import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { Skeleton } from "#/components/ui/skeleton";

/** Stable keys for the fixed three-card grid (no index keys). */
const CARD_KEYS = ["first", "second", "third"] as const;

/**
 * Router-wide pending UI (`defaultPendingComponent`). A route that declares no
 * `pendingComponent` of its own renders this while its `beforeLoad`/`loader`
 * settle, so the previous page is never left on screen under a new URL.
 *
 * Zero props on purpose: the router renders a pending component as
 * `jsx(Component, {})`, and a prop that switches on page type is the shape the
 * component conventions forbid. Pure shapes and `aria-busy`, no strings — the
 * same rule `CardShell` follows for card-level pending frames.
 *
 * The shapes mirror what the hubs actually paint — `PageHero`'s row (84px
 * mascot beside a title/description column), the 3-line card `CardSkeletons`
 * draws, and the list block most hubs end with — so the swap to real content
 * shifts as little as a generic frame can. It is deliberately not a pixel
 * match for any one page: a route that wants that declares its own
 * `pendingComponent`.
 */
export function PagePending() {
	return (
		<PageLayout>
			<div aria-busy className="flex flex-col gap-6">
				{/* Hero row: same frame as PageHero (icon beside title/description). */}
				<div className="flex items-center gap-4">
					<Skeleton className="size-[84px] shrink-0 rounded-full" />
					<div className="flex min-w-0 flex-1 flex-col gap-3">
						<Skeleton className="h-9 w-72 max-w-full" />
						<Skeleton className="h-5 w-96 max-w-full" />
					</div>
				</div>

				{/* Card grid: the same 3-line card `CardSkeletons` draws. */}
				<div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
					{CARD_KEYS.map((key) => (
						<CardShell
							key={key}
							pending
							className="p-5"
							skeleton={
								<div className="flex flex-col gap-2">
									<Skeleton className="h-3 w-24" />
									<Skeleton className="h-7 w-16" />
									<Skeleton className="h-3 w-32" />
								</div>
							}
						/>
					))}
				</div>

				{/* List block: stands in for the rows/table most hubs end with. */}
				<CardShell
					pending
					className="p-5"
					skeleton={
						<div className="flex flex-col gap-3">
							<Skeleton className="h-4 w-1/3" />
							<Skeleton className="h-4 w-2/3" />
							<Skeleton className="h-4 w-1/2" />
						</div>
					}
				/>
			</div>
		</PageLayout>
	);
}
