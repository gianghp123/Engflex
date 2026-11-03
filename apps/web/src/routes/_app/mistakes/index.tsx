import { ERROR_TYPE_VALUES, type ErrorType } from "@engflex/contracts";
import { createFileRoute } from "@tanstack/react-router";
import { ArrowRight } from "lucide-react";
import { useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { CardShell } from "#/components/common/card-shell";
import { EmptyList } from "#/components/common/empty-list";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { Badge } from "#/components/ui/badge";
import { Skeleton } from "#/components/ui/skeleton";
import { useMistakes } from "#/features/progress/queries";
import type { Mistake } from "#/features/progress/service";
import { enumOptions } from "#/lib/enum-options";
import { cn } from "#/lib/utils";
import { m } from "#/paraglide/messages";

const MISTAKES_LIMIT = 20;

type MistakesFilter = ErrorType | "all";

/** House answer to dynamic copy: filter chips render the code itself, so no
 *  prose record is needed — `enumOptions` without labels. */
const errorTypeOptions = enumOptions(ERROR_TYPE_VALUES);

const SOURCE_COPY = {
	feedback: () => m["mistakes.row.fromFeedback"](),
	writing: () => m["mistakes.row.fromWriting"](),
} as const;

function MistakeRow({ mistake }: { mistake: Mistake }) {
	return (
		<li className="flex flex-col gap-1 px-4 py-3">
			<div className="flex items-center justify-between gap-2">
				<Badge variant="outline" className="shrink-0">
					{mistake.errorType}
				</Badge>
				<span className="truncate text-xs font-medium text-muted-foreground">
					{SOURCE_COPY[mistake.source]()}
				</span>
			</div>
			<p className="flex flex-wrap items-center gap-2 text-[15px] font-medium text-foreground">
				<span className="text-ai-coral line-through opacity-85">
					{mistake.text}
				</span>
				<ArrowRight size={16} className="shrink-0 text-muted-foreground" />
				<span className="font-semibold text-accuracy">
					{mistake.correction}
				</span>
			</p>
			{mistake.reason ? (
				<p className="text-xs font-normal text-muted-foreground">
					{mistake.reason}
				</p>
			) : null}
		</li>
	);
}

function MistakeRowSkeleton() {
	return (
		<div className="flex flex-col gap-2 px-4 py-3">
			<Skeleton className="h-6 w-24 rounded-full" />
			<Skeleton className="h-4 w-3/4" />
			<Skeleton className="h-3 w-2/3" />
		</div>
	);
}

function MistakesPage() {
	useBreadcrumbs([
		{ label: m["nav.item.home"](), to: APP_ROUTES.HOME },
		{ label: m["mistakes.crumb.hub"]() },
	]);
	const [filter, setFilter] = useState<MistakesFilter>("all");
	const mistakes = useMistakes({ limit: MISTAKES_LIMIT, errorType: filter });
	const rows = mistakes.data ?? [];

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="thinking" size={84} />,
				title: m["mistakes.hub.title"](),
				description: m["mistakes.hub.subtitle"](),
			}}
			action={<BackButton to={APP_ROUTES.HOME} label={m["nav.item.home"]()} />}
		>
			<div className="flex flex-wrap items-center gap-2">
				<span className="text-[11px] font-semibold text-muted-foreground">
					{m["mistakes.filter.label"]()}
				</span>
				<button
					type="button"
					onClick={() => setFilter("all")}
					className={cn(
						"chip px-2.5 py-1 text-[11px] transition",
						filter === "all" ? "chip-selected" : "hover:border-primary",
					)}
				>
					{m["mistakes.filter.all"]()}
				</button>
				{errorTypeOptions.map((option) => (
					<button
						key={option.value}
						type="button"
						onClick={() => setFilter(option.value)}
						className={cn(
							"chip px-2.5 py-1 text-[11px] transition",
							filter === option.value
								? "chip-selected"
								: "hover:border-primary",
						)}
					>
						{option.label()}
					</button>
				))}
			</div>

			{mistakes.isPending ? (
				<CardShell
					pending
					skeleton={
						<>
							<MistakeRowSkeleton />
							<MistakeRowSkeleton />
							<MistakeRowSkeleton />
						</>
					}
					className="divide-y overflow-hidden"
				/>
			) : mistakes.isError ? (
				<div className="surface-card flex flex-col items-center gap-3 p-8 text-center">
					<MoMascot variant="confused" size={72} />
					<p className="max-w-md text-sm font-medium text-muted-foreground">
						{m["mistakes.hub.loadingFailed"]()}
					</p>
				</div>
			) : rows.length === 0 ? (
				<EmptyList
					title={m["mistakes.hub.emptyTitle"]()}
					description={m["mistakes.hub.emptyBody"]()}
					onReset={() => setFilter("all")}
				/>
			) : (
				<ul className="surface-card divide-y overflow-hidden">
					{rows.map((mistake) => (
						<MistakeRow
							key={`${mistake.errorType}-${mistake.text}-${mistake.createdAt}`}
							mistake={mistake}
						/>
					))}
				</ul>
			)}
		</PageLayout>
	);
}

export const Route = createFileRoute("/_app/mistakes/")({
	component: MistakesPage,
});
