import * as Sentry from "@sentry/tanstackstart-react";
import { Link } from "@tanstack/react-router";
import { useEffect } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { ErrorComponent } from "#/components/common/error-component";
import { RetryButton } from "#/components/common/retry-button";
import { Button } from "#/components/ui/button";
import { ApiError } from "#/lib/api";
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

/** Route error boundary fallback (unexpected render/query failures). A 403
 *  from the API is never a crash — it's a learner hitting premium-gated
 *  content — so it gets its own upgrade-prompt copy instead of the generic
 *  retry page, and is never reported to Sentry as an exception. */
export function RouteErrorFallback({
	error,
	reset,
}: {
	error: unknown;
	reset: () => void;
}) {
	const isPremiumRequired = error instanceof ApiError && error.status === 403;

	useEffect(() => {
		if (!isPremiumRequired) {
			Sentry.captureException(error);
		}
	}, [error, isPremiumRequired]);

	if (isPremiumRequired) {
		return (
			<ErrorComponent
				badge={m["errors.premiumRequired.badge"]()}
				mascot="heart"
				caption={m["errors.premiumRequired.caption"]()}
				title={m["errors.premiumRequired.title"]()}
				body={m["errors.premiumRequired.body"]()}
				primaryLabel={m["errors.premiumRequired.primary"]()}
				primaryTo={APP_ROUTES.PAYMENTS.PRICING}
				extraAction={
					<Button asChild type="button" variant="outline">
						<Link to={APP_ROUTES.HOME}>{m["common.actions.backHome"]()}</Link>
					</Button>
				}
			/>
		);
	}

	return (
		<ErrorComponent
			badge={m["errors.server.badge"]()}
			mascot="sad"
			caption={m["errors.server.caption"]()}
			title={m["errors.server.title"]()}
			body={m["errors.server.body"]()}
			primaryLabel={m["common.actions.backHome"]()}
			extraAction={<RetryButton reset={reset} />}
		/>
	);
}
