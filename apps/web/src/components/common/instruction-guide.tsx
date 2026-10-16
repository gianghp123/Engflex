import type { LucideIcon } from "lucide-react";
import { ArrowRight, CircleHelp } from "lucide-react";
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
 * One numbered step card. `body` inlines its action as `%icon%`, rendered as
 * the action's icon chip followed by its bold label (the real control's
 * label, so the two cannot drift).
 */
export interface GuideStep {
	icon: LucideIcon;
	title: string;
	body: string;
	actionIcon: LucideIcon;
	actionLabel: string;
	/** Rendered as the card's ruled footer line; omit for no footer. */
	footer?: string;
}

/**
 * The app's instruction guide shell: eyebrow + title + subtitle, a grid of
 * numbered step cards, and a footer pairing "don't show this tip again" with
 * the CTA. Controlled `open`/`onClose(dontShowAgain)`; callers own the copy
 * and the auto-open hook (useInstructionGuide).
 */
export function InstructionGuideDialog({
	open,
	onClose,
	eyebrow,
	title,
	subtitle,
	steps,
	ctaLabel,
}: {
	open: boolean;
	onClose: (dontShowAgain: boolean) => void;
	eyebrow: string;
	title: string;
	subtitle: string;
	steps: GuideStep[];
	ctaLabel: string;
}) {
	const [dontShowAgain, setDontShowAgain] = useState(false);

	useEffect(() => {
		if (open) {
			setDontShowAgain(false);
		}
	}, [open]);

	const renderBody = (
		body: string,
		ActionIcon: LucideIcon,
		actionLabel: string,
	) => {
		const [before, after] = body.split("%icon%");
		return (
			<>
				{before}
				<span
					className="mx-0.5 inline-flex h-5 w-5 items-center justify-center rounded-md bg-secondary text-secondary-foreground align-text-bottom"
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
							{eyebrow}
						</span>
					</div>
					<DialogTitle className="text-2xl">{title}</DialogTitle>
					<DialogDescription>{subtitle}</DialogDescription>
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
										<div className="tile bg-secondary text-secondary-foreground text-xs font-bold">
											{index + 1}
										</div>
										<h2 className="font-bold text-sm">{title}</h2>
									</div>
									<p className="text-muted-foreground text-xs leading-relaxed">
										{renderBody(body, actionIcon, actionLabel)}
									</p>
								</div>
								{footer ? (
									<div className="mt-4 flex items-center gap-2 border-t pt-3 text-xs">
										<Icon className="text-primary" size={16} />
										<span>{footer}</span>
									</div>
								) : null}
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
						{m["common.guide.dismiss"]()}
					</label>
					<Button type="button" onClick={() => onClose(dontShowAgain)}>
						{ctaLabel}
						<ArrowRight data-icon="inline-end" />
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}

/** Icon button + tooltip reopening a guide from a page header. */
export function GuideTriggerButton({
	onClick,
	label,
}: {
	onClick: () => void;
	label: string;
}) {
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
