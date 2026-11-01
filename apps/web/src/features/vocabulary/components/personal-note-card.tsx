import { Pencil, Trash2 } from "lucide-react";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import { Textarea } from "#/components/ui/textarea";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";

function formatAdded(createdAt: string) {
	return new Intl.DateTimeFormat(getLocale(), {
		month: "short",
		day: "numeric",
	}).format(new Date(createdAt));
}

export function PersonalNoteCard({
	note,
	createdAt,
	onSave,
}: {
	note: string | undefined;
	createdAt: string;
	onSave: (note: string) => void;
}) {
	const [editing, setEditing] = useState(false);
	const [draft, setDraft] = useState(note ?? "");

	return (
		<section className="flex flex-col gap-3 rounded-xl border bg-card p-5">
			<div className="flex items-center justify-between gap-2">
				<h2 className="text-sm font-bold tracking-tight text-foreground">
					{m["vocabulary.detail.noteTitle"]()}
				</h2>
				<span className="text-[11px] text-muted-foreground">
					{m["vocabulary.detail.addedOn"]({ date: formatAdded(createdAt) })}
				</span>
			</div>

			{editing ? (
				<div className="flex flex-col gap-2">
					<Textarea
						value={draft}
						onChange={(event) => setDraft(event.target.value)}
						rows={3}
					/>
					<div className="flex justify-end gap-2">
						<Button
							type="button"
							variant="ghost"
							size="sm"
							onClick={() => {
								setDraft(note ?? "");
								setEditing(false);
							}}
						>
							{m["common.actions.cancel"]()}
						</Button>
						<Button
							type="button"
							size="sm"
							onClick={() => {
								onSave(draft);
								setEditing(false);
							}}
						>
							{m["vocabulary.detail.saveNote"]()}
						</Button>
					</div>
				</div>
			) : (
				<>
					<p className="text-sm text-muted-foreground">
						{note ? `“${note}”` : m["vocabulary.detail.noNote"]()}
					</p>
					<div className="flex gap-2">
						<Button
							type="button"
							variant="outline"
							size="sm"
							onClick={() => {
								setDraft(note ?? "");
								setEditing(true);
							}}
						>
							<Pencil data-icon="inline-start" />
							{m["vocabulary.detail.editNote"]()}
						</Button>
						<Button
							type="button"
							variant="ghost"
							size="sm"
							onClick={() => onSave("")}
						>
							<Trash2 data-icon="inline-start" />
							{m["vocabulary.detail.deleteNote"]()}
						</Button>
					</div>
				</>
			)}
		</section>
	);
}
