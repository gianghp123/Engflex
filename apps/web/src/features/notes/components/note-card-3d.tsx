import { NoteMaxLength, type SavedItem } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { cn } from "cn";
import { useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { CardShell } from "#/components/common/card-shell";
import { SubmitButton } from "#/components/common/submit-button";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Textarea } from "#/components/ui/textarea";
import {
	useRemoveSavedItem,
	useUpdateSavedNote,
} from "#/features/bookmarks/queries";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
import type { NoteKind } from "../types";

const kindLabels: Record<NoteKind, () => string> = {
	vocab: () => m["notes.tabs.vocab"](),
	lesson: () => m["notes.tabs.lesson"](),
	video: () => m["notes.tabs.video"](),
	story: () => m["notes.tabs.story"](),
};

function contentLink(item: SavedItem) {
	if (item.lesson)
		return {
			to: APP_ROUTES.LESSONS.DETAIL,
			params: { slug: item.lesson.slug },
		};
	if (item.story)
		return { to: APP_ROUTES.STORIES.DETAIL, params: { slug: item.story.slug } };
	if (item.videoExercise)
		return {
			to: APP_ROUTES.VIDEOS.DICTATION,
			params: { videoId: item.contentId },
		};
	if (item.vocabularyItem)
		return {
			to: APP_ROUTES.VOCABULARY.DETAIL,
			params: { itemId: item.contentId },
		};
}

export function NoteCard3D({
	item,
	kind,
	onRemoved,
}: {
	item: SavedItem;
	kind: NoteKind;
	onRemoved: () => void;
}) {
	const [showNote, setShowNote] = useState(false);
	const [editing, setEditing] = useState(false);
	const [draft, setDraft] = useState("");
	const update = useUpdateSavedNote();
	const remove = useRemoveSavedItem();
	const link = contentLink(item);
	const title =
		item.lesson?.title ??
		item.videoExercise?.title ??
		item.story?.title ??
		item.vocabularyItem?.term ??
		m["notes.card.unavailable"]();
	const summary =
		item.lesson?.description ??
		item.videoExercise?.description ??
		item.vocabularyItem?.definition;
	const date = new Intl.DateTimeFormat(getLocale(), {
		dateStyle: "medium",
	}).format(new Date(item.createdAt));
	const onError = () => toast.error(m["notes.toast.error"]());

	function save(note: string) {
		update.mutate(
			{ type: kind, id: item.contentId, note },
			{
				onSuccess: () => {
					setEditing(false);
					toast.success(m["notes.toast.saved"]());
				},
				onError,
			},
		);
	}

	async function removeItem() {
		if (
			kind === "vocab" &&
			!window.confirm(m["notes.actions.confirmUnsaveWord"]())
		) {
			return;
		}
		try {
			await remove.mutateAsync({ type: kind, id: item.contentId });
			toast.success(m["notes.toast.removed"]());
			onRemoved();
		} catch {
			onError();
		}
	}

	return (
		<CardShell className="flex min-h-64 flex-col gap-3 rounded-2xl border p-5">
			<div className="flex flex-wrap items-center justify-between gap-2">
				<Badge variant="outline">{kindLabels[kind]()}</Badge>
				<span className="text-xs text-muted-foreground">
					{m["notes.card.savedDatePrefix"]()} {date}
				</span>
			</div>
			<h2 className="text-base font-bold">{title}</h2>
			<div className="[perspective:1200px]">
				<div
					className={cn(
						"grid [transform-style:preserve-3d] motion-safe:transition-transform motion-safe:duration-500",
						showNote && "[transform:rotateY(180deg)]",
					)}
				>
					<div
						inert={!showNote}
						className="col-start-1 row-start-1 space-y-2 [backface-visibility:hidden] [transform:rotateY(180deg)]"
					>
						<p className="text-xs font-semibold">
							{m["notes.card.personalNoteLabel"]()}
						</p>
						{editing ? (
							<>
								<Textarea
									aria-label={m["notes.card.personalNoteLabel"]()}
									value={draft}
									maxLength={NoteMaxLength}
									onChange={(event) => setDraft(event.target.value)}
								/>
								<div className="flex justify-end gap-2">
									<Button
										size="sm"
										variant="ghost"
										disabled={update.isPending}
										onClick={() => setEditing(false)}
									>
										{m["notes.actions.cancel"]()}
									</Button>
									<SubmitButton
										size="sm"
										pending={update.isPending}
										onClick={() => save(draft)}
									>
										{m["notes.actions.save"]()}
									</SubmitButton>
								</div>
							</>
						) : (
							<>
								<p className="whitespace-pre-wrap break-words text-sm">
									{item.note || m["notes.card.emptyNote"]()}
								</p>
								<div className="flex flex-wrap gap-2">
									<Button
										size="sm"
										variant="outline"
										disabled={update.isPending}
										onClick={() => {
											setDraft(item.note ?? "");
											setEditing(true);
										}}
									>
										{m["notes.actions.editNote"]()}
									</Button>
									{item.note && (
										<SubmitButton
											size="sm"
											variant="ghost"
											pending={update.isPending}
											onClick={() => save("")}
										>
											{m["notes.actions.clearNote"]()}
										</SubmitButton>
									)}
								</div>
							</>
						)}
					</div>
					<div
						inert={showNote}
						className="col-start-1 row-start-1 space-y-2 [backface-visibility:hidden]"
					>
						{summary && (
							<p className="line-clamp-2 text-sm text-muted-foreground">
								{summary}
							</p>
						)}
						<p className="line-clamp-2 text-sm italic">
							{item.note || m["notes.card.emptyNote"]()}
						</p>
					</div>
				</div>
			</div>
			<div className="mt-auto flex flex-wrap items-center gap-2">
				<Button
					size="sm"
					variant="outline"
					disabled={update.isPending}
					onClick={() => {
						setShowNote(!showNote);
						setEditing(false);
					}}
				>
					{showNote
						? m["notes.actions.flipBack"]()
						: m["notes.actions.flipNote"]()}
				</Button>
				{link && (
					<Button asChild size="sm">
						<Link {...link}>{m["notes.actions.openContent"]()}</Link>
					</Button>
				)}
				<SubmitButton
					size="sm"
					variant="ghost"
					pending={remove.isPending}
					disabled={update.isPending}
					onClick={() => void removeItem()}
				>
					{m["notes.actions.unsave"]()}
				</SubmitButton>
			</div>
		</CardShell>
	);
}
