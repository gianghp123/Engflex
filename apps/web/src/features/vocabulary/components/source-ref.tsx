import type { VocabularySource } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { cn } from "cn";
import { APP_ROUTES } from "#/app/app-route";
import { useLessonDetail } from "#/features/lessons/queries";
import { useVideo } from "#/features/videos/queries";
import { m } from "#/paraglide/messages";

/** Shared look for every source link; callers may add layout classes. */
const LINK_CLASS =
	"inline-flex max-w-full min-w-0 items-center truncate text-[11px] font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline";

interface RefProps {
	id: string;
	showLabel: boolean;
	className?: string;
}

/**
 * The word's origin as a link — the capture surface's name (lesson/video
 * title), never the raw source id. Resolves through the same cached reads the
 * source's own pages use, so a visited lesson costs no extra request. Renders
 * nothing while unresolved, for `manual`, and for conversation sources —
 * nothing writes those yet and no by-id read carries a display name.
 */
export function SourceRef({
	type,
	id,
	showLabel = true,
	className,
}: {
	type: VocabularySource;
	/** The capture surface's id (user_vocabulary.source_id). */
	id?: string | null;
	/** "<Type>: <title>" for the card line; the hero already shows the pill. */
	showLabel?: boolean;
	className?: string;
}) {
	if (!id || type === "manual" || type === "conversation") return null;
	if (type === "lesson") {
		return <LessonRef id={id} showLabel={showLabel} className={className} />;
	}
	return <VideoRef id={id} showLabel={showLabel} className={className} />;
}

function LessonRef({ id, showLabel, className }: RefProps) {
	const lesson = useLessonDetail(id);
	if (!lesson.data) return null;
	return (
		<Link
			to={APP_ROUTES.LESSONS.DETAIL}
			params={{ slug: lesson.data.slug }}
			title={lesson.data.title}
			className={cn(LINK_CLASS, className)}
		>
			{showLabel ? `${m["vocabulary.source.lesson"]()}: ` : null}
			<span className="truncate">{lesson.data.title}</span>
		</Link>
	);
}

function VideoRef({ id, showLabel, className }: RefProps) {
	const video = useVideo(id);
	if (!video.data) return null;
	return (
		<Link
			to={APP_ROUTES.VIDEOS.SHADOWING}
			params={{ videoId: id }}
			title={video.data.title}
			className={cn(LINK_CLASS, className)}
		>
			{showLabel ? `${m["vocabulary.source.video"]()}: ` : null}
			<span className="truncate">{video.data.title}</span>
		</Link>
	);
}
