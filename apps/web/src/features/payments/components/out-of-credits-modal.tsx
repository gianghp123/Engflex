import { Link } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import type { ApiError } from "#/lib/api";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";

function resetLabel(error: ApiError | null): string {
	const raw = error?.data?.resetAt;
	if (typeof raw !== "string")
		return m["payments.entitlements.modal.noReset"]();
	const date = new Date(raw);
	if (Number.isNaN(date.getTime()))
		return m["payments.entitlements.modal.noReset"]();
	return date.toLocaleDateString(getLocale());
}

/** Quota-denial dialog for a paid-AI mutation: explains the reset date and
 *  routes to pricing. No top-ups at launch — upgrade is the only action. */
export function OutOfCreditsModal({
	error,
	onClose,
}: {
	error: ApiError | null;
	onClose: () => void;
}) {
	return (
		<Dialog open={error !== null} onOpenChange={(open) => !open && onClose()}>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>{m["payments.entitlements.modal.title"]()}</DialogTitle>
					<DialogDescription>
						{m["payments.entitlements.modal.body"]({
							reset: resetLabel(error),
						})}
					</DialogDescription>
				</DialogHeader>
				<DialogFooter>
					<Button type="button" variant="outline" onClick={onClose}>
						{m["payments.entitlements.modal.later"]()}
					</Button>
					<Button type="button" asChild>
						<Link to={APP_ROUTES.PAYMENTS.PRICING}>
							{m["payments.entitlements.modal.upgrade"]()}
						</Link>
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
