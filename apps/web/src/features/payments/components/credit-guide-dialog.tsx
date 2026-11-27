import type { AccessAction, ActionCost } from "@engflex/contracts";
import { CalendarClock, Mic, ScanSearch, Sparkles, Wallet } from "lucide-react";
import {
	type GuideStep,
	GuideTriggerButton,
	InstructionGuideDialog,
} from "#/components/common/instruction-guide";
import { m } from "#/paraglide/messages";

function costFor(costs: ActionCost[], action: AccessAction): number {
	return costs.find((cost) => cost.action === action)?.credits ?? 0;
}

/**
 * Credit guide (mock: "How AI credits work"): the same instruction-guide shell
 * as the voice sparring guide, with live action costs and the AI gradient on
 * the credit-specific headings/action labels. Auto-open is owned by the usage
 * page; the dialog itself is controlled.
 */
export function CreditGuideDialog({
	open,
	onClose,
	costs,
}: {
	open: boolean;
	onClose: (dontShowAgain: boolean) => void;
	costs: ActionCost[];
}) {
	const voice = costFor(costs, "voice_session_start");
	const review = costFor(costs, "exam_ai_review");
	const plan = costFor(costs, "study_plan");

	const steps: GuideStep[] = [
		{
			icon: Sparkles,
			title: m["payments.creditGuide.step1.title"](),
			body: m["payments.creditGuide.step1.body"](),
			actionIcon: Sparkles,
			actionLabel: m["payments.creditGuide.step1.action"](),
			footer: m["payments.creditGuide.step1.footer"](),
		},
		{
			icon: Mic,
			title: m["payments.creditGuide.step2.title"]({ credits: voice }),
			body: m["payments.creditGuide.step2.body"]({ credits: voice }),
			actionIcon: Mic,
			actionLabel: m["payments.creditGuide.step2.action"](),
			footer: m["payments.creditGuide.step2.footer"](),
		},
		{
			icon: ScanSearch,
			title: m["payments.creditGuide.step3.title"]({ review, plan }),
			body: m["payments.creditGuide.step3.body"]({ review, plan }),
			actionIcon: ScanSearch,
			actionLabel: m["payments.creditGuide.step3.action"](),
			footer: m["payments.creditGuide.step3.footer"](),
		},
		{
			icon: CalendarClock,
			title: m["payments.creditGuide.step4.title"](),
			body: m["payments.creditGuide.step4.body"](),
			actionIcon: Wallet,
			actionLabel: m["payments.creditGuide.step4.action"](),
			footer: m["payments.creditGuide.step4.footer"](),
		},
	];

	return (
		<InstructionGuideDialog
			open={open}
			onClose={onClose}
			eyebrow={m["payments.creditGuide.eyebrow"]()}
			title={m["payments.creditGuide.title"]()}
			subtitle={m["payments.creditGuide.subtitle"]()}
			steps={steps}
			ctaLabel={m["payments.creditGuide.cta"]()}
			titleClassName="text-ai-gradient"
			eyebrowClassName="text-ai-gradient"
			stepTitleClassName="text-ai-gradient"
			actionLabelClassName="text-ai-gradient"
		/>
	);
}

/** Header help button reopening the credit guide. */
export function CreditGuideButton({ onClick }: { onClick: () => void }) {
	return (
		<GuideTriggerButton
			onClick={onClick}
			label={m["payments.creditGuide.reopen"]()}
		/>
	);
}
