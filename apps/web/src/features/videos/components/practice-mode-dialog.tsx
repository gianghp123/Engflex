import { Mic, PencilLine } from "lucide-react";
import { MoMascot } from "#/components/common/mo-mascot";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { m } from "#/paraglide/messages";

export type PracticeMode = "dictation" | "shadowing";

/**
 * "Choose a practice mode" dialog opened from a video card. UI only: the
 * chosen mode is reported through `onSelect`, the caller decides what to do.
 */
export function PracticeModeDialog({
	open,
	onOpenChange,
	onSelect,
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onSelect?: (mode: PracticeMode) => void;
}) {
	const modes = [
		{
			mode: "dictation",
			icon: PencilLine,
			variant: "headphones",
			label: m["videos.modeDialog.mode.dictation"](),
		},
		{
			mode: "shadowing",
			icon: Mic,
			variant: "talk",
			label: m["videos.modeDialog.mode.shadowing"](),
		},
	] as const;

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="gap-6 p-6 sm:max-w-xl sm:p-8">
				<DialogHeader className="items-center text-center sm:text-center">
					<DialogTitle className="text-2xl font-bold">
						{m["videos.modeDialog.title"]()}
					</DialogTitle>
					<DialogDescription>{m["videos.modeDialog.body"]()}</DialogDescription>
				</DialogHeader>
				<div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
					{modes.map(({ mode, icon: Icon, variant, label }) => (
						<button
							key={mode}
							type="button"
							onClick={() => onSelect?.(mode)}
							className="surface-card-interactive flex cursor-pointer flex-col items-center gap-5 px-5 pt-7 pb-6"
						>
							<span className="flex size-32 items-center justify-center rounded-full bg-primary/10">
								<MoMascot variant={variant} size={92} />
							</span>
							<span className="flex items-center justify-center gap-2 text-center text-sm font-extrabold tracking-wide text-foreground uppercase">
								<Icon aria-hidden="true" className="size-5 shrink-0" />
								{label}
							</span>
						</button>
					))}
				</div>
			</DialogContent>
		</Dialog>
	);
}
