import { BookPlus, Copy, FolderPlus, MousePointerClick } from "lucide-react";
import {
	type GuideStep,
	GuideTriggerButton,
	InstructionGuideDialog,
} from "#/components/common/instruction-guide";
import { useInstructionGuide } from "#/hooks/use-instruction-guide";
import { m } from "#/paraglide/messages";

/** localStorage flag behind the dock guide's "don't show again". */
const HIDE_GUIDE_KEY = "engflex:hide-dock-guide";

/**
 * The right-click dock guide. One guide for every dock surface (lesson
 * practice, shadowing, dictation) — dismissing it hides it on all three.
 * The "?" nests in each surface's header row.
 */
export function DockGuide() {
	const guide = useInstructionGuide(HIDE_GUIDE_KEY);

	const steps: GuideStep[] = [
		{
			icon: MousePointerClick,
			title: m["vocabulary.dock.guide.step1.title"](),
			body: m["vocabulary.dock.guide.step1.body"](),
			actionIcon: MousePointerClick,
			actionLabel: m["vocabulary.dock.guide.step1.action"](),
		},
		{
			icon: BookPlus,
			title: m["vocabulary.dock.guide.step2.title"](),
			body: m["vocabulary.dock.guide.step2.body"](),
			actionIcon: BookPlus,
			actionLabel: m["vocabulary.dock.saveWord"](),
		},
		{
			icon: FolderPlus,
			title: m["vocabulary.dock.guide.step3.title"](),
			body: m["vocabulary.dock.guide.step3.body"](),
			actionIcon: FolderPlus,
			actionLabel: m["vocabulary.dock.createDeck"](),
		},
		{
			icon: Copy,
			title: m["vocabulary.dock.guide.step4.title"](),
			body: m["vocabulary.dock.guide.step4.body"](),
			actionIcon: Copy,
			actionLabel: m["vocabulary.dock.copy"](),
		},
	];

	return (
		<>
			<GuideTriggerButton
				onClick={() => guide.setOpen(true)}
				label={m["vocabulary.dock.guide.reopen"]()}
			/>
			<InstructionGuideDialog
				open={guide.open}
				onClose={guide.close}
				eyebrow={m["vocabulary.dock.guide.eyebrow"]()}
				title={m["vocabulary.dock.guide.title"]()}
				subtitle={m["vocabulary.dock.guide.subtitle"]()}
				steps={steps}
				ctaLabel={m["vocabulary.dock.guide.cta"]()}
			/>
		</>
	);
}
