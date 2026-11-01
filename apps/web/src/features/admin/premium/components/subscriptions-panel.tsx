import type {
	Subscription,
	SubscriptionPlan,
	SubscriptionStatus,
} from "@engflex/contracts";
import { useState } from "react";
import { Badge } from "#/components/ui/badge";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import { useAdminSubscriptions } from "#/features/admin/premium/queries";
import { m } from "#/paraglide/messages";

const STATUS_OPTIONS: SubscriptionStatus[] = [
	"pending",
	"active",
	"expired",
	"cancelled",
	"failed",
];
const PLAN_OPTIONS: SubscriptionPlan[] = ["1m", "3m", "12m"];

function formatDate(iso: string | undefined | null): string {
	if (!iso) return "—";
	return new Date(iso).toLocaleDateString();
}

function statusVariant(status: SubscriptionStatus) {
	if (status === "active") return "default" as const;
	if (status === "failed" || status === "cancelled")
		return "destructive" as const;
	return "secondary" as const;
}

export function SubscriptionsPanel() {
	const [page, setPage] = useState(1);
	const [status, setStatus] = useState<SubscriptionStatus | "">("");
	const [plan, setPlan] = useState<SubscriptionPlan | "">("");
	const pageSize = 20;
	const query = useAdminSubscriptions(page, pageSize, status, plan);

	const columns: DataTableColumn<Subscription>[] = [
		{
			id: "userId",
			header: m["admin.premium.subscriptions.user"](),
			cell: (row) => row.userId,
		},
		{
			id: "plan",
			header: m["admin.premium.plans.plan"](),
			cell: (row) => row.plan,
		},
		{
			id: "status",
			header: m["admin.table.status"](),
			cell: (row) => (
				<Badge variant={statusVariant(row.status)}>{row.status}</Badge>
			),
		},
		{
			id: "expiresAt",
			header: m["admin.premium.subscriptions.expires"](),
			cell: (row) => formatDate(row.expiresAt),
		},
		{
			id: "createdAt",
			header: m["admin.premium.subscriptions.created"](),
			cell: (row) => formatDate(row.createdAt),
		},
	];

	return (
		<div className="flex flex-col gap-4">
			<div className="flex gap-2">
				<Select
					value={status || "all"}
					onValueChange={(value) => {
						setStatus(value === "all" ? "" : (value as SubscriptionStatus));
						setPage(1);
					}}
				>
					<SelectTrigger className="w-48">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="all">
							{m["admin.premium.subscriptions.allStatuses"]()}
						</SelectItem>
						{STATUS_OPTIONS.map((option) => (
							<SelectItem key={option} value={option}>
								{option}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
				<Select
					value={plan || "all"}
					onValueChange={(value) => {
						setPlan(value === "all" ? "" : (value as SubscriptionPlan));
						setPage(1);
					}}
				>
					<SelectTrigger className="w-48">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="all">
							{m["admin.premium.subscriptions.allPlans"]()}
						</SelectItem>
						{PLAN_OPTIONS.map((option) => (
							<SelectItem key={option} value={option}>
								{option}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
			</div>
			<DataTable
				columns={columns}
				rows={query.data?.items ?? []}
				rowKey={(row) => row.id}
				isLoading={query.isPending}
				pagination={query.data?.pagination}
				onPageChange={setPage}
			/>
		</div>
	);
}
