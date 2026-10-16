import type { ActivityType } from "@engflex/contracts";
import type { ReactNode } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { BackButton } from "#/components/common/back-button";
import { CefrBadge } from "#/components/common/cefr-badge";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { useLessonDetail } from "#/features/lessons/queries";
import { m } from "#/paraglide/messages";
import { LessonStepNavigator } from "./lesson-step-navigator";

/** Stitch practice shell: step pills + level meta, no header card. */
export function PartShell({
	lessonId,
	part,
	onSelect,
	headerExtra,
	children,
}: {
	lessonId: string;
	part: ActivityType;
	onSelect: (part: ActivityType) => void;
	/** Rendered at the end of the step row — page-level help actions. */
	headerExtra?: ReactNode;
	children: ReactNode;
}) {
	const detailQuery = useLessonDetail(lessonId);
	const detail = detailQuery.data;
	if (!detail) return null;

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="heart" />,
				title: detail.title,
				description: detail.description,
			}}
			action={
				<BackButton
					to={APP_ROUTES.LESSONS.DETAIL}
					params={{ slug: detail.slug }}
					label={m["lessons.detail.backToLesson"]()}
				/>
			}
		>
			<div className="flex flex-wrap items-center justify-between gap-3">
				<LessonStepNavigator
					lessonId={lessonId}
					activities={detail.activities}
					currentPart={part}
					onSelect={onSelect}
				/>
				<span className="flex items-center gap-2">
					<span className="chip px-3 py-1 text-xs">
						<CefrBadge value={detail.cefrLevel} />
					</span>
					{headerExtra}
				</span>
			</div>
			{children}
		</PageLayout>
	);
}
