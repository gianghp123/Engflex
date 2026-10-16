import { Copy, FolderPlus, NotebookPen } from "lucide-react";
import { type ReactNode, useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import {
	ContextMenu,
	ContextMenuContent,
	ContextMenuItem,
	ContextMenuSeparator,
	ContextMenuTrigger,
} from "#/components/ui/context-menu";
import { AddWordDialog } from "#/features/vocabulary/components/add-word-dialog";
import { NewDeckDialog } from "#/features/vocabulary/components/new-deck-dialog";
import {
	readSelection,
	selectionContextSentence,
} from "#/features/vocabulary/selection";
import type { AddWordSource } from "#/features/vocabulary/types";
import { m } from "#/paraglide/messages";

/**
 * The right-click dock for exercise content. Wraps a region; right-clicking
 * inside it shows Save word / Create deck / Copy instead of the browser's
 * native menu — so Copy lives here too. The selection is captured when the
 * menu opens (focus moves later); Save word and Copy stay disabled without
 * one. The wrapper is `contents`, so it adds no box to the layout.
 */
export function TextActionDock({
	source,
	children,
}: {
	/** Where the save was started from; rides the API payload. */
	source: AddWordSource;
	children: ReactNode;
}) {
	const regionRef = useRef<HTMLDivElement>(null);
	const [captured, setCaptured] = useState<{
		term: string;
		context: string | null;
	}>({ term: "", context: null });
	const [savedOpen, setSavedOpen] = useState(false);
	const [deckOpen, setDeckOpen] = useState(false);

	// Text controls keep the browser's own menu: the dock's selection-based
	// actions cannot see input/textarea selections, and paste/cut/spellcheck
	// live there. Stopping the capture-phase event keeps Radix from opening
	// (its open handler is a React bubble-phase listener, dispatched at the
	// root) while the browser's default menu still appears — stopPropagation
	// cancels dispatch, not the default action.
	useEffect(() => {
		const region = regionRef.current;
		if (!region) return;
		const passThrough = (event: MouseEvent) => {
			const target = event.target;
			if (!(target instanceof HTMLElement)) return;
			if (target.closest("input, textarea, [contenteditable]")) {
				event.stopPropagation();
			}
		};
		region.addEventListener("contextmenu", passThrough, { capture: true });
		return () =>
			region.removeEventListener("contextmenu", passThrough, {
				capture: true,
			});
	}, []);

	function capture() {
		const region = regionRef.current;
		if (!region) {
			setCaptured({ term: "", context: null });
			return;
		}
		setCaptured({
			term: readSelection(region),
			context: selectionContextSentence(region),
		});
	}

	async function copySelection() {
		try {
			await navigator.clipboard.writeText(captured.term);
			toast.success(m["vocabulary.dock.copied"]());
		} catch {
			toast.error(m["vocabulary.dock.copyFailed"]());
		}
	}

	const hasSelection = captured.term !== "";

	return (
		<>
			<ContextMenu
				onOpenChange={(open) => {
					if (open) capture();
				}}
			>
				<ContextMenuTrigger asChild>
					<div ref={regionRef} className="contents">
						{children}
					</div>
				</ContextMenuTrigger>
				<ContextMenuContent>
					<ContextMenuItem
						disabled={!hasSelection}
						onSelect={() => setSavedOpen(true)}
					>
						<NotebookPen data-icon="inline-start" />
						{m["vocabulary.dock.saveWord"]()}
					</ContextMenuItem>
					<ContextMenuItem onSelect={() => setDeckOpen(true)}>
						<FolderPlus data-icon="inline-start" />
						{m["vocabulary.dock.createDeck"]()}
					</ContextMenuItem>
					<ContextMenuSeparator />
					<ContextMenuItem
						disabled={!hasSelection}
						onSelect={() => void copySelection()}
					>
						<Copy data-icon="inline-start" />
						{m["vocabulary.dock.copy"]()}
					</ContextMenuItem>
				</ContextMenuContent>
			</ContextMenu>
			<AddWordDialog
				open={savedOpen}
				onOpenChange={setSavedOpen}
				source={source}
				initialTerm={captured.term}
				initialContext={captured.context ?? undefined}
			/>
			<NewDeckDialog open={deckOpen} onOpenChange={setDeckOpen} />
		</>
	);
}
