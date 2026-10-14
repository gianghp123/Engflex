import { Link } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { MoMascot } from "#/components/common/mo-mascot";
import { Button } from "#/components/ui/button";
import {
	Empty,
	EmptyContent,
	EmptyDescription,
	EmptyHeader,
	EmptyTitle,
} from "#/components/ui/empty";
import { m } from "#/paraglide/messages";

/**
 * Production placeholder for DEV-only preview routes: renders in place
 * (staying inside AppShell) instead of redirecting or 404ing, so a dev
 * page never reads as a broken app.
 */
export function InDevelopment() {
	return (
		<Empty className="items-center text-center">
			<MoMascot variant="thinking" size={72} />
			<EmptyHeader>
				<EmptyTitle>{m["common.dev.title"]()}</EmptyTitle>
				<EmptyDescription>{m["common.dev.body"]()}</EmptyDescription>
			</EmptyHeader>
			<EmptyContent>
				<Button asChild className="btn btn-primary">
					<Link to={APP_ROUTES.HOME}>{m["common.actions.backHome"]()}</Link>
				</Button>
			</EmptyContent>
		</Empty>
	);
}
