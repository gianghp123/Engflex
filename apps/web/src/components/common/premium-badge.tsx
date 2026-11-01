import { cn } from "cn";
import { Badge } from "#/components/ui/badge";
import { m } from "#/paraglide/messages";

/** Marks admin-flagged premium content on a learner-facing card. Text
 *  only, no lock icon — the card's own action decides what clicking does
 *  (open normally vs. redirect to pricing), this just labels the content. */
export function PremiumBadge({ className }: { className?: string }) {
	return (
		<Badge
			variant="default"
			className={cn("bg-accent-violet text-white", className)}
		>
			{m["common.premium.badge"]()}
		</Badge>
	);
}
