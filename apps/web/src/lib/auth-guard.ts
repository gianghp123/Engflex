import { auth } from "@clerk/tanstack-react-start/server";
import { createServerFn } from "@tanstack/react-start";
import { isAdminClaims } from "#/lib/roles";

/** Lets a local dev session reach /admin without setting up a Clerk admin
 *  role. `import.meta.env.DEV` keeps it out of production builds even if
 *  the env var leaks into a deployed environment. Scoped to `isAdmin` only
 *  — `isAuthenticated` still reflects the real Clerk session, so the admin
 *  route guard's sign-in redirect is unaffected. */
const ADMIN_DEV_BYPASS =
	import.meta.env.DEV && import.meta.env.VITE_ADMIN_DEV_BYPASS === "true";

/**
 * Server-side auth probe for route `beforeLoad` guards.
 * Clerk's `<Show>` only controls what renders (SSR would return 200 and could
 * flash blank); probing `auth()` lets the guard issue a real redirect.
 */
export const getAuthState = createServerFn().handler(async () => {
	const { isAuthenticated, sessionClaims } = await auth();
	return {
		isAuthenticated,
		isAdmin: ADMIN_DEV_BYPASS || isAdminClaims(sessionClaims),
	};
});
