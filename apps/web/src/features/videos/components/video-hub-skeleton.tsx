import { Skeleton } from "#/components/ui/skeleton";
import { m } from "#/paraglide/messages";

export function VideoHubSkeleton() {
	return (
		<output
			aria-label={m["videos.hub.loading"]()}
			className="flex flex-col gap-8"
		>
			{[0, 1].map((row) => (
				<div key={row} className="flex flex-col gap-4">
					<Skeleton className="h-14 rounded-[24px]" />
					<div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
						{[0, 1, 2, 3].map((card) => (
							<Skeleton key={card} className="aspect-[4/3] rounded-[24px]" />
						))}
					</div>
				</div>
			))}
		</output>
	);
}
