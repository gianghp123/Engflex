import { TitleMaxLength } from "@engflex/contracts";
import { createFileRoute } from "@tanstack/react-router";
import { useEffect } from "react";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { CardShell } from "#/components/common/card-shell";
import { EmptyList } from "#/components/common/empty-list";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { Pager } from "#/components/common/pager";
import { SubmitButton } from "#/components/common/submit-button";
import { useSavedLibrary } from "#/features/bookmarks/queries";
import { NoteCard3D } from "#/features/notes/components/note-card-3d";
import { NotesFilterBar } from "#/features/notes/components/notes-filter-bar";
import { NotesStats } from "#/features/notes/components/notes-stats";
import { NOTE_KINDS } from "#/features/notes/types";
import { parseEnumParam, parsePage } from "#/lib/router/search-params";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/notes/")({
	validateSearch: (search: Record<string, unknown>) => ({
		type: parseEnumParam(NOTE_KINDS, search.type),
		q:
			typeof search.q === "string"
				? search.q.slice(0, TitleMaxLength)
				: undefined,
		page: parsePage(search.page),
	}),
	component: NotesPage,
});

function NotesPage() {
	useBreadcrumbs([{ label: m["notes.hub.title"]() }]);
	const { type, q, page = 1 } = Route.useSearch();
	const navigate = Route.useNavigate();
	const library = useSavedLibrary(type ?? "", q?.trim() ?? "", page);
	const stats = library.data?.data;
	const pagination = library.data?.pagination;
	const lastPage = Math.max(1, pagination?.totalPages ?? page);

	useEffect(() => {
		if (page > lastPage) {
			void navigate({ replace: true, search: { type, q, page: lastPage } });
		}
	}, [lastPage, navigate, page, q, type]);

	const goToPage = (next: number) =>
		void navigate({ search: { type, q, page: next } });

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="heart" size={72} />,
				title: m["notes.hub.title"](),
				description: m["notes.hub.description"](),
			}}
		>
			<div className="flex flex-col gap-6">
				<NotesStats stats={stats} pending={library.isPending} />
				<NotesFilterBar
					tab={type ?? "all"}
					search={q ?? ""}
					stats={stats}
					onTab={(next) =>
						void navigate({
							search: {
								type: next === "all" ? undefined : next,
								q,
								page: undefined,
							},
						})
					}
					onSearch={(next) =>
						void navigate({
							replace: true,
							search: { type, q: next || undefined, page: undefined },
						})
					}
				/>
				{library.isPending ? (
					<CardShell pending className="min-h-64 rounded-2xl" />
				) : library.isError ? (
					<CardShell className="space-y-3 rounded-2xl p-5">
						<p role="alert">{m["notes.toast.error"]()}</p>
						<SubmitButton
							pending={library.isFetching}
							onClick={() => void library.refetch()}
						>
							{m["notes.actions.retry"]()}
						</SubmitButton>
					</CardShell>
				) : stats?.total === 0 ? (
					<CardShell className="space-y-2 rounded-2xl p-5 text-center">
						<h2 className="font-semibold">{m["notes.empty.libraryTitle"]()}</h2>
						<p className="text-muted-foreground">
							{m["notes.empty.libraryDescription"]()}
						</p>
					</CardShell>
				) : pagination?.total === 0 ? (
					<EmptyList
						title={m["notes.empty.title"]()}
						description={m["notes.empty.description"]()}
						onReset={() =>
							void navigate({
								search: { type: undefined, q: undefined, page: undefined },
							})
						}
					/>
				) : (
					<>
						<div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
							{stats?.items.map((item) => {
								const kind = parseEnumParam(NOTE_KINDS, item.contentType);
								return (
									kind && (
										<NoteCard3D
											key={`${kind}:${item.contentId}`}
											item={item}
											kind={kind}
										/>
									)
								);
							})}
						</div>
						{pagination && (
							<Pager
								page={Math.min(page, lastPage)}
								totalPages={pagination.totalPages}
								onPageChange={goToPage}
							/>
						)}
					</>
				)}
			</div>
		</PageLayout>
	);
}
