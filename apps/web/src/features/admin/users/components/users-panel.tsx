import type { User } from "@engflex/contracts";
import { useState } from "react";
import { toast } from "sonner";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "#/components/ui/alert-dialog";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import {
	useAdminUsers,
	useUpdateUserRole,
} from "#/features/admin/users/queries";
import { m } from "#/paraglide/messages";

function formatDate(iso: string | undefined): string {
	if (!iso) return "—";
	return new Date(iso).toLocaleDateString();
}

export function UsersPanel() {
	const [page, setPage] = useState(1);
	const [search, setSearch] = useState("");
	const pageSize = 20;
	const query = useAdminUsers(page, pageSize, search);
	const updateRoleMutation = useUpdateUserRole();

	const [roleTarget, setRoleTarget] = useState<{
		user: User;
		nextRole: "admin" | "user";
	} | null>(null);

	function confirmRoleChange() {
		if (!roleTarget) return;
		updateRoleMutation.mutate(
			{ id: roleTarget.user.id, role: roleTarget.nextRole },
			{
				onSuccess: () => {
					toast.success(m["admin.toast.updated"]());
					setRoleTarget(null);
				},
				onError: (error) => {
					toast.error(
						error instanceof Error ? error.message : m["admin.toast.error"](),
					);
				},
			},
		);
	}

	const columns: DataTableColumn<User>[] = [
		{ id: "email", header: "Email", cell: (row) => row.email || "—" },
		{
			id: "name",
			header: "Name",
			cell: (row) =>
				[row.firstName, row.lastName].filter(Boolean).join(" ") || "—",
		},
		{
			id: "role",
			header: "Role",
			cell: (row) => (
				<Badge variant={row.role === "admin" ? "default" : "secondary"}>
					{row.role}
				</Badge>
			),
		},
		{
			id: "created",
			header: "Created",
			cell: (row) => formatDate(row.createdAt),
		},
		{
			id: "lastActive",
			header: "Last active",
			cell: (row) => formatDate(row.lastActiveAt),
		},
		{
			id: "actions",
			header: m["admin.table.actions"](),
			cell: (row) => (
				<Button
					type="button"
					variant="outline"
					size="sm"
					onClick={() =>
						setRoleTarget({
							user: row,
							nextRole: row.role === "admin" ? "user" : "admin",
						})
					}
				>
					{row.role === "admin" ? "Remove admin" : "Make admin"}
				</Button>
			),
		},
	];

	return (
		<div className="flex flex-col gap-4">
			<Input
				placeholder="Search by name, email or username"
				value={search}
				onChange={(event) => {
					setSearch(event.target.value);
					setPage(1);
				}}
				className="max-w-sm"
			/>
			<DataTable
				columns={columns}
				rows={query.data?.items ?? []}
				rowKey={(row) => row.id}
				isLoading={query.isPending}
				pagination={query.data?.pagination}
				onPageChange={setPage}
			/>
			<AlertDialog
				open={Boolean(roleTarget)}
				onOpenChange={(open) => !open && setRoleTarget(null)}
			>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>
							{roleTarget?.nextRole === "admin"
								? "Grant admin access?"
								: "Remove admin access?"}
						</AlertDialogTitle>
						<AlertDialogDescription>
							{roleTarget?.user.email} will become {roleTarget?.nextRole}.
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>{m["admin.actions.cancel"]()}</AlertDialogCancel>
						<AlertDialogAction
							disabled={updateRoleMutation.isPending}
							onClick={confirmRoleChange}
						>
							{m["admin.actions.save"]()}
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</div>
	);
}
