import type { Activity } from "@engflex/contracts";
import { SubmitButton } from "#/components/common/submit-button";
import { m } from "#/paraglide/messages";

export function ActivityRow({
	activity,
	onStart,
	startPending,
}: {
	activity: Activity;
	onStart: () => void;
	startPending: boolean;
}) {
	return (
		<div className="surface-card flex flex-col gap-3 p-5 sm:flex-row sm:items-center sm:gap-4">
			<span className="inline-flex size-11 shrink-0 items-center justify-center rounded-full bg-primary text-lg font-bold text-primary-foreground shadow-edge-primary-sm">
				{activity.partNumber}
			</span>
			<div className="flex min-w-0 flex-1 flex-col gap-1">
				<div className="flex flex-wrap items-center gap-2">
					<h3 className="text-base font-bold text-foreground">
						{activity.title}
					</h3>
					<span className="chip bg-accent px-2 py-0 text-[11px]">
						{m["lessons.card.activityDuration"]({
							minutes: activity.durationMin,
						})}
					</span>
				</div>
				<p className="text-[15px] font-medium text-muted-foreground">
					{activity.description}
				</p>
			</div>
			<SubmitButton
				type="button"
				variant="outline"
				size="sm"
				className="btn btn-outline sm:ml-auto"
				onClick={onStart}
				pending={startPending}
			>
				{m["lessons.card.start"]()}
			</SubmitButton>
		</div>
	);
}
