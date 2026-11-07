import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "#/components/ui/dialog";
import { Input } from "#/components/ui/input";
import { Label } from "#/components/ui/label";
import { Textarea } from "#/components/ui/textarea";
import { m } from "#/paraglide/messages";
import { notesActions } from "../store";
import type { NoteKind } from "../types";

const KIND_OPTIONS: Array<{ value: NoteKind; label: () => string }> = [
	{ value: "vocab", label: () => m["notes.tabs.vocab"]() },
	{ value: "lesson", label: () => m["notes.tabs.lesson"]() },
	{ value: "video", label: () => m["notes.tabs.video"]() },
	{ value: "story", label: () => m["notes.tabs.story"]() },
];

export function NoteCreateDialog() {
	const [open, setOpen] = useState(false);
	const [kind, setKind] = useState<NoteKind>("vocab");
	const [title, setTitle] = useState("");
	const [personalNote, setPersonalNote] = useState("");
	const [tagsString, setTagsString] = useState("");

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (!title.trim()) return;

		const tags = tagsString
			.split(",")
			.map((t) => t.trim().toLowerCase())
			.filter(Boolean);

		const linkMap: Record<NoteKind, string> = {
			vocab: APP_ROUTES.VOCABULARY.LIST,
			lesson: APP_ROUTES.LESSONS.LIST,
			video: APP_ROUTES.VIDEOS.LIST,
			story: APP_ROUTES.STORIES.LIST,
		};

		notesActions.addItem({
			kind,
			title: title.trim(),
			personalNote: personalNote.trim(),
			tags: tags.length > 0 ? tags : ["personal-note"],
			linkTo: linkMap[kind],
			summary: personalNote.trim(),
		});

		toast.success(m["notes.toast.created"]());
		setTitle("");
		setPersonalNote("");
		setTagsString("");
		setOpen(false);
	};

	return (
		<Dialog open={open} onOpenChange={setOpen}>
			<DialogTrigger asChild>
				<Button className="gap-2 font-medium">
					<Plus className="size-4" />
					<span>{m["notes.actions.newNote"]()}</span>
				</Button>
			</DialogTrigger>

			<DialogContent className="sm:max-w-[480px]">
				<form onSubmit={handleSubmit}>
					<DialogHeader>
						<DialogTitle>{m["notes.dialog.newTitle"]()}</DialogTitle>
						<DialogDescription>
							{m["notes.hub.description"]()}
						</DialogDescription>
					</DialogHeader>

					<div className="grid gap-4 py-4">
						<div className="grid gap-2">
							<Label htmlFor="kind">{m["notes.dialog.kindLabel"]()}</Label>
							<div className="grid grid-cols-4 gap-2">
								{KIND_OPTIONS.map((opt) => (
									<Button
										key={opt.value}
										type="button"
										variant={kind === opt.value ? "default" : "outline"}
										size="sm"
										className="text-xs"
										onClick={() => setKind(opt.value)}
									>
										{opt.label()}
									</Button>
								))}
							</div>
						</div>

						<div className="grid gap-2">
							<Label htmlFor="title">{m["notes.dialog.titleLabel"]()}</Label>
							<Input
								id="title"
								value={title}
								onChange={(e) => setTitle(e.target.value)}
								placeholder={m["notes.dialog.titlePlaceholder"]()}
								required
							/>
						</div>

						<div className="grid gap-2">
							<Label htmlFor="personalNote">
								{m["notes.dialog.noteLabel"]()}
							</Label>
							<Textarea
								id="personalNote"
								value={personalNote}
								onChange={(e) => setPersonalNote(e.target.value)}
								placeholder={m["notes.dialog.notePlaceholder"]()}
								rows={4}
							/>
						</div>

						<div className="grid gap-2">
							<Label htmlFor="tags">{m["notes.dialog.tagsLabel"]()}</Label>
							<Input
								id="tags"
								value={tagsString}
								onChange={(e) => setTagsString(e.target.value)}
								placeholder={m["notes.dialog.tagsPlaceholder"]()}
							/>
						</div>
					</div>

					<DialogFooter>
						<DialogClose asChild>
							<Button type="button" variant="outline">
								{m["notes.actions.cancel"]()}
							</Button>
						</DialogClose>
						<Button type="submit">{m["notes.actions.save"]()}</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
