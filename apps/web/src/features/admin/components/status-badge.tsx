import { CircleCheck, CircleDashed } from "lucide-react";
import { Badge } from "#/components/ui/badge";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#/components/ui/tooltip";
import { m } from "#/paraglide/messages";

type PublishStatus = "draft" | "published";

export function StatusBadge({ status }: { status: PublishStatus }) {
	const label =
		status === "published"
			? m["admin.status.published"]()
			: m["admin.status.draft"]();
	return (
		<Tooltip>
			<TooltipTrigger asChild>
				<Badge
					variant={status === "published" ? "default" : "secondary"}
					aria-label={label}
				>
					{status === "published" ? <CircleCheck /> : <CircleDashed />}
				</Badge>
			</TooltipTrigger>
			<TooltipContent>{label}</TooltipContent>
		</Tooltip>
	);
}
