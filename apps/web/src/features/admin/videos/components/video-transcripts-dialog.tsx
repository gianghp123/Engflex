import type {
	VideoExerciseResponse,
	VideoTranscriptResponse,
} from "@engflex/contracts";
import { useState } from "react";
import { toast } from "sonner";
import { CardShell } from "#/components/common/card-shell";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Field, FieldError, FieldLabel } from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import { Textarea } from "#/components/ui/textarea";
import { ConfirmDeleteDialog } from "#/features/admin/components/confirm-delete-dialog";
import {
	useAdminVideoTranscripts,
	useCreateVideoTranscript,
	useDeleteVideoTranscript,
	useUpdateVideoTranscript,
} from "#/features/admin/videos/queries";
import { m } from "#/paraglide/messages";

export function VideoTranscriptsDialog({
	video,
	onOpenChange,
}: {
	video: VideoExerciseResponse;
	onOpenChange: (open: boolean) => void;
}) {
	const query = useAdminVideoTranscripts(video.id);
	const create = useCreateVideoTranscript(video.id);
	const update = useUpdateVideoTranscript(video.id);
	const remove = useDeleteVideoTranscript(video.id);
	const [editing, setEditing] = useState<VideoTranscriptResponse | null>(null);
	const [target, setTarget] = useState<string | null>(null);
	const [formKey, setFormKey] = useState(0);
	function reset() {
		setEditing(null);
		setFormKey((value) => value + 1);
	}
	const rows = [...(query.data ?? [])].sort((a, b) => a.sequence - b.sequence);
	return (
		<Dialog open onOpenChange={onOpenChange}>
			<DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-3xl">
				<DialogHeader>
					<DialogTitle>
						{m["admin.videos.transcripts.title"]()} — {video.title}
					</DialogTitle>
				</DialogHeader>
				{query.isSuccess && (
					<TranscriptForm
						key={`${editing?.id ?? "new"}-${formKey}`}
						initial={editing}
						nextSequence={Math.max(0, ...rows.map((row) => row.sequence)) + 1}
						usedSequences={rows
							.filter((row) => row.id !== editing?.id)
							.map((row) => row.sequence)}
						duration={video.duration}
						pending={create.isPending || update.isPending}
						onCancel={reset}
						onSave={async (payload) => {
							if (editing)
								await update.mutateAsync({ id: editing.id, payload });
							else
								await create.mutateAsync({
									...payload,
									videoExerciseId: video.id,
								});
							toast.success(m["admin.toast.updated"]());
							reset();
						}}
					/>
				)}
				<CardShell
					pending={query.isPending}
					className="flex flex-col gap-3 p-4"
				>
					{query.isError ? (
						<div role="alert">
							{m["admin.toast.error"]()}{" "}
							<SubmitButton
								pending={query.isFetching}
								onClick={() => void query.refetch()}
							>
								{m["admin.videos.transcripts.retry"]()}
							</SubmitButton>
						</div>
					) : rows.length === 0 ? (
						<p>{m["admin.table.empty"]()}</p>
					) : (
						rows.map((row) => (
							<div
								key={row.id}
								className="flex items-start justify-between gap-3 border-b pb-3"
							>
								<div>
									<p className="text-xs text-muted-foreground">
										{row.sequence} · {row.startTimestamp}s–{row.endTimestamp}s
									</p>
									<p>{row.content}</p>
									{row.phonetic && (
										<p className="text-sm text-muted-foreground">
											{row.phonetic}
										</p>
									)}
									{row.vietnamese && (
										<p className="text-sm text-muted-foreground">
											{row.vietnamese}
										</p>
									)}
								</div>
								<div className="flex gap-2">
									<Button variant="outline" onClick={() => setEditing(row)}>
										{m["admin.actions.edit"]()}
									</Button>
									<Button variant="ghost" onClick={() => setTarget(row.id)}>
										{m["admin.actions.delete"]()}
									</Button>
								</div>
							</div>
						))
					)}
				</CardShell>
				<ConfirmDeleteDialog
					open={target !== null}
					onOpenChange={(open) => {
						if (!open && !remove.isPending) setTarget(null);
					}}
					pending={remove.isPending}
					onConfirm={() => {
						if (!target || remove.isPending) return;
						remove.mutate(target, {
							onSuccess: () => {
								if (editing?.id === target) reset();
								setTarget(null);
							},
							onError: () => toast.error(m["admin.toast.error"]()),
						});
					}}
				/>
			</DialogContent>
		</Dialog>
	);
}

