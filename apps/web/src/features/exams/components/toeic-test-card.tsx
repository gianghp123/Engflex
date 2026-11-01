import { ExamDurationMinutes } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { Clock, FileText } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { m } from "#/paraglide/messages";

export interface ToeicTestItem {
	id: string;
	categoryId: string;
	title: string;
	questionsCount: number;
}

export function ToeicTestCard({ item }: { item: ToeicTestItem }) {
	return (
		<div className="flex flex-col gap-3 rounded-2xl border border-border/80 bg-card p-4 shadow-sm transition-all hover:border-primary/40 hover:shadow-md sm:flex-row sm:items-center sm:justify-between md:p-5">
			<div className="flex min-w-0 flex-1 flex-col gap-2.5">
				<div className="flex flex-wrap items-center gap-2">
					<h3 className="truncate text-sm font-bold text-card-foreground md:text-base">
						{item.title}
					</h3>

					<span className="inline-flex items-center rounded bg-emerald-600 px-2 py-0.5 text-[11px] font-bold text-white">
						{m["exams.toeic.category.freeBadge"]()}
					</span>
				</div>

				<div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground md:gap-4">
					<span className="inline-flex items-center gap-1.5">
						<FileText className="size-3.5" aria-hidden="true" />
						{m["exams.toeic.category.questionsCount"]({
							count: item.questionsCount,
						})}
					</span>
					<span className="inline-flex items-center gap-1.5">
						<Clock className="size-3.5" aria-hidden="true" />
						{m["exams.toeic.category.fullExamDuration"]({
							count: ExamDurationMinutes,
						})}
					</span>
				</div>
			</div>

			<div className="self-end sm:self-center shrink-0">
				<Link
					to={APP_ROUTES.EXAMS.TOEIC.TEST}
					params={{ categoryId: item.categoryId, testId: item.id }}
					className="inline-flex items-center justify-center rounded-xl border border-border/80 bg-background px-4 py-2 text-xs font-bold text-foreground transition-colors hover:bg-muted"
				>
					{m["exams.toeic.category.viewDetail"]()}
				</Link>
			</div>
		</div>
	);
}
