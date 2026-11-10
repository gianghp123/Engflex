import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { AdminExams } from "#/features/admin/exams/components/admin-exams";
import { parsePage } from "#/lib/router/search-params";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/admin/exams")({
	validateSearch: (
		search: Record<string, unknown>,
	): { categoryId?: string; page?: number } => ({
		categoryId:
			typeof search.categoryId === "string" ? search.categoryId : undefined,
		page: parsePage(search.page),
	}),
	component: AdminExamsPage,
});
function AdminExamsPage() {
	useBreadcrumbs([
		{ label: m["admin.nav.exams"](), to: APP_ROUTES.ADMIN.EXAMS },
	]);
	const { categoryId, page } = Route.useSearch();
	const navigate = Route.useNavigate();
	return (
		<AdminExams
			categoryId={categoryId}
			page={page ?? 1}
			onNavigate={(id, nextPage) => {
				void navigate({ search: { categoryId: id, page: nextPage } });
			}}
		/>
	);
}
