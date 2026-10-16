import { Link } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";
import { Button } from "#/components/ui/button";

/**
 * The app's one back button: chunky outline pill with a leading arrow — the
 * drill-down pages' way back. Always a router `<Link>`: pass the
 * `APP_ROUTES` target and the page's own label; the look lives here so it
 * cannot drift page to page.
 */
export function BackButton({
	to,
	params,
	label,
}: {
	to: string;
	params?: Record<string, string>;
	label: string;
}) {
	return (
		<Button asChild variant="outline" className="btn btn-outline">
			<Link to={to} params={params}>
				<ArrowLeft data-icon="inline-start" />
				{label}
			</Link>
		</Button>
	);
}
