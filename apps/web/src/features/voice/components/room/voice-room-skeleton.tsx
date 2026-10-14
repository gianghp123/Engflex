import { Skeleton } from "#/components/ui/skeleton";

/**
 * Loading placeholder for the voice room (and its DEV preview). Mirrors the
 * dual-panel layout so the client snapping in doesn't shift content.
 */
export function VoiceRoomSkeleton() {
	return (
		<div className="flex min-h-0 flex-1 flex-col gap-4 p-4">
			<div className="grid min-h-0 flex-1 items-stretch gap-4 lg:grid-cols-2">
				<div className="surface-card flex h-full flex-col gap-6 py-6">
					<div className="flex flex-col gap-2 px-6">
						<Skeleton className="h-6 w-40" />
						<Skeleton className="h-4 w-64" />
					</div>
					<div className="flex flex-1 flex-col items-center gap-6 px-6 pb-6">
						<div className="flex flex-1 items-center justify-center">
							<Skeleton className="size-40 rounded-full" />
						</div>
						<Skeleton className="h-4 w-48" />
						<div className="flex gap-3">
							<Skeleton className="size-12 rounded-full" />
							<Skeleton className="size-12 rounded-full" />
						</div>
					</div>
				</div>
				<div className="surface-card hidden h-full flex-col gap-6 py-6 lg:flex">
					<div className="flex flex-col gap-2 px-6">
						<Skeleton className="h-6 w-32" />
					</div>
					<div className="flex flex-1 flex-col gap-4 px-6">
						<Skeleton className="h-14 w-3/4" />
						<Skeleton className="h-14 w-2/3 self-end" />
						<Skeleton className="h-14 w-3/4" />
					</div>
				</div>
			</div>
		</div>
	);
}
