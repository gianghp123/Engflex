import { ArrowLeft, RotateCcw } from "lucide-react";
import { MoMascot } from "#/components/common/mo-mascot";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { m } from "#/paraglide/messages";

export interface DictationSummaryDialogProps {
	open: boolean;
	accuracy: number;
	correctWords: number;
	incorrectWords: number;
	onRestart: () => void;
	onBackToHub: () => void;
}

export function DictationSummaryDialog({
	open,
	accuracy,
	correctWords,
	incorrectWords,
	onRestart,
	onBackToHub,
}: DictationSummaryDialogProps) {
	return (
		<Dialog open={open}>
			<DialogContent className="gap-6 p-6 sm:max-w-md sm:p-8">
				<DialogHeader className="items-center text-center sm:text-center">
					<div className="flex size-24 items-center justify-center rounded-full bg-primary/10">
						<MoMascot variant="win" size={80} />
					</div>
					<DialogTitle className="text-2xl font-bold">
						{m["videos.dictation.summary.title"]()}
					</DialogTitle>
					<DialogDescription>
						{m["videos.dictation.summary.body"]()}
					</DialogDescription>
				</DialogHeader>

				<div className="surface-card flex flex-col gap-3 rounded-2xl border-2 border-border p-4 bg-muted/40">
					<div className="flex items-center justify-between font-bold text-base">
						<span className="text-foreground">
							{m["videos.dictation.summary.accuracy"]({
								percent: Math.round(accuracy),
							})}
						</span>
					</div>
					<div className="grid grid-cols-2 gap-2 text-sm">
						<div className="flex items-center gap-1.5 font-semibold text-accuracy">
							<span>
								{m["videos.dictation.summary.correctWords"]({
									count: correctWords,
								})}
							</span>
						</div>
						<div className="flex items-center gap-1.5 font-semibold text-ai-coral">
							<span>
								{m["videos.dictation.summary.incorrectWords"]({
									count: incorrectWords,
								})}
							</span>
						</div>
					</div>
				</div>

				<div className="flex flex-col gap-2.5">
					<Button
						type="button"
						onClick={onRestart}
						className="btn btn-primary w-full uppercase"
					>
						<RotateCcw data-icon="inline-start" />
						{m["videos.dictation.summary.restart"]()}
					</Button>
					<Button
						type="button"
						variant="outline"
						onClick={onBackToHub}
						className="btn btn-outline w-full uppercase"
					>
						<ArrowLeft data-icon="inline-start" />
						{m["videos.dictation.summary.backToHub"]()}
					</Button>
				</div>
			</DialogContent>
		</Dialog>
	);
}
