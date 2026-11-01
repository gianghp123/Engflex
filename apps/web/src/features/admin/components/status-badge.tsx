import { Badge } from "#/components/ui/badge";
import { m } from "#/paraglide/messages";

type PublishStatus = "draft" | "published";

export function StatusBadge({ status }: { status: PublishStatus }) {
	return (
		<Badge variant={status === "published" ? "default" : "secondary"}>
			{status === "published"
				? m["admin.status.published"]()
				: m["admin.status.draft"]()}
		</Badge>
	);
}
