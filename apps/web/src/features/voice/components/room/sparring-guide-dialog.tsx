import {
	AudioWaveform,
	MessagesSquare,
	Mic,
	PencilLine,
	Sparkles,
	SquarePen,
} from "lucide-react";
import {
	type GuideStep,
	GuideTriggerButton,
	InstructionGuideDialog,
} from "#/components/common/instruction-guide";
import { m } from "#/paraglide/messages";

/**
 * Sparring guide dialog (mock: "How voice sparring works"). Controlled open
 * state — the room auto-opens it on mount via useSparringGuide and reopens
 * it from the header help button. Dismissible via Esc, the X button, or the
 * CTA; the checkbox persists "don't show again" to localStorage. The shell is
 * the shared instruction-guide dialog.
 */
export function SparringGuideDialog({
	open,
	onClose,
}: {
	open: boolean;
	onClose: (dontShowAgain: boolean) => void;
}) {
	const steps: GuideStep[] = [
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

	return (
		<InstructionGuideDialog
			open={open}
			onClose={onClose}
			eyebrow={m["voice.room.guide.eyebrow"]()}
			title={m["voice.room.guide.title"]()}
			subtitle={m["voice.room.guide.subtitle"]()}
			steps={steps}
			ctaLabel={m["voice.room.guide.start"]()}
		/>
	);
}

/** Header help button reopening the guide. */
export function SparringGuideButton({ onClick }: { onClick: () => void }) {
	return (
		<GuideTriggerButton
			onClick={onClick}
			label={m["voice.room.guide.reopen"]()}
		/>
	);
}
