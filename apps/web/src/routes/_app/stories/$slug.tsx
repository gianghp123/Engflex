import type { StorySentence } from "@engflex/contracts";
import { createFileRoute } from "@tanstack/react-router";
import { cn } from "cn";
import { Volume2 } from "lucide-react";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { CardShell } from "#/components/common/card-shell";
import { SubmitButton } from "#/components/common/submit-button";
import {
	useStoryBookmarks,
	useToggleStoryBookmark,
} from "#/features/bookmarks/queries";
import { StoryAudioPlayer } from "#/features/stories/components/story-audio-player";
import { StoryReaderHeader } from "#/features/stories/components/story-reader-header";
import {
	storyDetailBySlugQueryOptions,
	useStoryDetail,
} from "#/features/stories/queries";
import { readerPreferences } from "#/features/stories/store";
import { loadOr404 } from "#/lib/router/route-loader";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/stories/$slug")({
	loader: async ({ context: { queryClient }, params: { slug } }) => {
		await loadOr404(queryClient, storyDetailBySlugQueryOptions(slug));
	},
	component: StoryDetailPage,
});

function StoryDetailPage() {
	const { slug } = Route.useParams();
	const storyQuery = useStoryDetail(slug);
	const story = storyQuery.data;

	useBreadcrumbs(
		story
			? [
					{ label: m["stories.title"](), to: APP_ROUTES.STORIES.LIST },
					{ label: story.title },
				]
			: [],
	);

	const [mode, setMode] = useState<"listen" | "read">("listen");
	const [settings, setSettings] = readerPreferences.usePreferences();
	const { showTranslation } = settings;
	const bookmarks = useStoryBookmarks();
	const toggleBookmark = useToggleStoryBookmark();
	const saved =
		bookmarks.data?.some((bookmark) => bookmark.contentId === story?.id) ??
		false;
	const [currentTime, setCurrentTime] = useState<number>(0);

	function handleSentenceClick(startTimestamp: number) {
		setCurrentTime(startTimestamp);
	}

	const sentences = useMemo(
		() =>
			(story?.sentences ?? []).filter((s): s is StorySentence => Boolean(s)),
		[story?.sentences],
	);

	const activeSentence = useMemo(() => {
		if (sentences.length === 0) return undefined;
		const exact = sentences.find(
			(sent) =>
				currentTime >= sent.startTimestamp && currentTime < sent.endTimestamp,
		);
		if (exact) return exact;
		if (currentTime <= 0) return sentences[0];
		return (
			[...sentences]
				.reverse()
				.find((sent) => currentTime >= sent.startTimestamp) ?? sentences[0]
		);
	}, [currentTime, sentences]);

	const currentIllustrationUrl = useMemo(() => {
		if (!story) return undefined;
		if (!activeSentence) return story.thumbnailUrl;
		const activeSeq = activeSentence.sequence;
		for (let i = sentences.length - 1; i >= 0; i--) {
			const s = sentences[i];
			if (s.sequence <= activeSeq && s.imageUrl) {
				return s.imageUrl;
			}
		}
		return story.thumbnailUrl;
	}, [activeSentence, sentences, story]);

	if (!story) return <CardShell pending className="h-64" />;

	return (
		<div className="flex min-h-screen flex-col bg-background pb-28">
			<StoryReaderHeader
				story={story}
				mode={mode}
				onModeChange={setMode}
				showTranslation={showTranslation}
				onToggleTranslation={() =>
					setSettings({ showTranslation: !showTranslation })
				}
				saved={saved}
				bookmarkPending={toggleBookmark.isPending}
				bookmarkDisabled={bookmarks.isPending || bookmarks.isError}
				onToggleSave={() => {
					if (toggleBookmark.isPending || !bookmarks.data) return;
					toggleBookmark.mutate(
						{ id: story.id, saved },
						{ onError: () => toast.error(m["stories.bookmarks.error"]()) },
					);
				}}
			/>
			{bookmarks.isError && (
				<CardShell className="p-4">
					<p role="alert">{m["stories.bookmarks.error"]()}</p>
					<SubmitButton
						pending={bookmarks.isFetching}
						onClick={() => void bookmarks.refetch()}
					>
						{m["stories.bookmarks.retry"]()}
					</SubmitButton>
				</CardShell>
			)}

			<main className="mx-auto flex w-full max-w-7xl flex-1 flex-col gap-8 p-4 md:flex-row md:p-6">
				<div className="flex flex-col gap-3 md:w-5/12 lg:w-1/2">
					<div className="overflow-hidden rounded-2xl border-2 border-border/80 bg-muted/40 shadow-sm">
						{currentIllustrationUrl ? (
							<img
								src={currentIllustrationUrl}
								alt={story.title}
								className="aspect-[4/3] w-full object-cover transition-all duration-300"
							/>
						) : null}
					</div>

					<p className="text-[11px] leading-relaxed text-muted-foreground/80 italic">
						{m["stories.reader.disclaimer"]()}
					</p>
				</div>

				<div
					className="flex flex-1 flex-col md:w-7/12 lg:w-1/2"
					style={{
						fontSize: settings.fontSize,
						lineHeight: settings.lineHeight,
						fontFamily:
							settings.fontFamily === "serif"
								? "ui-serif, Georgia, serif"
								: "var(--font-sans), sans-serif",
					}}
				>
					<div className="flex flex-col gap-5 py-1">
						{sentences.map((sent) => {
							const isActive =
								mode === "listen"
									? currentTime >= sent.startTimestamp &&
										currentTime < sent.endTimestamp
									: activeSentence?.id === sent.id;

							return (
								<button
									type="button"
									key={sent.id}
									onClick={() => handleSentenceClick(sent.startTimestamp)}
									className={cn(
										"group relative w-full cursor-pointer rounded-xl p-3 text-left transition-colors",
										isActive
											? "border-l-4 border-primary bg-primary/10 shadow-xs"
											: "hover:bg-muted/50",
									)}
								>
									<div className="flex items-start justify-between gap-3">
										<p
											className={cn(
												"font-medium transition-colors",
												isActive
													? "font-semibold text-foreground"
													: "text-foreground/90",
											)}
										>
											{sent.englishText}
										</p>
										<span
											className="mt-1 flex shrink-0 items-center justify-center rounded-full p-1 text-muted-foreground/60 transition-colors group-hover:text-primary"
											aria-hidden="true"
										>
											<Volume2 className="size-4" />
										</span>
									</div>

									{showTranslation && sent.vietnameseText ? (
										<p
											className="mt-1.5 text-muted-foreground italic"
											style={{ fontSize: settings.fontSize - 2 }}
										>
											{sent.vietnameseText}
										</p>
									) : null}
								</button>
							);
						})}
					</div>
				</div>
			</main>

			{mode === "listen" ? (
				<StoryAudioPlayer
					audioUrl={story.audioUrl}
					fallbackDuration={story.duration}
					currentTime={currentTime}
					onTimeUpdate={setCurrentTime}
					onSeek={setCurrentTime}
				/>
			) : null}
		</div>
	);
}
