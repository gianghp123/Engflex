import type { StoryDetail } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { cn } from "cn";
import {
	ArrowLeft,
	Bookmark,
	BookOpen,
	Headphones,
	Languages,
} from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { CefrBadge } from "#/components/common/cefr-badge";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";
import { StoryReaderSettings } from "./story-reader-settings";

export interface StoryReaderHeaderItem extends StoryDetail {
	bookmarks?: number;
	saved?: boolean;
}

export function StoryReaderHeader({
	story,
	mode,
	onModeChange,
	showTranslation,
	onToggleTranslation,
	saved = false,
	bookmarkCount = 0,
	onToggleSave,
	bookmarkPending = false,
	bookmarkDisabled = false,
}: {
	story: StoryReaderHeaderItem;
	mode: "listen" | "read";
	onModeChange: (m: "listen" | "read") => void;
	showTranslation: boolean;
	onToggleTranslation: () => void;
	saved?: boolean;
	bookmarkCount?: number;
	onToggleSave?: () => void;
	bookmarkPending?: boolean;
	bookmarkDisabled?: boolean;
}) {
	return (
		<header className="flex flex-wrap items-center justify-between gap-4 border-b border-border bg-card px-4 py-3">
			<div className="flex items-center gap-3">
				<Link
					to={APP_ROUTES.STORIES.LIST}
					className="flex size-9 items-center justify-center rounded-xl border border-border bg-background text-foreground transition-colors hover:bg-muted"
					aria-label={m["stories.reader.settings.back"]()}
				>
					<ArrowLeft className="size-4" />
				</Link>

				<div className="flex flex-col">
					<div className="flex items-center gap-2">
						<h1 className="font-extrabold text-foreground text-lg tracking-tight">
							{story.title}
						</h1>
						{story.cefrLevel ? <CefrBadge value={story.cefrLevel} /> : null}
					</div>
					{story.titleVietnamese ? (
						<p className="text-muted-foreground text-xs font-medium">
							{story.titleVietnamese}
						</p>
					) : null}
				</div>
			</div>

			<div className="flex flex-wrap items-center gap-2">
				<div className="flex items-center rounded-xl border border-border bg-muted/50 p-1">
					<button
						type="button"
						onClick={() => onModeChange("listen")}
						className={cn(
							"flex items-center gap-1.5 rounded-lg px-3 py-1.5 font-bold text-xs transition-colors",
							mode === "listen"
								? "bg-slate-900 text-white shadow-xs dark:bg-white dark:text-slate-900"
								: "text-muted-foreground hover:text-foreground",
						)}
					>
						<Headphones className="size-3.5" />
						<span>{m["stories.reader.listenAndRead"]()}</span>
					</button>

					<button
						type="button"
						onClick={() => onModeChange("read")}
						className={cn(
							"flex items-center gap-1.5 rounded-lg px-3 py-1.5 font-bold text-xs transition-colors",
							mode === "read"
								? "bg-slate-900 text-white shadow-xs dark:bg-white dark:text-slate-900"
								: "text-muted-foreground hover:text-foreground",
						)}
					>
						<BookOpen className="size-3.5" />
						<span>{m["stories.reader.readOnly"]()}</span>
					</button>
				</div>

				<Button
					variant="outline"
					size="sm"
					onClick={onToggleTranslation}
					className={cn(
						"gap-1.5 rounded-xl border-border font-bold text-xs",
						showTranslation
							? "bg-primary/10 text-primary border-primary/30 hover:bg-primary/20"
							: "text-muted-foreground hover:text-foreground",
					)}
				>
					<Languages className="size-3.5" />
					<span>
						{showTranslation
							? m["stories.reader.hideTranslation"]()
							: m["stories.reader.showTranslation"]()}
					</span>
				</Button>

				<StoryReaderSettings />

				{onToggleSave ? (
					<SubmitButton
						pending={bookmarkPending}
						disabled={bookmarkDisabled}
						aria-pressed={saved}
						variant="outline"
						size="sm"
						onClick={onToggleSave}
						className={cn(
							"gap-1.5 rounded-xl border-border font-bold text-xs transition-colors",
							saved
								? "border-primary/40 bg-primary/10 text-primary hover:bg-primary/20"
								: "text-muted-foreground hover:text-foreground",
						)}
					>
						<Bookmark
							className={cn(
								"size-3.5",
								saved ? "fill-primary text-primary" : "",
							)}
						/>
						<span>
							{saved ? m["stories.reader.saved"]() : m["stories.reader.save"]()}
						</span>
						{bookmarkCount > 0 ? (
							<span className="font-semibold">{bookmarkCount}</span>
						) : null}
					</SubmitButton>
				) : null}
			</div>
		</header>
	);
}
