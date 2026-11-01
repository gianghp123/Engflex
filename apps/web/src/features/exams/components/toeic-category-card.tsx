import { Link } from "@tanstack/react-router";
import { ArrowRight, BookOpen } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { m } from "#/paraglide/messages";

export interface ToeicCategoryCardItem {
	id: string;
	title: string;
	tags: string[];
	testsCount: number;
	iconBgColor: string;
}

export function ToeicCategoryCard({ item }: { item: ToeicCategoryCardItem }) {
	return (
		<div className="relative flex items-center justify-between gap-4 rounded-2xl border border-border/80 bg-card p-5 shadow-sm transition-all hover:border-primary/40 hover:shadow-md">
			<div className="flex min-w-0 flex-1 flex-col gap-3">
				<h3 className="truncate text-base font-bold text-card-foreground md:text-lg">
					{item.title}
				</h3>

				<div className="flex flex-wrap items-center gap-1.5">
					{item.tags.map((tag) => (
						<span
							key={tag}
							className="inline-flex items-center rounded-md bg-muted px-2 py-0.5 text-xs font-semibold text-muted-foreground"
						>
							{tag}
						</span>
					))}
				</div>

				<div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground md:text-sm">
					<span className="inline-flex items-center gap-1.5">
						<BookOpen className="size-4" aria-hidden="true" />
						{m["exams.toeic.hub.testsCount"]({ count: item.testsCount })}
					</span>
				</div>

				<div className="pt-1">
					<Link
						to={APP_ROUTES.EXAMS.TOEIC.CATEGORY}
						params={{ categoryId: item.id }}
						className="inline-flex items-center gap-1.5 text-sm font-bold text-foreground transition-colors hover:text-primary"
					>
						<span>{m["exams.toeic.hub.startPractice"]()}</span>
						<ArrowRight className="size-4" aria-hidden="true" />
					</Link>
				</div>
			</div>

			<div
				className={`flex size-14 shrink-0 items-center justify-center rounded-2xl text-white shadow-sm md:size-16 ${item.iconBgColor}`}
				aria-hidden="true"
			>
				<BookOpen className="size-7 md:size-8" />
			</div>
		</div>
	);
}
