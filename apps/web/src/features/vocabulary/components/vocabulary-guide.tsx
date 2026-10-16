import { CheckCheck, Layers, Plus, Sparkles } from "lucide-react";
import {
	type GuideStep,
	GuideTriggerButton,
	InstructionGuideDialog,
} from "#/components/common/instruction-guide";
import { useInstructionGuide } from "#/hooks/use-instruction-guide";
import { m } from "#/paraglide/messages";

/** localStorage flag behind the vocabulary guide's "don't show again". */
const HIDE_GUIDE_KEY = "engflex:hide-vocabulary-guide";

/**
 * The vocabulary guide: the add-word / generate / decks / review loop. The
 * "?" next to the add-word control reopens it; `autoOpen` is on for the words
 * hub (where the flow starts) and off for deck pages (trigger only).
 */
export function VocabularyGuide({ autoOpen = false }: { autoOpen?: boolean }) {
	const guide = useInstructionGuide(HIDE_GUIDE_KEY, autoOpen);

	const steps: GuideStep[] = [
		{
			icon: Plus,
			title: m["vocabulary.guide.step1.title"](),
			body: m["vocabulary.guide.step1.body"](),
			actionIcon: Plus,
			actionLabel: m["vocabulary.addWord.button"](),
		},
		{
			icon: Sparkles,
			title: m["vocabulary.guide.step2.title"](),
			body: m["vocabulary.guide.step2.body"](),
			actionIcon: Sparkles,
			actionLabel: m["vocabulary.addWord.generate"](),
		},
		{
			icon: Layers,
			title: m["vocabulary.guide.step3.title"](),
			body: m["vocabulary.guide.step3.body"](),
			actionIcon: Layers,
			actionLabel: m["vocabulary.decks.new"](),
		},
		{
			icon: CheckCheck,
			title: m["vocabulary.guide.step4.title"](),
			body: m["vocabulary.guide.step4.body"](),
			actionIcon: CheckCheck,
			actionLabel: m["vocabulary.card.markMastered"](),
			footer: m["vocabulary.guide.step4.footer"](),
		},
	];

	return (
		<>
			<GuideTriggerButton
				onClick={() => guide.setOpen(true)}
				label={m["vocabulary.guide.reopen"]()}
			/>
			<InstructionGuideDialog
				open={guide.open}
				onClose={guide.close}
				eyebrow={m["vocabulary.guide.eyebrow"]()}
				title={m["vocabulary.guide.title"]()}
				subtitle={m["vocabulary.guide.subtitle"]()}
				steps={steps}
				ctaLabel={m["vocabulary.guide.cta"]()}
			/>
		</>
	);
}
