import { cn } from "cn";
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react";
import type { ReactNode } from "react";
import {
	Pagination,
	PaginationContent,
	PaginationEllipsis,
	PaginationItem,
	PaginationLink,
} from "#/components/ui/pagination";
import { m } from "#/paraglide/messages";

/** A gap between numbers: `1 … 4 5 6 … 20`. Keys are the two possible gaps. */
type PagerItem = { kind: "page"; page: number } | { kind: "gap"; key: string };

/** The windowed page numbers, with gaps: ≤7 pages shows all of them. */
function pageItems(page: number, totalPages: number): PagerItem[] {
	if (totalPages <= 7) {
		return Array.from({ length: totalPages }, (_, i) => ({
			kind: "page" as const,
			page: i + 1,
		}));
	}
	const items: PagerItem[] = [{ kind: "page", page: 1 }];
	const from = Math.max(2, page - 1);
	const to = Math.min(totalPages - 1, page + 1);
	if (from > 2) items.push({ kind: "gap", key: "gap-start" });
	for (let p = from; p <= to; p++) items.push({ kind: "page", page: p });
	if (to < totalPages - 1) items.push({ kind: "gap", key: "gap-end" });
	items.push({ kind: "page", page: totalPages });
	return items;
}

/**
 * Numbered pagination built from the shadcn Pagination primitives. Data-driven
 * and router-agnostic: the caller owns the URL via `onPageChange`, so any list
 * page can adopt it. Rendered even for a single page — the control must not
 * pop in and out as filters change the page count. The caller's range/count
 * copy rides in `summary`, so the pager row reads as one unit and no locale
 * string lives in this component.
 */
export function Pager({
	page,
	totalPages,
	onPageChange,
	summary,
	className,
}: {
	page: number;
	totalPages: number;
	onPageChange: (page: number) => void;
	/** Left-hand context for the controls, e.g. "Showing 1–24 of 137". */
	summary?: ReactNode;
	className?: string;
}) {
	if (totalPages < 1) return null;
	return (
		<div
			className={cn(
				"flex flex-wrap items-center justify-between gap-3",
				className,
			)}
		>
			{summary ? (
				<p className="text-xs font-medium text-muted-foreground">{summary}</p>
			) : null}
			<Pagination className="mx-0 w-auto">
				<PaginationContent>
					<PaginationItem>
						<PaginationLink
							asChild
							size="default"
							aria-label={m["common.pager.previous"]()}
							className="gap-1 px-2.5 sm:pl-2.5"
						>
							<button
								type="button"
								disabled={page <= 1}
								onClick={() => onPageChange(page - 1)}
							>
								<ChevronLeftIcon />
								<span className="hidden sm:block">
									{m["common.pager.previous"]()}
								</span>
							</button>
						</PaginationLink>
					</PaginationItem>
					{pageItems(page, totalPages).map((item) =>
						item.kind === "gap" ? (
							<PaginationItem key={item.key}>
								<PaginationEllipsis />
							</PaginationItem>
						) : (
							<PaginationItem key={item.page}>
								<PaginationLink asChild isActive={item.page === page}>
									<button type="button" onClick={() => onPageChange(item.page)}>
										{item.page}
									</button>
								</PaginationLink>
							</PaginationItem>
						),
					)}
					<PaginationItem>
						<PaginationLink
							asChild
							size="default"
							aria-label={m["common.pager.next"]()}
							className="gap-1 px-2.5 sm:pr-2.5"
						>
							<button
								type="button"
								disabled={page >= totalPages}
								onClick={() => onPageChange(page + 1)}
							>
								<span className="hidden sm:block">
									{m["common.pager.next"]()}
								</span>
								<ChevronRightIcon />
							</button>
						</PaginationLink>
					</PaginationItem>
				</PaginationContent>
			</Pagination>
		</div>
	);
}
