import type { CEFR, SentenceInput, StoryItem } from "@engflex/contracts";
import { useQueryClient } from "@tanstack/react-query";
import {
	Eye,
	EyeOff,
	Loader2,
	Pencil,
	Plus,
	Trash2,
	Upload,
} from "lucide-react";
import { useRef, useState } from "react";
import { toast } from "sonner";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Field, FieldError, FieldLabel } from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { Switch } from "#/components/ui/switch";
import { Textarea } from "#/components/ui/textarea";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#/components/ui/tooltip";
import { ConfirmDeleteDialog } from "#/features/admin/components/confirm-delete-dialog";
import {
	DataTable,
	type DataTableColumn,
} from "#/features/admin/components/data-table";
import { ResourceFormDialog } from "#/features/admin/components/resource-form-dialog";
import { StatusBadge } from "#/features/admin/components/status-badge";
import {
	useAdminStories,
	useCreateStory,
	useDeleteStory,
	useUpdateStory,
	useUploadStoryAsset,
} from "#/features/admin/stories/queries";
import { storyDetailByIdQueryOptions } from "#/features/stories/queries";
import { m } from "#/paraglide/messages";

/** File-picker + upload button for one URL field: picks a file, uploads it
 *  to the stories MinIO bucket, and hands the resulting same-origin media
 *  URL back to the caller to drop into its text
 *  input. The text input stays editable too — pasting an external URL
 *  still works, this is just an alternative to typing one in. */
function UploadButton({
	accept,
	onUploaded,
}: {
	accept: string;
	onUploaded: (url: string) => void;
}) {
	const inputRef = useRef<HTMLInputElement>(null);
	const uploadMutation = useUploadStoryAsset();

	return (
		<>
			<input
				ref={inputRef}
				type="file"
				accept={accept}
				className="hidden"
				onChange={(event) => {
					const file = event.target.files?.[0];
					event.target.value = "";
					if (!file) return;
					uploadMutation.mutate(file, {
						onSuccess: ({ url }) => onUploaded(url),
						onError: () => toast.error(m["admin.toast.error"]()),
					});
				}}
			/>
			<Button
				type="button"
				variant="outline"
				size="icon-sm"
				aria-label="Upload"
				disabled={uploadMutation.isPending}
				onClick={() => inputRef.current?.click()}
			>
				{uploadMutation.isPending ? (
					<Loader2 className="animate-spin" />
				) : (
					<Upload />
				)}
			</Button>
		</>
	);
}

/** The admin pastes the full sentence timeline as one JSON array — there's
 *  no per-sentence editor (see the Stories admin scoping decision), so this
 *  is the only way to put reading content on a story at all. Empty text
 *  means "leave the existing timeline untouched" on edit (see UpdateStory);
 *  an explicit `[]` clears it. */
function parseSentences(raw: string): SentenceInput[] | undefined {
	const trimmed = raw.trim();
	if (!trimmed) return undefined;
	const parsed: unknown = JSON.parse(trimmed);
	if (!Array.isArray(parsed)) {
		throw new Error("Sentences must be a JSON array");
	}
	return parsed as SentenceInput[];
}

const CEFR_OPTIONS: CEFR[] = ["A1", "A2", "B1", "B2", "C1", "C2"];

interface StoryFormState {
	slug: string;
	title: string;
	titleVietnamese: string;
	cefrLevel: CEFR | "";
	topic: string;
	thumbnailUrl: string;
	audioUrl: string;
	duration: string;
	status: "draft" | "published";
	isPremium: boolean;
	sentencesJson: string;
}

const EMPTY_FORM: StoryFormState = {
	slug: "",
	title: "",
	titleVietnamese: "",
	cefrLevel: "",
	topic: "",
	thumbnailUrl: "",
	audioUrl: "",
	duration: "",
	status: "draft",
	isPremium: false,
	sentencesJson: "",
};

