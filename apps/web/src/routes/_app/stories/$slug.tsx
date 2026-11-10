import type { StorySentence } from "@engflex/contracts";
import { createFileRoute } from "@tanstack/react-router";
import { cn } from "cn";
import { Volume2 } from "lucide-react";
import { useEffect, useMemo, useRef, useState } from "react";
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
	const transcriptContainerRef = useRef<HTMLDivElement | null>(null);

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

	const activeSentenceId = activeSentence?.id;

	useEffect(() => {
		if (mode !== "listen" || !activeSentenceId) return;
		const container = transcriptContainerRef.current;
		if (!container) return;
		const activeEl = container.querySelector<HTMLElement>(
			`[data-sentence-id="${activeSentenceId}"]`,
		);
		if (!activeEl) return;
		activeEl.scrollIntoView({ behavior: "smooth", block: "nearest" });
	}, [activeSentenceId, mode]);

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
		<div className="flex min-h-screen flex-col bg-background md:h-[calc(100dvh-3.5rem)] md:min-h-0 md:overflow-hidden">
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
				<CardShell className="m-4 shrink-0 p-4">
					<p role="alert">{m["stories.bookmarks.error"]()}</p>
					<SubmitButton
						pending={bookmarks.isFetching}
						onClick={() => void bookmarks.refetch()}
					>
						{m["stories.bookmarks.retry"]()}
					</SubmitButton>
				</CardShell>
			)}

			<main
				className={cn(
					"mx-auto flex min-h-0 w-full max-w-7xl flex-1 flex-col gap-6 p-4 md:flex-row md:gap-8 md:overflow-hidden md:p-6",
					mode === "listen" ? "pb-28 md:pb-24" : "",
				)}
			>
				<div className="flex shrink-0 flex-col gap-3 md:h-full md:w-5/12 md:overflow-hidden lg:w-1/2">
					<div className="min-h-0 overflow-hidden rounded-2xl border-2 border-border/80 bg-muted/40 shadow-sm">
						{currentIllustrationUrl ? (
							<img
								src={currentIllustrationUrl}
								alt={story.title}
								className="aspect-[4/3] max-h-[calc(100dvh-16rem)] w-full object-cover transition-all duration-300"
							/>
						) : null}
					</div>

					<p className="shrink-0 text-[11px] leading-relaxed text-muted-foreground/80 italic">
						{m["stories.reader.disclaimer"]()}
					</p>
				</div>

				<div
					ref={transcriptContainerRef}
					className="flex min-h-0 flex-1 flex-col overflow-y-auto pr-1 md:h-full md:w-7/12 md:pr-2 lg:w-1/2"
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
									data-sentence-id={sent.id}
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
