import type {
	AttemptStatus,
	AttemptType,
	User,
	UserAttempt,
} from "@engflex/contracts";
import { useState } from "react";
import { Badge } from "#/components/ui/badge";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
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
import { useAdminUserAttempts } from "#/features/admin/users/queries";
import { m } from "#/paraglide/messages";

interface UserAttemptsDialogProps {
	user: User | null;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

function formatDate(iso: string | undefined): string {
	if (!iso) return "—";
	const d = new Date(iso);
	return `${d.toLocaleDateString()} ${d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`;
}

function formatDuration(sec: number | undefined): string {
	if (sec == null) return "—";
	if (sec < 60) return `${sec}s`;
	const mins = Math.floor(sec / 60);
	const remainSec = sec % 60;
	return `${mins}m ${remainSec}s`;
}

function formatScore(score: number | undefined): string {
	if (score == null) return "—";
	return Number.isInteger(score) ? score.toString() : score.toFixed(1);
}

function typeLabel(type: AttemptType): string {
	switch (type) {
		case "lesson":
			return m["admin.users.attempts.typeLesson"]();
		case "exam":
			return m["admin.users.attempts.typeExam"]();
		case "dictation":
			return m["admin.users.attempts.typeDictation"]();
		case "shadowing":
			return m["admin.users.attempts.typeShadowing"]();
		default:
			return m["admin.users.attempts.typeOther"]();
	}
}

function statusBadge(status: AttemptStatus) {
	switch (status) {
		case "completed":
			return (
				<Badge variant="default">
					{m["admin.users.attempts.statusCompleted"]()}
				</Badge>
			);
		case "in_progress":
			return (
				<Badge variant="secondary">
					{m["admin.users.attempts.statusInProgress"]()}
				</Badge>
			);
		case "abandoned":
			return (
				<Badge variant="destructive">
					{m["admin.users.attempts.statusAbandoned"]()}
				</Badge>
			);
		default:
			return <Badge variant="outline">{status}</Badge>;
	}
}

function resolveTitle(attempt: UserAttempt): string {
	return (
		attempt.lesson?.title ??
		attempt.examTest?.title ??
		attempt.videoExercise?.title ??
		"—"
	);
}

export function UserAttemptsDialog({
	user,
	open,
	onOpenChange,
}: UserAttemptsDialogProps) {
	const [page, setPage] = useState(1);
	const [typeFilter, setTypeFilter] = useState<string>("all");
	const pageSize = 10;

	const filterType =
		typeFilter === "all" ? undefined : (typeFilter as AttemptType);

	const query = useAdminUserAttempts(
		user?.id ?? "",
		page,
		pageSize,
		filterType,
	);

	const userName =
		[user?.firstName, user?.lastName].filter(Boolean).join(" ") ||
		user?.email ||
		"";

	const columns: DataTableColumn<UserAttempt>[] = [
		{
			id: "type",
			header: m["admin.users.attempts.columnType"](),
			cell: (row) => (
				<Badge variant="outline" className="font-normal">
					{typeLabel(row.type)}
				</Badge>
			),
		},
		{
			id: "title",
			header: m["admin.users.attempts.columnTitle"](),
			cell: (row) => (
				<span className="font-medium text-foreground">{resolveTitle(row)}</span>
			),
		},
		{
			id: "status",
			header: m["admin.users.attempts.columnStatus"](),
			cell: (row) => statusBadge(row.status),
		},
		{
			id: "score",
			header: m["admin.users.attempts.columnScore"](),
			cell: (row) => formatScore(row.score),
		},
		{
			id: "duration",
			header: m["admin.users.attempts.columnDuration"](),
			cell: (row) => formatDuration(row.durationSec),
		},
		{
			id: "date",
			header: m["admin.users.attempts.columnDate"](),
			cell: (row) => formatDate(row.submittedAt ?? row.createdAt),
		},
	];

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="max-w-4xl max-h-[85vh] overflow-y-auto">
				<DialogHeader>
					<DialogTitle>
						{m["admin.users.attempts.dialogTitle"]({ name: userName })}
					</DialogTitle>
					<DialogDescription>
						{m["admin.users.attempts.dialogDescription"]()}
					</DialogDescription>
				</DialogHeader>

				<div className="flex flex-col gap-4 py-2">
					<div className="flex items-center gap-3">
						<span className="text-sm font-medium text-muted-foreground">
							{m["admin.users.attempts.typeFilter"]()}:
						</span>
						<Select
							value={typeFilter}
							onValueChange={(val) => {
								setTypeFilter(val);
								setPage(1);
							}}
						>
							<SelectTrigger className="w-48">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="all">
									{m["admin.users.attempts.allTypes"]()}
								</SelectItem>
								<SelectItem value="lesson">
									{m["admin.users.attempts.typeLesson"]()}
								</SelectItem>
								<SelectItem value="exam">
									{m["admin.users.attempts.typeExam"]()}
								</SelectItem>
								<SelectItem value="dictation">
									{m["admin.users.attempts.typeDictation"]()}
								</SelectItem>
								<SelectItem value="shadowing">
									{m["admin.users.attempts.typeShadowing"]()}
								</SelectItem>
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
			</DialogContent>
		</Dialog>
	);
}
