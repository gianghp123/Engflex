import type { SubscriptionPlan, User } from "@engflex/contracts";
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
import { Field, FieldLabel } from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { Textarea } from "#/components/ui/textarea";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import { ResourceFormDialog } from "#/features/admin/components/resource-form-dialog";
import {
	useAdminUsers,
	useGrantSubscription,
	useRevokeSubscription,
	useUpdateUserRole,
} from "#/features/admin/users/queries";
import { m } from "#/paraglide/messages";

type GrantPlanChoice = SubscriptionPlan | "none";

function planLabel(plan: GrantPlanChoice): string {
	switch (plan) {
		case "1m":
			return m["payments.pricing.plan.1m"]();
		case "3m":
			return m["payments.pricing.plan.3m"]();
		case "12m":
			return m["payments.pricing.plan.12m"]();
		case "none":
			return m["admin.users.premium.planNone"]();
	}
}

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
	const grantMutation = useGrantSubscription();
	const revokeMutation = useRevokeSubscription();

	const [roleTarget, setRoleTarget] = useState<{
		user: User;
		nextRole: "admin" | "user";
	} | null>(null);

	const [grantTarget, setGrantTarget] = useState<User | null>(null);
	const [grantPlan, setGrantPlan] = useState<GrantPlanChoice>("1m");
	const [grantNote, setGrantNote] = useState("");

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

	function openGrantDialog(user: User) {
		setGrantTarget(user);
		setGrantPlan("1m");
		setGrantNote("");
	}

	function confirmGrant() {
		if (!grantTarget) return;
		const onSettled = {
			onSuccess: () => {
				toast.success(m["admin.toast.updated"]());
				setGrantTarget(null);
			},
			onError: (error: unknown) => {
				toast.error(
					error instanceof Error ? error.message : m["admin.toast.error"](),
				);
			},
		};
		if (grantPlan === "none") {
			revokeMutation.mutate(grantTarget.id, onSettled);
			return;
		}
		grantMutation.mutate(
			{
				userId: grantTarget.id,
				body: { plan: grantPlan, note: grantNote },
			},
			onSettled,
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
			id: "premium",
			header: m["admin.users.premium.column"](),
			cell: (row) => (
				<div className="flex flex-col gap-1">
					<Badge variant={row.premium ? "default" : "secondary"}>
						{row.premium
							? m["admin.users.premium.badgeActive"]()
							: m["admin.users.premium.badgeNone"]()}
					</Badge>
					{row.premium && row.premiumExpiresAt && (
						<span className="text-xs text-muted-foreground">
							{m["admin.users.premium.expiresOn"]({
								date: formatDate(row.premiumExpiresAt),
							})}
						</span>
					)}
				</div>
			),
		},
		{
			id: "actions",
			header: m["admin.table.actions"](),
			cell: (row) => (
				<div className="flex gap-2">
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
					<Button
						type="button"
						variant="outline"
						size="sm"
						onClick={() => openGrantDialog(row)}
					>
						{m["admin.users.premium.grantAction"]()}
					</Button>
				</div>
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
			<ResourceFormDialog
				open={Boolean(grantTarget)}
				onOpenChange={(open) => !open && setGrantTarget(null)}
				title={
					grantPlan === "none"
						? m["admin.users.premium.revokeTitle"]({
								email: grantTarget?.email ?? "",
							})
						: m["admin.users.premium.grantTitle"]({
								email: grantTarget?.email ?? "",
							})
				}
				description={
					grantPlan === "none"
						? m["admin.users.premium.revokeDescription"]()
						: m["admin.users.premium.grantDescription"]()
				}
				onSubmit={confirmGrant}
				pending={grantMutation.isPending || revokeMutation.isPending}
			>
				<Field>
					<FieldLabel>{m["admin.users.premium.planLabel"]()}</FieldLabel>
					<Select
						value={grantPlan}
						onValueChange={(value) => setGrantPlan(value as GrantPlanChoice)}
					>
						<SelectTrigger className="w-full">
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							{(["none", "1m", "3m", "12m"] as const).map((plan) => (
								<SelectItem key={plan} value={plan}>
									{planLabel(plan)}
								</SelectItem>
							))}
						</SelectContent>
					</Select>
				</Field>
				{grantPlan === "none" ? null : (
					<Field>
						<FieldLabel htmlFor="grant-note">
							{m["admin.users.premium.noteLabel"]()}
						</FieldLabel>
						<Textarea
							id="grant-note"
							value={grantNote}
							onChange={(event) => setGrantNote(event.target.value)}
						/>
					</Field>
				)}
			</ResourceFormDialog>
		</div>
	);
}
