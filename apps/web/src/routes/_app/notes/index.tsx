import { createFileRoute } from "@tanstack/react-router";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { EmptyList } from "#/components/common/empty-list";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { NoteCard3D } from "#/features/notes/components/note-card-3d";
import { NotesFilterBar } from "#/features/notes/components/notes-filter-bar";
import { NotesStats } from "#/features/notes/components/notes-stats";
import { notesActions, useFilteredNotes } from "#/features/notes/store";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/notes/")({
	component: NotesPage,
});

function NotesPage() {
	useBreadcrumbs([{ label: m["notes.hub.title"]() }]);
	const notes = useFilteredNotes();

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="heart" size={72} />,
				title: m["notes.hub.title"](),
				description: m["notes.hub.description"](),
			}}
		>
			<div className="flex flex-col gap-6">
				<NotesStats />
				<NotesFilterBar />

				{notes.length === 0 ? (
					<EmptyList
						title={m["notes.empty.title"]()}
						description={m["notes.empty.description"]()}
						onReset={notesActions.resetFilters}
					/>
				) : (
					<div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
						{notes.map((item) => (
							<NoteCard3D key={item.id} item={item} />
						))}
					</div>
				)}
			</div>
		</PageLayout>
	);
}
