import * as Sentry from "@sentry/tanstackstart-react";
import { useEffect } from "react";
import { ErrorComponent } from "#/components/common/error-component";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";

/** Global 404 (root notFoundComponent + thrown notFound() fallbacks). */
export function RouteNotFound() {
	return (
		<ErrorComponent
			badge={m["errors.notFound.badge"]()}
			mascot="confused"
			caption={m["errors.notFound.caption"]()}
			title={m["errors.notFound.title"]()}
			body={m["errors.notFound.body"]()}
			primaryLabel={m["errors.notFound.primary"]()}
		/>
	);
}

/** Route error boundary fallback (unexpected render/query failures). */
export function RouteErrorFallback({
	error,
	reset,
}: {
	error: unknown;
	reset: () => void;
}) {
	useEffect(() => {
		Sentry.captureException(error);
	}, [error]);

	return (
		<ErrorComponent
			badge={m["errors.server.badge"]()}
			mascot="sad"
			caption={m["errors.server.caption"]()}
			title={m["errors.server.title"]()}
			body={m["errors.server.body"]()}
			primaryLabel={m["common.actions.backHome"]()}
			extraAction={
				<Button type="button" variant="outline" onClick={() => reset()}>
					{m["common.actions.retry"]()}
				</Button>
			}
		/>
	);
}
