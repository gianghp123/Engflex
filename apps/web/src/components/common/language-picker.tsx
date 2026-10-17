import { Languages } from "lucide-react";
import { useState } from "react";
import { LanguageSelect } from "#/components/common/language-select";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { FieldDescription, FieldLabel } from "#/components/ui/field";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "#/components/ui/popover";
import { Switch } from "#/components/ui/switch";
import { explanationLanguage } from "#/lib/language/explanation-language";
import { m } from "#/paraglide/messages";

/**
 * The explanation-language picker every LLM surface mounts: a compact trigger
 * that opens the shared LanguageSelect, with a Don't-ask-again switch so the
 * first-use ask (see useExplanationLanguage) never nags twice.
 *
 * `open` is controlled by that hook rather than owned here, because the ask is
 * driven by the learner's own click — never by an effect on mount.
 */
export function LanguagePicker({
	open,
	onOpenChange,
	onConfirm,
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** `askAgain` is the learner's answer to "Don't ask again", inverted. */
	onConfirm: (code: string, askAgain: boolean) => void;
}) {
	return (
		<Popover open={open} onOpenChange={onOpenChange}>
			<PopoverTrigger asChild>
				<Button
					type="button"
					variant="secondary"
					size="icon-sm"
					aria-label={m["common.language.trigger"]()}
				>
					<Languages />
				</Button>
			</PopoverTrigger>
			<PopoverContent align="end" className="w-72">
				{/* key resets the draft on every open — the house pattern for
				    resetting state on an entity change, instead of an effect. */}
				<LanguagePickerForm
					key={open ? "open" : "closed"}
					onConfirm={onConfirm}
				/>
			</PopoverContent>
		</Popover>
	);
}

function LanguagePickerForm({
	onConfirm,
}: {
	onConfirm: (code: string, askAgain: boolean) => void;
}) {
	const [code, setCode] = useState(explanationLanguage);
	const [dontAsk, setDontAsk] = useState(true);

	return (
		<div className="flex flex-col gap-3">
			<div className="flex flex-col gap-2">
				<FieldLabel htmlFor="explanation-language">
					{m["common.language.label"]()}
				</FieldLabel>
				<LanguageSelect
					id="explanation-language"
					value={code}
					onValueChange={setCode}
				/>
				<FieldDescription>{m["common.language.hint"]()}</FieldDescription>
			</div>
			<label
				htmlFor="explanation-language-dont-ask"
				className="flex items-center justify-between gap-2 text-xs font-medium"
			>
				{m["common.language.dontAskAgain"]()}
				<Switch
					id="explanation-language-dont-ask"
					checked={dontAsk}
					onCheckedChange={setDontAsk}
				/>
			</label>
			<SubmitButton
				type="button"
				size="sm"
				onClick={() => onConfirm(code, !dontAsk)}
			>
				{m["common.actions.save"]()}
			</SubmitButton>
		</div>
	);
}
