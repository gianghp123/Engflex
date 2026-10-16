import {
	ArrowRight,
	AudioWaveform,
	CircleHelp,
	MessagesSquare,
	Mic,
	PencilLine,
	Sparkles,
	SquarePen,
} from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#/components/ui/tooltip";
import { m } from "#/paraglide/messages";

/**
 * Sparring guide dialog (mock: "How voice sparring works"). Controlled open
 * state — the room auto-opens it on mount via useSparringGuide and reopens
 * it from the header help button. Dismissible via Esc, the X button, or the
 * CTA; the checkbox persists "don't show again" to localStorage.
 */
export function SparringGuideDialog({
	open,
	onClose,
}: {
	open: boolean;
	onClose: (dontShowAgain: boolean) => void;
}) {
	const [dontShowAgain, setDontShowAgain] = useState(false);

	useEffect(() => {
		if (open) {
			setDontShowAgain(false);
		}
	}, [open]);

	const steps = [
		{
			icon: AudioWaveform,
			title: m["voice.room.guide.step1.title"](),
			body: m["voice.room.guide.step1.body"](),
			actionIcon: Mic,
			actionLabel: m["voice.room.guide.step1.action"](),
			footer: m["voice.room.guide.step1.footer"](),
		},
		{
			icon: MessagesSquare,
			title: m["voice.room.guide.step2.title"](),
			body: m["voice.room.guide.step2.body"](),
			actionIcon: Sparkles,
			actionLabel: m["voice.room.transcript.analyze"](),
			footer: m["voice.room.guide.step2.footer"](),
		},
		{
			icon: SquarePen,
			title: m["voice.room.guide.step3.title"](),
			body: m["voice.room.guide.step3.body"](),
			actionIcon: PencilLine,
			actionLabel: m["voice.room.recovery.reviewTranscript"](),
			footer: m["voice.room.guide.step3.footer"](),
		},
		{
			icon: Sparkles,
			title: m["voice.room.guide.step4.title"](),
			body: m["voice.room.guide.step4.body"](),
			actionIcon: Sparkles,
			actionLabel: m["voice.room.transcript.reanalyze"](),
			footer: m["voice.room.guide.step4.footer"](),
		},
	];

	const renderBody = (
		body: string,
		ActionIcon: typeof Mic,
		actionLabel: string,
	) => {
		const [before, after] = body.split("%icon%");
		return (
			<>
				{before}
				<span
					className="mx-0.5 inline-flex h-5 w-5 items-center justify-center rounded-md bg-secondary text-primary align-text-bottom"
					aria-hidden="true"
				>
					<ActionIcon size={12} />
				</span>{" "}
				<strong className="text-foreground font-semibold">{actionLabel}</strong>
				{after ?? ""}
			</>
		);
	};

	return (
		<Dialog
			open={open}
			onOpenChange={(next) => {
				if (!next) {
					onClose(dontShowAgain);
				}
			}}
		>
			<DialogContent className="max-h-[calc(100dvh-2rem)] overflow-y-auto sm:max-w-3xl">
				<DialogHeader>
					<div className="flex items-center gap-2">
						<span className="h-2 w-2 rounded-full bg-primary" />
						<span className="text-primary text-xs font-semibold uppercase tracking-wide">
							{m["voice.room.guide.eyebrow"]()}
						</span>
					</div>
					<DialogTitle className="text-2xl">
						{m["voice.room.guide.title"]()}
					</DialogTitle>
					<DialogDescription>
						{m["voice.room.guide.subtitle"]()}
					</DialogDescription>
				</DialogHeader>
				<div className="grid grid-cols-1 gap-4 bg-muted/50 p-6 md:grid-cols-2 md:p-8">
					{steps.map(
						(
							{ icon: Icon, title, body, actionIcon, actionLabel, footer },
							index,
						) => (
							<div
								key={title}
								className="surface-card flex flex-col justify-between p-5"
							>
								<div>
									<div className="mb-3 flex items-center gap-2.5">
										<div className="tile bg-secondary text-primary text-xs font-bold">
											{index + 1}
										</div>
										<h2 className="font-bold text-sm">{title}</h2>
									</div>
									<p className="text-muted-foreground text-xs leading-relaxed">
										{renderBody(body, actionIcon, actionLabel)}
									</p>
								</div>
								<div className="mt-4 flex items-center gap-2 border-t pt-3 text-xs">
									<Icon className="text-primary" size={16} />
									<span>{footer}</span>
								</div>
							</div>
						),
					)}
				</div>
				<DialogFooter className="flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
					<label className="text-muted-foreground flex cursor-pointer select-none items-center gap-2.5 text-sm">
						<input
							type="checkbox"
							checked={dontShowAgain}
							onChange={(event) => setDontShowAgain(event.target.checked)}
							className="accent-primary h-4 w-4 rounded"
						/>
						{m["voice.room.guide.dismiss"]()}
					</label>
					<Button type="button" onClick={() => onClose(dontShowAgain)}>
						{m["voice.room.guide.start"]()}
						<ArrowRight data-icon="inline-end" />
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}

/** Header help button reopening the guide. */
export function SparringGuideButton({ onClick }: { onClick: () => void }) {
	const label = m["voice.room.guide.reopen"]();
	return (
		<Tooltip>
			<TooltipTrigger asChild>
				<Button
					type="button"
					variant="ghost"
					size="icon"
					onClick={onClick}
					aria-label={label}
				>
					<CircleHelp />
				</Button>
			</TooltipTrigger>
			<TooltipContent>{label}</TooltipContent>
		</Tooltip>
	);
}