function TranscriptForm({
	initial,
	nextSequence,
	usedSequences,
	duration,
	pending,
	onSave,
	onCancel,
}: {
	initial: VideoTranscriptResponse | null;
	nextSequence: number;
	usedSequences: number[];
	duration: number;
	pending: boolean;
	onSave: (payload: {
		sequence: number;
		content: string;
		phonetic: string;
		vietnamese: string;
		startTimestamp: number;
		endTimestamp: number;
	}) => Promise<void>;
	onCancel: () => void;
}) {
	const [form, setForm] = useState({
		sequence: String(initial?.sequence ?? nextSequence),
		content: initial?.content ?? "",
		phonetic: initial?.phonetic ?? "",
		vietnamese: initial?.vietnamese ?? "",
		startTimestamp: String(initial?.startTimestamp ?? 0),
		endTimestamp: String(initial?.endTimestamp ?? 1),
	});
	const [error, setError] = useState(false);
	async function submit() {
		if (pending) return;
		const sequence = Number(form.sequence),
			startTimestamp = Number(form.startTimestamp),
			endTimestamp = Number(form.endTimestamp);
		if (
			!Number.isInteger(sequence) ||
			sequence < 1 ||
			usedSequences.includes(sequence) ||
			!form.content.trim() ||
			!form.startTimestamp.trim() ||
			!form.endTimestamp.trim() ||
			!Number.isFinite(startTimestamp) ||
			!Number.isFinite(endTimestamp) ||
			startTimestamp < 0 ||
			endTimestamp <= startTimestamp ||
			(duration > 0 && endTimestamp > duration)
		) {
			setError(true);
			return;
		}
		try {
			await onSave({
				...form,
				content: form.content.trim(),
				sequence,
				startTimestamp,
				endTimestamp,
			});
		} catch {
			setError(true);
		}
	}
	return (
		<form
			className="flex flex-col gap-3"
			onSubmit={(event) => {
				event.preventDefault();
				void submit();
			}}
		>
			<div className="grid gap-3 sm:grid-cols-3">
				<Field>
					<FieldLabel htmlFor="transcript-sequence">
						{m["admin.videos.transcripts.sequence"]()}
					</FieldLabel>
					<Input
						id="transcript-sequence"
						type="number"
						min={1}
						value={form.sequence}
						onChange={(e) => setForm({ ...form, sequence: e.target.value })}
					/>
				</Field>
				<Field>
					<FieldLabel htmlFor="transcript-start">
						{m["admin.videos.transcripts.start"]()}
					</FieldLabel>
					<Input
						id="transcript-start"
						type="number"
						min={0}
						step="any"
						value={form.startTimestamp}
						onChange={(e) =>
							setForm({ ...form, startTimestamp: e.target.value })
						}
					/>
				</Field>
				<Field>
					<FieldLabel htmlFor="transcript-end">
						{m["admin.videos.transcripts.end"]()}
					</FieldLabel>
					<Input
						id="transcript-end"
						type="number"
						min={0}
						step="any"
						value={form.endTimestamp}
						onChange={(e) => setForm({ ...form, endTimestamp: e.target.value })}
					/>
				</Field>
			</div>
			<Field>
				<FieldLabel htmlFor="transcript-content">
					{m["admin.videos.transcripts.content"]()}
				</FieldLabel>
				<Textarea
					id="transcript-content"
					value={form.content}
					onChange={(e) => setForm({ ...form, content: e.target.value })}
				/>
			</Field>
			<Field>
				<FieldLabel htmlFor="transcript-phonetic">
					{m["admin.videos.transcripts.phonetic"]()}
				</FieldLabel>
				<Input
					id="transcript-phonetic"
					value={form.phonetic}
					onChange={(e) => setForm({ ...form, phonetic: e.target.value })}
				/>
			</Field>
			<Field>
				<FieldLabel htmlFor="transcript-vietnamese">
					{m["admin.videos.transcripts.vietnamese"]()}
				</FieldLabel>
				<Textarea
					id="transcript-vietnamese"
					value={form.vietnamese}
					onChange={(e) => setForm({ ...form, vietnamese: e.target.value })}
				/>
			</Field>
			{error && (
				<FieldError>{m["admin.videos.transcripts.invalid"]()}</FieldError>
			)}
			<div className="flex justify-end gap-2">
				{initial && (
					<Button type="button" variant="ghost" onClick={onCancel}>
						{m["admin.actions.cancel"]()}
					</Button>
				)}
				<SubmitButton type="submit" pending={pending}>
					{initial ? m["admin.actions.save"]() : m["admin.actions.create"]()}
				</SubmitButton>
			</div>
		</form>
	);
}
