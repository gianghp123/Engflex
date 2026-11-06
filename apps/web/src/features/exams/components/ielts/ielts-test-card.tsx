import { Link } from "@tanstack/react-router";
import { Clock, FileText } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { m } from "#/paraglide/messages";

export interface IeltsTestCardItem {
	id: string;
	categoryId: string;
	title: string;
	skill?: string | null;
	questionsCount: number;
	durationMinutes: number;
}

export function IeltsTestCard({ item }: { item: IeltsTestCardItem }) {
	return (
		<div className="flex flex-col gap-3 rounded-2xl border border-border/80 bg-card p-4 shadow-sm transition-all hover:border-primary/40 hover:shadow-md sm:flex-row sm:items-center sm:justify-between md:p-5">
			<div className="flex min-w-0 flex-1 flex-col gap-2.5">
				<div className="flex flex-wrap items-center gap-2">
					<h3 className="truncate text-sm font-bold text-card-foreground md:text-base">
						{item.title}
					</h3>
					{item.skill && (
						<span className="inline-flex items-center rounded-md bg-sky-500/10 px-2 py-0.5 text-xs font-semibold text-sky-600 dark:text-sky-400 capitalize">
							{item.skill}
						</span>
					)}
				</div>

				<div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground md:gap-4">
					<span className="inline-flex items-center gap-1.5">
						<FileText className="size-3.5" aria-hidden="true" />
						{m["exams.ielts.category.questionsCount"]({
							count: item.questionsCount,
						})}
					</span>
					<span className="inline-flex items-center gap-1.5">
						<Clock className="size-3.5" aria-hidden="true" />
						{m["exams.ielts.category.durationMinutes"]({
							count: item.durationMinutes,
						})}
					</span>
				</div>
			</div>

			<div className="self-end sm:self-center shrink-0">
				<Link
					to={APP_ROUTES.EXAMS.IELTS.TEST}
					params={{ categoryId: item.categoryId, testId: item.id }}
					className="inline-flex items-center justify-center rounded-xl border border-border/80 bg-background px-4 py-2 text-xs font-bold text-foreground transition-colors hover:bg-muted"
				>
					{m["exams.ielts.category.viewDetail"]()}
				</Link>
			</div>
		</div>
	);
}
