import type { StoryDetail } from "@engflex/contracts";
import { useQueries } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { CardShell } from "#/components/common/card-shell";
import { EmptyList } from "#/components/common/empty-list";
import { SubmitButton } from "#/components/common/submit-button";
import {
	useStoryBookmarks,
	useToggleStoryBookmark,
} from "#/features/bookmarks/queries";
import { storyDetailByIdQueryOptions } from "#/features/stories/queries";
import {
	STORIES_PAGE_SIZE,
	type StoriesSearch,
} from "#/features/stories/types";
import { ApiError } from "#/lib/api";
import { m } from "#/paraglide/messages";
import { StoryCard, StoryCardSkeleton } from "./story-card";

const SAVED_SKELETON_KEYS = Array.from(
	{ length: STORIES_PAGE_SIZE },
	(_, i) => `saved-story-skeleton-${i + 1}`,
);

export function SavedStories({
	filters,
	onReset,
}: {
	filters: StoriesSearch;
	onReset: () => void;
}) {
	const bookmarks = useStoryBookmarks();
	const toggle = useToggleStoryBookmark();
	const [visibleCount, setVisibleCount] = useState(STORIES_PAGE_SIZE);
	const sentinelRef = useRef<HTMLDivElement | null>(null);

	const queries = useQueries({
		queries: (bookmarks.data ?? []).map((bookmark) => ({
			...storyDetailByIdQueryOptions(bookmark.contentId),
			retry: (count: number, error: unknown) =>
				!(
					error instanceof ApiError && [401, 403, 404].includes(error.status)
				) && count < 3,
		})),
	});

	const items = queries
		.map((query) => query.data)
		.filter((story): story is StoryDetail => !!story)
		.filter(
			(story) =>
				(!filters.cefr ||
					filters.cefr === "all" ||
					story.cefrLevel === filters.cefr) &&
				(!filters.topic || story.topic === filters.topic) &&
				(!filters.search ||
					`${story.title} ${story.titleVietnamese}`
						.toLowerCase()
						.includes(filters.search.toLowerCase())),
		);

	const hasMore = visibleCount < items.length;

	useEffect(() => {
		const node = sentinelRef.current;
		if (!node || !hasMore) return;
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) {
					setVisibleCount((prev) =>
						Math.min(items.length, prev + STORIES_PAGE_SIZE),
					);
				}
			},
			{ rootMargin: "200px" },
		);
		observer.observe(node);
		return () => observer.disconnect();
	}, [hasMore, items.length]);

	if (bookmarks.isError)
		return (
			<CardShell className="p-4">
				<p role="alert">{m["stories.bookmarks.error"]()}</p>
				<SubmitButton
					pending={bookmarks.isFetching}
					onClick={() => void bookmarks.refetch()}
				>
					{m["stories.bookmarks.retry"]()}
				</SubmitButton>
			</CardShell>
		);
	if (bookmarks.isPending || queries.some((query) => query.isPending))
		return (
			<div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
				{SAVED_SKELETON_KEYS.map((key) => (
					<StoryCardSkeleton key={key} />
				))}
			</div>
		);
	const errors = queries.filter(
		(query) =>
			query.isError &&
			!(query.error instanceof ApiError && query.error.status === 404),
	);
	if (errors.length)
		return (
			<CardShell className="p-4">
				<p role="alert">{m["stories.bookmarks.error"]()}</p>
				<SubmitButton
					pending={errors.some((query) => query.isFetching)}
					onClick={() => {
						for (const query of errors) void query.refetch();
					}}
				>
					{m["stories.bookmarks.retry"]()}
				</SubmitButton>
			</CardShell>
		);
	const missing = queries.flatMap((query, index) => {
		const bookmark = bookmarks.data?.[index];
		return bookmark &&
			query.error instanceof ApiError &&
			query.error.status === 404
			? [bookmark.contentId]
			: [];
	});

	const visibleItems = items.slice(0, visibleCount);

	return (
		<div className="flex flex-col gap-4">
			{missing.map((id) => (
				<CardShell
					key={id}
					className="flex items-center justify-between gap-3 p-4"
				>
					<p>{m["stories.bookmarks.unavailable"]()}</p>
					<SubmitButton
						pending={toggle.isPending}
						onClick={() =>
							toggle.mutate(
								{ id, saved: true },
								{ onError: () => toast.error(m["stories.bookmarks.error"]()) },
							)
						}
					>
						{m["stories.bookmarks.remove"]()}
					</SubmitButton>
				</CardShell>
			))}
			{items.length === 0 ? (
				<EmptyList title={m["stories.bookmarks.empty"]()} onReset={onReset} />
			) : (
				<>
					<div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
						{visibleItems.map((story) => (
							<StoryCard key={story.id} story={{ ...story, saved: true }} />
						))}
					</div>
					{hasMore ? (
						<div ref={sentinelRef} className="h-4 w-full" aria-hidden="true" />
					) : null}
				</>
			)}
		</div>
	);
}
