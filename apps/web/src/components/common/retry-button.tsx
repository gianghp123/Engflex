import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { useState } from "react";
import { SubmitButton } from "#/components/common/submit-button";
import { m } from "#/paraglide/messages";

/**
 * Retry for error surfaces. `reset()` alone only clears the boundary UI — it
 * never re-runs loaders, so the stored match error re-throws on the next
 * render (verified in @tanstack/react-router's CatchBoundary). The working
 * retry is `router.invalidate()` (re-runs loaders and reloads matches) plus a
 * refetch of every errored query (covers routes with no loader). `reset` runs
 * last, after the awaits, so a re-render cannot race the refetch.
 *
 * `reset` is optional on purpose: the server render path passes
 * `reset: undefined`.
 */
export function RetryButton({ reset }: { reset?: () => void }) {
	const router = useRouter();
	const queryClient = useQueryClient();
	const [retrying, setRetrying] = useState(false);

	async function retry() {
		if (retrying) return;
		setRetrying(true);
		try {
			await Promise.allSettled([
				router.invalidate(),
				queryClient.refetchQueries({
					predicate: (query) => query.state.status === "error",
				}),
			]);
			reset?.();
		} finally {
			setRetrying(false);
		}
	}

	return (
		<SubmitButton
			type="button"
			variant="outline"
			pending={retrying}
			onClick={() => {
				void retry();
			}}
		>
			{m["common.actions.retry"]()}
		</SubmitButton>
	);
}
