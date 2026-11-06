import { useAuth } from "@clerk/tanstack-react-start";
import { useQuery } from "@tanstack/react-query";
import { useStore } from "@tanstack/react-store";
import { useState } from "react";
import { CardShell } from "#/components/common/card-shell";
import { Pager } from "#/components/common/pager";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { FieldError } from "#/components/ui/field";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "#/components/ui/tabs";
import { ConfirmDeleteDialog } from "#/features/admin/components/confirm-delete-dialog";
import { DataTable } from "#/features/admin/components/data-table";
import { ResourceFormDialog } from "#/features/admin/components/resource-form-dialog";
import {
	examCategoriesQueryOptions,
	examTestsQueryOptions,
} from "#/features/exams/queries";
import { ApiError } from "#/lib/api";
import { m } from "#/paraglide/messages";
import {
	type CategoryDraft,
	createExamsStore,
	type ExamDraft,
	newExamDraft,
} from "../store";
import { AuthorField, ExamEditor } from "./exam-editor";

export function AdminExams({
	categoryId,
	page,
	onNavigate,
}: {
	categoryId?: string;
	page: number;
	onNavigate: (categoryId: string, page: number) => void;
}) {
	const { isLoaded } = useAuth();
	const retry = (count: number, error: unknown) =>
		!(error instanceof ApiError && [401, 403, 404].includes(error.status)) &&
		count < 3;
	const catalog = useQuery({
		...examCategoriesQueryOptions({ type: "toeic" }),
		enabled: isLoaded,
		retry,
	});
	const tests = useQuery({
		...examTestsQueryOptions(categoryId ?? "", { page, pageSize: 20 }),
		enabled: isLoaded && !!categoryId,
		retry,
	});
	const [store] = useState(createExamsStore);
	const categories = useStore(store);
	const [selectedId, setSelectedId] = useState("demo-category");
	const category = categories.find((row) => row.id === selectedId);
	const [categoryForm, setCategoryForm] = useState<CategoryDraft | null>(null);
	const [testForm, setTestForm] = useState<ExamDraft | null>(null);
	const [error, setError] = useState(false);
	const [remove, setRemove] = useState<{
		categoryId?: string;
		testId?: string;
		reset?: boolean;
	} | null>(null);
	function saveTest(test: ExamDraft) {
		if (
			!category ||
			categories.some((c) =>
				c.tests.some(
					(old) =>
						old.id !== test.id &&
						(old.slug === test.slug || old.sourceKey === test.sourceKey),
				),
			)
		) {
			setError(true);
			return;
		}
		store.setState((rows) =>
			rows.map((row) =>
				row.id === category.id
					? {
							...row,
							tests: row.tests.some((old) => old.id === test.id)
								? row.tests.map((old) => (old.id === test.id ? test : old))
								: [...row.tests, test],
						}
					: row,
			),
		);
		setTestForm(null);
		setError(false);
	}
	return (
		<div className="container-content flex flex-col gap-5 py-8">
			<h1 className="text-2xl font-bold">{m["admin.exams.page.title"]()}</h1>
			<Tabs defaultValue="catalog">
				<TabsList>
					<TabsTrigger value="catalog">
						{m["admin.exams.page.catalog"]()}
					</TabsTrigger>
					<TabsTrigger value="authoring">
						{m["admin.exams.page.authoring"]()}
					</TabsTrigger>
				</TabsList>
				<TabsContent value="catalog" className="flex flex-col gap-4">
					<CardShell
						pending={catalog.isPending}
						className="flex flex-wrap gap-2 p-4"
					>
						{catalog.isError ? (
							<div role="alert">
								{m["admin.toast.error"]()}
								<SubmitButton
									pending={catalog.isFetching}
									onClick={() => void catalog.refetch()}
								>
									{m["admin.overview.metrics.retry"]()}
								</SubmitButton>
							</div>
						) : catalog.data?.length === 0 ? (
							<p>{m["admin.table.empty"]()}</p>
						) : (
							catalog.data?.map((row) => (
								<Button
									key={row.id}
									variant={categoryId === row.id ? "default" : "outline"}
									onClick={() => onNavigate(row.id, 1)}
								>
									{row.title} ({row.testsCount})
								</Button>
							))
						)}
					</CardShell>
					{!categoryId ? (
						<p>{m["admin.exams.page.select"]()}</p>
					) : tests.isError ? (
						<CardShell className="p-4">
							<p role="alert">{m["admin.toast.error"]()}</p>
							<SubmitButton
								pending={tests.isFetching}
								onClick={() => void tests.refetch()}
							>
								{m["admin.overview.metrics.retry"]()}
							</SubmitButton>
						</CardShell>
					) : (
						<>
							<DataTable
								rows={tests.data?.items ?? []}
								rowKey={(row) => row.id}
								isLoading={tests.isPending}
								columns={[
									{
										id: "title",
										header: m["admin.exams.fields.title"](),
										cell: (row) => row.title,
									},
									{
										id: "parts",
										header: m["admin.exams.editor.parts"](),
										cell: (row) => row.content?.parts.length ?? 0,
									},
								]}
							/>
							{!!tests.data?.items.length && (
								<Pager
									page={page}
									totalPages={tests.data.pagination.totalPages}
									onPageChange={(next) => onNavigate(categoryId, next)}
								/>
							)}
						</>
					)}
				</TabsContent>
				<TabsContent value="authoring" className="flex flex-col gap-4">
					<div className="flex flex-wrap gap-2">
						<Button
							onClick={() => {
								setCategoryForm({
									id: crypto.randomUUID(),
									slug: "",
									title: "",
									type: "toeic",
									tests: [],
								});
								setError(false);
							}}
						>
							{m["admin.actions.create"]()} —{" "}
							{m["admin.exams.page.categories"]()}
						</Button>
						<Button
							variant="outline"
							onClick={() => setRemove({ reset: true })}
						>
							{m["admin.exams.page.reset"]()}
						</Button>
					</div>
					<DataTable
						rows={categories}
						rowKey={(row) => row.id}
						columns={[
							{
								id: "title",
								header: m["admin.exams.fields.title"](),
								cell: (row) => row.title,
							},
							{
								id: "tests",
								header: m["admin.exams.page.tests"](),
								cell: (row) => row.tests.length,
							},
							{
								id: "actions",
								header: m["admin.table.actions"](),
								cell: (row) => (
									<div className="flex flex-wrap gap-2">
										<Button
											variant="outline"
											onClick={() => setSelectedId(row.id)}
										>
											{m["admin.exams.page.open"]()}
										</Button>
										<Button
											variant="outline"
											onClick={() => {
												setCategoryForm(structuredClone(row));
												setError(false);
											}}
										>
											{m["admin.actions.edit"]()}
										</Button>
										<Button
											variant="ghost"
											onClick={() => setRemove({ categoryId: row.id })}
										>
											{m["admin.actions.delete"]()}
										</Button>
									</div>
								),
							},
						]}
					/>
					{category ? (
						<>
							<h2 className="text-lg font-bold">{category.title}</h2>
							<Button
								className="self-start"
								onClick={() => {
									const test = newExamDraft();
									test.content.collectionId = category.slug;
									setTestForm(test);
									setError(false);
								}}
							>
								{m["admin.actions.create"]()} — {m["admin.exams.page.tests"]()}
							</Button>
							<DataTable
								rows={category.tests}
								rowKey={(row) => row.id}
								columns={[
									{
										id: "title",
										header: m["admin.exams.fields.title"](),
										cell: (row) => row.content.title,
									},
									{
										id: "parts",
										header: m["admin.exams.editor.parts"](),
										cell: (row) => row.content.parts.length,
									},
									{
										id: "actions",
										header: m["admin.table.actions"](),
										cell: (row) => (
											<div className="flex gap-2">
												<Button
													variant="outline"
													onClick={() => {
														setTestForm(row);
														setError(false);
													}}
												>
													{m["admin.actions.edit"]()}
												</Button>
												<Button
													variant="ghost"
													onClick={() =>
														setRemove({
															categoryId: category.id,
															testId: row.id,
														})
													}
												>
													{m["admin.actions.delete"]()}
												</Button>
											</div>
										),
									},
								]}
							/>
						</>
					) : (
						<p>{m["admin.exams.page.select"]()}</p>
					)}
				</TabsContent>
			</Tabs>
			{categoryForm && (
				<ResourceFormDialog
					open
					title={m["admin.exams.page.categories"]()}
					onOpenChange={(open) => {
						if (!open) setCategoryForm(null);
					}}
					onSubmit={() => {
						if (
							!categoryForm.title.trim() ||
							!categoryForm.slug.trim() ||
							categories.some(
								(row) =>
									row.id !== categoryForm.id &&
									row.slug === categoryForm.slug.trim(),
							)
						) {
							setError(true);
							return;
						}
						const value = {
							...categoryForm,
							title: categoryForm.title.trim(),
							slug: categoryForm.slug.trim(),
						};
						store.setState((rows) =>
							rows.some((row) => row.id === value.id)
								? rows.map((row) => (row.id === value.id ? value : row))
								: [...rows, value],
						);
						setSelectedId(value.id);
						setCategoryForm(null);
					}}
				>
					<AuthorField
						label={m["admin.exams.fields.title"]()}
						value={categoryForm.title}
						onChange={(title) => setCategoryForm({ ...categoryForm, title })}
					/>
					<AuthorField
						label={m["admin.exams.fields.slug"]()}
						value={categoryForm.slug}
						onChange={(slug) => setCategoryForm({ ...categoryForm, slug })}
					/>
					{error && (
						<FieldError>{m["admin.exams.page.invalidCategory"]()}</FieldError>
					)}
				</ResourceFormDialog>
			)}
			{testForm && (
				<ExamEditor
					key={testForm.id}
					initial={testForm}
					onSave={saveTest}
					onClose={() => setTestForm(null)}
				/>
			)}
			{error && !categoryForm && (
				<p role="alert" className="text-destructive">
					{m["admin.exams.editor.invalid"]()}
				</p>
			)}
			<ConfirmDeleteDialog
				open={remove !== null}
				onOpenChange={(open) => {
					if (!open) setRemove(null);
				}}
				onConfirm={() => {
					if (remove?.reset) {
						store.setState(() => createExamsStore().state);
						setSelectedId("demo-category");
					} else if (remove?.testId)
						store.setState((rows) =>
							rows.map((row) =>
								row.id === remove.categoryId
									? {
											...row,
											tests: row.tests.filter(
												(test) => test.id !== remove.testId,
											),
										}
									: row,
							),
						);
					else
						store.setState((rows) =>
							rows.filter((row) => row.id !== remove?.categoryId),
						);
					setRemove(null);
				}}
			/>
		</div>
	);
}
