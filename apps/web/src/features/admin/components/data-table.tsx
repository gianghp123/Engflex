import type { Pagination } from "@engflex/contracts";
import { ChevronLeft, ChevronRight } from "lucide-react";
import type { ReactNode } from "react";
import { Button } from "#/components/ui/button";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "#/components/ui/table";
import { m } from "#/paraglide/messages";

export interface DataTableColumn<T> {
	id: string;
	header: ReactNode;
	cell: (row: T) => ReactNode;
	className?: string;
}

/**
 * Plain column-based grid for admin resource lists: rows + pagination only,
 * no sorting/filtering machinery — each admin page owns its own query state
 * and passes already-fetched rows in.
 */
export function DataTable<T>({
	columns,
	rows,
	rowKey,
	isLoading = false,
	pagination,
	onPageChange,
}: {
	columns: DataTableColumn<T>[];
	rows: T[];
	rowKey: (row: T) => string;
	isLoading?: boolean;
	pagination?: Pagination;
	onPageChange?: (page: number) => void;
}) {
	return (
		<div className="flex flex-col gap-3">
			<div className="rounded-md border">
				<Table>
					<TableHeader>
						<TableRow>
							{columns.map((column) => (
								<TableHead key={column.id} className={column.className}>
									{column.header}
								</TableHead>
							))}
						</TableRow>
					</TableHeader>
					<TableBody>
						{isLoading ? (
							<TableRow>
								<TableCell
									colSpan={columns.length}
									className="py-8 text-center text-muted-foreground"
								>
									{m["admin.table.loading"]()}
								</TableCell>
							</TableRow>
						) : rows.length === 0 ? (
							<TableRow>
								<TableCell
									colSpan={columns.length}
									className="py-8 text-center text-muted-foreground"
								>
									{m["admin.table.empty"]()}
								</TableCell>
							</TableRow>
						) : (
							rows.map((row) => (
								<TableRow key={rowKey(row)}>
									{columns.map((column) => (
										<TableCell key={column.id} className={column.className}>
											{column.cell(row)}
										</TableCell>
									))}
								</TableRow>
							))
						)}
					</TableBody>
				</Table>
			</div>
			{pagination && onPageChange && pagination.totalPages > 1 ? (
				<div className="flex items-center justify-end gap-2">
					<span className="text-sm text-muted-foreground">
						{pagination.page}/{pagination.totalPages}
					</span>
					<Button
						type="button"
						variant="outline"
						size="icon-sm"
						disabled={pagination.page <= 1}
						onClick={() => onPageChange(pagination.page - 1)}
					>
						<ChevronLeft />
					</Button>
					<Button
						type="button"
						variant="outline"
						size="icon-sm"
						disabled={pagination.page >= pagination.totalPages}
						onClick={() => onPageChange(pagination.page + 1)}
					>
						<ChevronRight />
					</Button>
				</div>
			) : null}
		</div>
	);
}
