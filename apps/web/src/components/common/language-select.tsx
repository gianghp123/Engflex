import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { getLocale, locales } from "#/paraglide/runtime";

/** The platform's name for a language, in the learner's current UI locale. */
function languageLabel(code: string): string {
	return (
		new Intl.DisplayNames([getLocale(), "en"], { type: "language" }).of(code) ??
		code
	);
}

/**
 * The explanation-language control shared by every LLM surface. Options derive
 * from the app's configured locales — a locale added in project.inlang shows
 * up here with no code change — and the value is the BCP-47 code the API
 * resolves to a language name. Controlled: the caller owns the value so its
 * request can carry it.
 */
export function LanguageSelect({
	id,
	value,
	onValueChange,
}: {
	id?: string;
	value: string;
	onValueChange: (code: string) => void;
}) {
	return (
		<Select value={value} onValueChange={onValueChange}>
			<SelectTrigger id={id} className="w-full">
				<SelectValue />
			</SelectTrigger>
			<SelectContent>
				{locales.map((locale) => (
					<SelectItem key={locale} value={locale}>
						{languageLabel(locale)}
					</SelectItem>
				))}
			</SelectContent>
		</Select>
	);
}
