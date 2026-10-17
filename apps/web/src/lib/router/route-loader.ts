import type { QueryClient } from "@tanstack/react-query";
import { notFound, redirect } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { ApiError } from "../api/api";

/**
 * Preload a query in a route `loader` with correct 404/error routing.
 * `throw notFound()` is only honored in loaders (and beforeLoad) — thrown
 * during render it escapes down the error-boundary path instead of rendering
 * the route's notFoundComponent, which is how an API outage once masqueraded
 * as a crash page. So: fetch here, render from the warm cache in components.
 *
 * A 404 from the API becomes notFound(); anything else rethrows to the
 * route's errorComponent (retryable). Pair with the same query options
 * object the component's useQuery uses so loader and component share one
 * cache key.
 *
 * A 401 becomes a sign-in redirect instead of a crash. Every caller lives
 * behind the `_app` auth guard, so a 401 is token plumbing — an SSR load
 * with no Bearer, or Clerk still hydrating — never a signed-out user. The
 * sign-in page bounces an active session straight back with fresh
 * credentials, which heals both cases with no error page.
 *
 * Retries exclude control flow: 404 settles instantly to notFound, 401 to
 * the redirect; real failures (timeouts, 500s, down API) retry with the
 * default backoff (~7s worst case for 3) before the retryable error page.
 */
export async function loadOr404<T, TKey extends readonly unknown[]>(
	queryClient: QueryClient,
	options: { queryKey: TKey; queryFn: () => Promise<T> },
): Promise<T> {
	try {
		return await queryClient.query({
			...options,
			retry: (failureCount: number, error: unknown) =>
				!(
					error instanceof ApiError &&
					(error.status === 404 || error.status === 401)
				) && failureCount < 3,
		});
	} catch (error) {
		if (error instanceof ApiError && error.status === 404) {
			throw notFound();
		}
		if (error instanceof ApiError && error.status === 401) {
			throw redirect({ href: APP_ROUTES.AUTH.SIGN_IN });
		}
		throw error;
	}
}
