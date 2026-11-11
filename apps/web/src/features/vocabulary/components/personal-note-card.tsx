import { NoteMaxLength } from "@engflex/contracts";
import { Pencil, Trash2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { SubmitButton } from "#/components/common/submit-button";
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
	pending,
}: {
	note: string | undefined;
	createdAt: string;
	onSave: (note: string) => Promise<void>;
	pending: boolean;
}) {
	const [editing, setEditing] = useState(false);
	const [draft, setDraft] = useState(note ?? "");

	async function save(value: string) {
		try {
			await onSave(value);
			setEditing(false);
		} catch {
			toast.error(m["common.toast.error"]());
		}
	}

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
						maxLength={NoteMaxLength}
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
						<SubmitButton
							type="button"
							size="sm"
							pending={pending}
							onClick={() => void save(draft)}
						>
							{m["vocabulary.detail.saveNote"]()}
						</SubmitButton>
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
						<SubmitButton
							type="button"
							variant="ghost"
							size="sm"
							pending={pending}
							onClick={() => void save("")}
						>
							<Trash2 data-icon="inline-start" />
							{m["vocabulary.detail.deleteNote"]()}
						</SubmitButton>
					</div>
				</>
			)}
		</section>
	);
}
