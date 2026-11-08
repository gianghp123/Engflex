import type { ContentType } from "@engflex/contracts";
import { Bookmark } from "lucide-react";
import { toast } from "sonner";
import { SubmitButton } from "#/components/common/submit-button";
import { m } from "#/paraglide/messages";
import {
	useContentBookmarks,
	useRemoveSavedItem,
	useSaveContentBookmark,
} from "./queries";

export function BookmarkButton({
	type,
	id,
}: {
	type: ContentType;
	id: string;
}) {
	const bookmarks = useContentBookmarks(type);
	const save = useSaveContentBookmark();
	const remove = useRemoveSavedItem();
	const saved =
		bookmarks.data?.some((bookmark) => bookmark.contentId === id) ?? false;
	const onError = () => toast.error(m["notes.toast.error"]());

	return (
		<SubmitButton
			size="sm"
			variant="outline"
			aria-pressed={saved}
			pending={bookmarks.isPending || save.isPending || remove.isPending}
			onClick={() => {
				if (bookmarks.isError) {
					void bookmarks.refetch();
				} else if (saved) {
					remove.mutate({ type, id }, { onError });
				} else {
					save.mutate({ contentType: type, contentId: id }, { onError });
				}
			}}
		>
			<Bookmark
				data-icon="inline-start"
				fill={saved ? "currentColor" : "none"}
			/>
			{bookmarks.isError
				? m["notes.actions.retry"]()
				: saved
					? m["notes.actions.unsave"]()
					: m["notes.actions.bookmark"]()}
		</SubmitButton>
	);
}
