import { Link } from "@tanstack/react-router";
import { Sparkles } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { useEntitlements } from "#/features/payments/entitlements";
import { m } from "#/paraglide/messages";

/** Header credit meter: remaining AI credits, linking to the usage page.
 *  Renders nothing until the entitlements load (or when signed out, where
 *  the query 401s and stays empty). */
export function CreditsPill() {
	const { data } = useEntitlements();
	if (!data) return null;
	const label = m["payments.entitlements.pill"]({
		remaining: data.credits.remaining,
	});
	return (
		<Link
			to={APP_ROUTES.PAYMENTS.USAGE}
			aria-label={label}
			className="flex items-center gap-1 rounded-full border bg-card px-2.5 py-1 text-xs shadow-xs hover:bg-accent"
		>
			<Sparkles aria-hidden className="size-3.5 text-muted-foreground" />
			<span className="tabular-nums">{label}</span>
		</Link>
	);
}