export function StoriesPanel() {
	const [page, setPage] = useState(1);
	const [search, setSearch] = useState("");
	const [cefrFilter, setCefrFilter] = useState<CEFR | "">("");
	const pageSize = 20;
	const query = useAdminStories(page, pageSize, {
		search,
		cefr: cefrFilter,
	});
	const createMutation = useCreateStory();
	const updateMutation = useUpdateStory();
	const deleteMutation = useDeleteStory();

	const [dialogOpen, setDialogOpen] = useState(false);
	const [editing, setEditing] = useState<StoryItem | null>(null);
	const [form, setForm] = useState<StoryFormState>(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const [deleteTarget, setDeleteTarget] = useState<StoryItem | null>(null);
	const [loadingSentences, setLoadingSentences] = useState(false);
	const queryClient = useQueryClient();

	function openCreate() {
		setEditing(null);
		setForm(EMPTY_FORM);
		setError(null);
		setDialogOpen(true);
	}

	async function openEdit(story: StoryItem) {
		setEditing(story);
		setForm({
			slug: story.slug,
			title: story.title,
			titleVietnamese: story.titleVietnamese ?? "",
			cefrLevel: story.cefrLevel ?? "",
			topic: story.topic,
			thumbnailUrl: story.thumbnailUrl ?? "",
			audioUrl: story.audioUrl,
			duration: String(story.duration),
			status: story.status === "published" ? "published" : "draft",
			isPremium: story.isPremium,
			sentencesJson: "",
		});
		setError(null);
		setDialogOpen(true);
		// The list row has no sentences (StoryItem omits them) — fetch the
		// full detail just-in-time so the textarea starts from what's
		// actually saved, not blank (which would read as "clear on submit").
		setLoadingSentences(true);
		try {
			const detail = await queryClient.fetchQuery(
				storyDetailByIdQueryOptions(story.id),
			);
			const sentences = detail.sentences
				.filter((sentence) => sentence != null)
				.map(
					({
						sequence,
						englishText,
						vietnameseText,
						imageUrl,
						startTimestamp,
						endTimestamp,
					}) => ({
						sequence,
						englishText,
						vietnameseText,
						imageUrl: imageUrl ?? undefined,
						startTimestamp,
						endTimestamp,
					}),
				);
			setForm((current) => ({
				...current,
				sentencesJson: JSON.stringify(sentences, null, 2),
			}));
		} finally {
			setLoadingSentences(false);
		}
	}

	function submit() {
		if (!form.slug.trim() || !form.title.trim() || !form.audioUrl.trim()) {
			setError(m["admin.toast.error"]());
			return;
		}
		let sentences: SentenceInput[] | undefined;
		try {
			sentences = parseSentences(form.sentencesJson);
		} catch {
			setError("Sentences must be valid JSON array");
			return;
		}
		const base = {
			slug: form.slug.trim(),
			title: form.title.trim(),
			titleVietnamese: form.titleVietnamese.trim() || undefined,
			cefrLevel: form.cefrLevel || undefined,
			topic: form.topic.trim() || "general",
			thumbnailUrl: form.thumbnailUrl.trim() || undefined,
			audioUrl: form.audioUrl.trim(),
			duration: Number(form.duration) || 0,
			isPremium: form.isPremium,
			sentences,
		};
		if (editing) {
			updateMutation.mutate(
				{ id: editing.id, payload: { ...base, status: form.status } },
				{
					onSuccess: () => {
						toast.success(m["admin.toast.updated"]());
						setDialogOpen(false);
					},
					onError: () => toast.error(m["admin.toast.error"]()),
				},
			);
		} else {
			createMutation.mutate(
				{ ...base, sentences: sentences ?? [] },
				{
					onSuccess: () => {
						toast.success(m["admin.toast.created"]());
						setDialogOpen(false);
					},
					onError: () => toast.error(m["admin.toast.error"]()),
				},
			);
		}
	}

	function togglePublish(story: StoryItem) {
		const nextStatus = story.status === "published" ? "draft" : "published";
		updateMutation.mutate(
			{
				id: story.id,
				payload: {
					slug: story.slug,
					title: story.title,
					titleVietnamese: story.titleVietnamese,
					cefrLevel: story.cefrLevel,
					topic: story.topic,
					thumbnailUrl: story.thumbnailUrl,
					audioUrl: story.audioUrl,
					duration: story.duration,
					status: nextStatus,
					isPremium: story.isPremium,
				},
			},
			{
				onSuccess: () =>
					toast.success(
						nextStatus === "published"
							? m["admin.toast.published"]()
							: m["admin.toast.unpublished"](),
					),
				onError: () => toast.error(m["admin.toast.error"]()),
			},
		);
	}

	function confirmDelete() {
		if (!deleteTarget) return;
		deleteMutation.mutate(deleteTarget.id, {
			onSuccess: () => {
				toast.success(m["admin.toast.deleted"]());
				setDeleteTarget(null);
			},
			onError: () => toast.error(m["admin.toast.error"]()),
		});
	}

	const columns: DataTableColumn<StoryItem>[] = [
		{ id: "title", header: "Title", cell: (row) => row.title },
		{
			id: "topic",
			header: "Topic",
			className: "w-36",
			cell: (row) => row.topic,
		},
		{
			id: "cefr",
			header: "CEFR",
			className: "w-20",
			cell: (row) => row.cefrLevel ?? "—",
		},
		{
			id: "status",
			header: m["admin.table.status"](),
			className: "w-28",
			cell: (row) => (
				<div className="flex items-center gap-1.5">
					<StatusBadge
						status={row.status === "published" ? "published" : "draft"}
					/>
					{row.isPremium && (
						<Badge variant="default">{m["admin.content.premiumBadge"]()}</Badge>
					)}
				</div>
			),
		},
		{
			id: "actions",
			header: m["admin.table.actions"](),
			className: "w-28",
			cell: (row) => (
				<div className="flex items-center gap-1">
					<Tooltip>
						<TooltipTrigger asChild>
							<Button
								type="button"
								variant="outline"
								size="icon-sm"
								aria-label={m["admin.actions.edit"]()}
								onClick={() => openEdit(row)}
							>
								<Pencil />
							</Button>
						</TooltipTrigger>
						<TooltipContent>{m["admin.actions.edit"]()}</TooltipContent>
					</Tooltip>
					<Tooltip>
						<TooltipTrigger asChild>
							<Button
								type="button"
								variant="outline"
								size="icon-sm"
								aria-label={
									row.status === "published"
										? m["admin.actions.unpublish"]()
										: m["admin.actions.publish"]()
								}
								onClick={() => togglePublish(row)}
							>
								{row.status === "published" ? <EyeOff /> : <Eye />}
							</Button>
						</TooltipTrigger>
						<TooltipContent>
							{row.status === "published"
								? m["admin.actions.unpublish"]()
								: m["admin.actions.publish"]()}
						</TooltipContent>
					</Tooltip>
					<Tooltip>
						<TooltipTrigger asChild>
							<Button
								type="button"
								variant="ghost"
								size="icon-sm"
								aria-label={m["admin.actions.delete"]()}
								onClick={() => setDeleteTarget(row)}
							>
								<Trash2 />
							</Button>
						</TooltipTrigger>
						<TooltipContent>{m["admin.actions.delete"]()}</TooltipContent>
					</Tooltip>
				</div>
			),
		},
	];

	return (
		<div className="flex flex-col gap-4">
			<div className="flex flex-wrap items-center gap-3">
				<Input
					placeholder="Search by title"
					value={search}
					onChange={(event) => {
						setSearch(event.target.value);
						setPage(1);
					}}
					className="max-w-sm"
				/>
				<Select
					value={cefrFilter || "all"}
					onValueChange={(value) => {
						setCefrFilter(value === "all" ? "" : (value as CEFR));
						setPage(1);
					}}
				>
					<SelectTrigger className="w-32">
						<SelectValue />
					</SelectTrigger>
					<SelectContent>
						<SelectItem value="all">All CEFR</SelectItem>
						{CEFR_OPTIONS.map((option) => (
							<SelectItem key={option} value={option}>
								{option}
							</SelectItem>
						))}
					</SelectContent>
				</Select>
				<Button type="button" onClick={openCreate} className="ml-auto">
					<Plus data-icon="inline-start" />
					{m["admin.actions.create"]()}
				</Button>
			</div>
			<DataTable
				columns={columns}
				rows={query.data?.items ?? []}
				rowKey={(row) => row.id}
				isLoading={query.isPending}
				pagination={query.data?.pagination}
				onPageChange={setPage}
			/>
			<ResourceFormDialog
				open={dialogOpen}
				onOpenChange={setDialogOpen}
				title={
					editing ? m["admin.actions.edit"]() : m["admin.actions.create"]()
				}
				onSubmit={submit}
				pending={createMutation.isPending || updateMutation.isPending}
			>
				<Field data-invalid={Boolean(error)}>
					<FieldLabel htmlFor="story-title">Title</FieldLabel>
					<Input
						id="story-title"
						value={form.title}
						onChange={(event) =>
							setForm((current) => ({ ...current, title: event.target.value }))
						}
					/>
				</Field>
				<div className="grid gap-3 sm:grid-cols-2">
					<Field data-invalid={Boolean(error)}>
						<FieldLabel htmlFor="story-slug">Slug</FieldLabel>
						<Input
							id="story-slug"
							value={form.slug}
							onChange={(event) =>
								setForm((current) => ({ ...current, slug: event.target.value }))
							}
						/>
					</Field>
					<Field>
						<FieldLabel htmlFor="story-title-vi">Title (Vietnamese)</FieldLabel>
						<Input
							id="story-title-vi"
							value={form.titleVietnamese}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									titleVietnamese: event.target.value,
								}))
							}
						/>
					</Field>
				</div>
				<div className="grid gap-3 sm:grid-cols-3">
					<Field>
						<FieldLabel>CEFR</FieldLabel>
						<Select
							value={form.cefrLevel || "none"}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									cefrLevel: value === "none" ? "" : (value as CEFR),
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">—</SelectItem>
								{CEFR_OPTIONS.map((option) => (
									<SelectItem key={option} value={option}>
										{option}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
					<Field>
						<FieldLabel htmlFor="story-topic">Topic</FieldLabel>
						<Input
							id="story-topic"
							value={form.topic}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									topic: event.target.value,
								}))
							}
						/>
					</Field>
					<Field>
						<FieldLabel htmlFor="story-duration">Duration (s)</FieldLabel>
						<Input
							id="story-duration"
							type="number"
							value={form.duration}
							onChange={(event) =>
								setForm((current) => ({
									...current,
									duration: event.target.value,
								}))
							}
						/>
					</Field>
				</div>
				<div className="grid gap-3 sm:grid-cols-2">
					<Field>
						<FieldLabel htmlFor="story-thumbnail">Thumbnail URL</FieldLabel>
						<div className="flex gap-2">
							<Input
								id="story-thumbnail"
								value={form.thumbnailUrl}
								onChange={(event) =>
									setForm((current) => ({
										...current,
										thumbnailUrl: event.target.value,
									}))
								}
							/>
							<UploadButton
								accept="image/*"
								onUploaded={(url) =>
									setForm((current) => ({ ...current, thumbnailUrl: url }))
								}
							/>
						</div>
					</Field>
					<Field data-invalid={Boolean(error)}>
						<FieldLabel htmlFor="story-audio">Audio URL</FieldLabel>
						<div className="flex gap-2">
							<Input
								id="story-audio"
								value={form.audioUrl}
								onChange={(event) =>
									setForm((current) => ({
										...current,
										audioUrl: event.target.value,
									}))
								}
							/>
							<UploadButton
								accept="audio/*"
								onUploaded={(url) =>
									setForm((current) => ({ ...current, audioUrl: url }))
								}
							/>
						</div>
					</Field>
				</div>
				<Field>
					<FieldLabel htmlFor="story-sentences">
						Sentences (JSON array)
						{loadingSentences ? " — loading…" : ""}
					</FieldLabel>
					<Textarea
						id="story-sentences"
						rows={10}
						className="font-mono text-xs"
						placeholder={
							'[\n  {\n    "sequence": 1,\n    "englishText": "...",\n    "vietnameseText": "...",\n    "startTimestamp": 0,\n    "endTimestamp": 3\n  }\n]'
						}
						value={form.sentencesJson}
						disabled={loadingSentences}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								sentencesJson: event.target.value,
							}))
						}
					/>
				</Field>
				<Field className="flex-row items-center justify-between">
					<FieldLabel htmlFor="story-premium">
						{m["admin.content.isPremium"]()}
					</FieldLabel>
					<Switch
						id="story-premium"
						checked={form.isPremium}
						onCheckedChange={(checked) =>
							setForm((current) => ({ ...current, isPremium: checked }))
						}
					/>
				</Field>
				{editing ? (
					<Field>
						<FieldLabel>{m["admin.table.status"]()}</FieldLabel>
						<Select
							value={form.status}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									status: value as "draft" | "published",
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="draft">
									{m["admin.status.draft"]()}
								</SelectItem>
								<SelectItem value="published">
									{m["admin.status.published"]()}
								</SelectItem>
							</SelectContent>
						</Select>
					</Field>
				) : null}
				{error ? <FieldError>{error}</FieldError> : null}
			</ResourceFormDialog>
			<ConfirmDeleteDialog
				open={Boolean(deleteTarget)}
				onOpenChange={(open) => !open && setDeleteTarget(null)}
				onConfirm={confirmDelete}
				pending={deleteMutation.isPending}
			/>
		</div>
	);
}
