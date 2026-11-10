import { createFileRoute } from "@tanstack/react-router";
import { AdminOverview } from "#/features/admin/components/overview-dashboard";

export const Route = createFileRoute("/admin/")({
	component: AdminOverview,
});
