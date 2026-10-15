import { enUS } from "@clerk/localizations/en-US";
import { viVN } from "@clerk/localizations/vi-VN";
import { ClerkProvider, useAuth } from "@clerk/tanstack-react-start";
import { useEffect } from "react";
import { registerTokenGetter } from "#/lib/auth-token";
import { getLocale } from "#/paraglide/runtime";

/** Clerk's component copy follows the app locale (vi base + en). */
const CLERK_LOCALIZATION = { en: enUS, vi: viVN } as const;

export default function AppClerkProvider({
	children,
}: {
	children: React.ReactNode;
}) {
	return (
		<ClerkProvider localization={CLERK_LOCALIZATION[getLocale()]}>
			<TokenBridge />
			{children}
		</ClerkProvider>
	);
}

/**
 * Bridges Clerk's session token to non-component call sites (`lib/api.ts`).
 * Clerk's documented external-API pattern: `await getToken()` → Bearer token.
 */
function TokenBridge() {
	const { getToken } = useAuth();

	useEffect(() => {
		registerTokenGetter(() => getToken());
		return () => registerTokenGetter(null);
	}, [getToken]);

	return null;
}
