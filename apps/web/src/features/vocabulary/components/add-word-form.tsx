import type {
	CEFR,
	VocabularyDetails,
	VocabularyPartOfSpeech,
} from "@engflex/contracts";
import {
	CEFR_VALUES,
	MaxDeckWords,
	MaxUserWords,
	VOCABULARY_PART_OF_SPEECH_VALUES,
} from "@engflex/contracts";
import { ChevronDown, Sparkles } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import {
	Collapsible,
	CollapsibleContent,
	CollapsibleTrigger,
} from "#/components/ui/collapsible";
import { DialogDescription, DialogFooter } from "#/components/ui/dialog";
import {
	Field,
	FieldError,
	FieldGroup,
	FieldLabel,
} from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "#/components/ui/popover";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { Textarea } from "#/components/ui/textarea";
import { POS_LABELS } from "#/features/vocabulary/components/pos-label";
import {
	useAddWord,
	useDecks,
	useGenerateWordDraft,
	useLibrarySearch,
	useVocabularyStats,
} from "#/features/vocabulary/queries";
import { termKey } from "#/features/vocabulary/term-key";
import type { AddWordSource } from "#/features/vocabulary/types";
import { ApiError } from "#/lib/api";
import { enumOptions } from "#/lib/enum-options";
import { m } from "#/paraglide/messages";
import { getLocale, locales } from "#/paraglide/runtime";

const CEFR_OPTIONS = enumOptions(CEFR_VALUES);
const POS_OPTIONS = enumOptions(VOCABULARY_PART_OF_SPEECH_VALUES, POS_LABELS);

/** The last generate output language, remembered per browser. */
const GENERATE_LANGUAGE_KEY = "engflex:vocabulary-generate-language";
const DEFAULT_GENERATE_LANGUAGE = "en";

/** The stored pick when it is still one of the app's locales, else the default. */
function initialGenerateLanguage(): string {
	if (typeof window === "undefined") return DEFAULT_GENERATE_LANGUAGE;
	try {
		const stored = window.localStorage.getItem(GENERATE_LANGUAGE_KEY);
		return stored && (locales as readonly string[]).includes(stored)
			? stored
			: DEFAULT_GENERATE_LANGUAGE;
	} catch {
		return DEFAULT_GENERATE_LANGUAGE;
	}
}

/**
 * Language options derive from the app's configured locales — a locale added
 * in project.inlang shows up here with no code change; the label is the
 * platform's language name in the current UI locale.
 */
function languageLabel(locale: string): string {
	return (
		new Intl.DisplayNames([getLocale(), "en"], { type: "language" }).of(
			locale,
		) ?? locale
	);
}

const EMPTY_FORM: {
	term: string;
	cefr: CEFR | "";
	partOfSpeech: VocabularyPartOfSpeech;
	definition: string;
	example: string;
	details: VocabularyDetails | null;
} = {
	term: "",
	cefr: "",
	partOfSpeech: "noun",
	definition: "",
	example: "",
	details: null,
};

/**
 * The "add a word" mode: a custom word, saved to the reader's own list and
 * enrolled into `deckId` when the flow was opened from a deck — or into the
 * deck picked from the select when there is no deck in context.
 *
 * Opened from the context dock, the term and its sentence arrive as prefill
 * and the capture surface rides along as the enrolment source.
 *
 * A typed term that already exists in the library is called out, because saving
 * it reuses that entry instead of creating a second one. The check is a
 * debounced server search over the whole library (term or definition), so it
 * no longer depends on how many words a page holds. The system/global
 * dictionary half is still deferred (it needs `GET /vocabulary-items?search=`)
 * and is reported as no match rather than faked.
 *
 * Product caps ride the same read: at `MaxUserWords` a new save is blocked
 * (a save that reuses an existing entry stays allowed), and a full target deck
 * (`MaxDeckWords`, the `deckFull` prop) blocks before the server's 422. The
 * server remains the source of truth either way.
 */
export function AddWordForm({
	deckId,
	deckFull = false,
	onClose,
	initialTerm,
	initialContext,
	source,
}: {
	deckId?: string;
	/** The target deck already holds its cap of words (the caller knows its
	 *  size); saving into it is blocked before the server has to refuse. */
	deckFull?: boolean;
	onClose: () => void;
	/** Prefill from the context dock: the selection and its sentence. */
	initialTerm?: string;
	initialContext?: string;
	source?: AddWordSource;
}) {
	const addWord = useAddWord();
	const stats = useVocabularyStats();
	const decks = useDecks(0);
	const [form, setForm] = useState(() => ({
		...EMPTY_FORM,
		term: initialTerm ?? "",
		example: initialContext ?? "",
	}));
	const [deckChoice, setDeckChoice] = useState("none");
	const [error, setError] = useState<string | null>(null);
	const generate = useGenerateWordDraft();
	const [generateError, setGenerateError] = useState<string | null>(null);
	const [generateLanguage, setGenerateLanguage] = useState(
		initialGenerateLanguage,
	);
	const [configOpen, setConfigOpen] = useState(false);

	/** On demand: fills the editable fields; nothing applies or saves itself. */
	function generateDraft() {
		const term = form.term.trim();
		if (!term) return;
		setGenerateError(null);
		generate.mutate(
			{ term, context: form.example.trim(), outputLanguage: generateLanguage },
			{
				onSuccess: (draft) => {
					setForm((current) => ({
						...current,
						definition: draft.definition,
						cefr: draft.cefr as CEFR,
						partOfSpeech: (draft.partOfSpeech ||
							current.partOfSpeech) as VocabularyPartOfSpeech,
						details: draft.details,
					}));
				},
				onError: () =>
					setGenerateError(m["vocabulary.addWord.generateFailed"]()),
			},
		);
	}

	/** Remember the pick for the next draft (client-only preference). */
	function chooseLanguage(value: string) {
		setGenerateLanguage(value);
		try {
			window.localStorage.setItem(GENERATE_LANGUAGE_KEY, value);
		} catch {
			// Private mode etc: the pick still applies to this session.
		}
	}

	/** The config's confirm: take the pick, close the config, generate. */
	function confirmGenerate() {
		setConfigOpen(false);
		generateDraft();
	}

	const typedKey = termKey(form.term);
	const typedTerm = form.term.trim();
	// Debounced needle for the library check: typing alone never queries, and
	// the check never depends on which page of the library is loaded.
	const [checkedTerm, setCheckedTerm] = useState("");
	useEffect(() => {
		const timer = window.setTimeout(() => setCheckedTerm(typedTerm), 300);
		return () => window.clearTimeout(timer);
	}, [typedTerm]);
	const check = useLibrarySearch(checkedTerm || null);
	const existingInLibrary =
		typedKey.length > 0
			? (check.data?.items ?? []).some(
					(item) => termKey(item.term) === typedKey,
				)
			: false;
	const atWordLimit = (stats.data?.totalSaved ?? 0) >= MaxUserWords;
	const wordLimitBlocked = atWordLimit && !existingInLibrary;
	const saveBlocked = wordLimitBlocked || deckFull;

	function submit() {
		const term = form.term.trim();
		const definition = form.definition.trim();
		const cefr = form.cefr;
		if (!term || !definition || !cefr) {
			setError(m["vocabulary.addWord.required"]());
			return;
		}
		setError(null);
		const targetDeck =
			deckId ?? (deckChoice === "none" ? undefined : deckChoice);
		const input = {
			term,
			cefr,
			partOfSpeech: form.partOfSpeech,
			definition,
			example: form.example.trim(),
			details: form.details ?? undefined,
			sourceType: source?.type,
			sourceId: source?.id,
		};
		addWord.mutate(
			{ ...input, deckId: targetDeck },
			{
				onSuccess: () => {
					toast.success(m["common.toast.addedToVocabulary"]());
					setForm(EMPTY_FORM);
					setDeckChoice("none");
					setError(null);
					onClose();
				},
				onError: (err) => {
					// 422 is the server's product-cap refusal: the word list or the
					// target deck is full (the client prechecks usually beat it).
					if (err instanceof ApiError && err.status === 422) {
						setError(
							deckId
								? m["vocabulary.deck.limitReached"]({ max: MaxDeckWords })
								: m["vocabulary.addWord.limitReached"]({ max: MaxUserWords }),
						);
						return;
					}
					toast.error(m["vocabulary.addWord.failed"]());
				},
			},
		);
	}

	return (
		<>
			<div className="flex items-start justify-between gap-3">
				<DialogDescription className="font-medium">
					{m["vocabulary.addWord.description"]()}
				</DialogDescription>
				<Popover open={configOpen} onOpenChange={setConfigOpen}>
					<PopoverTrigger asChild>
						<SubmitButton
							type="button"
							size="sm"
							variant="outline"
							pending={generate.isPending}
							// The save reuses an existing entry and never updates it, so
							// a draft generated for an already-saved word would be
							// silently discarded — the hint above the definition field
							// explains the disabled state.
							disabled={!form.term.trim() || existingInLibrary}
						>
							<Sparkles
								data-icon="inline-start"
								className="text-violet-500 dark:text-violet-400"
							/>
							{m["vocabulary.addWord.generate"]()}
						</SubmitButton>
					</PopoverTrigger>
					<PopoverContent align="end" className="flex flex-col gap-3">
						<div className="flex flex-col gap-2">
							<FieldLabel htmlFor="generate-language">
								{m["vocabulary.addWord.outputLanguage"]()}
							</FieldLabel>
							<Select value={generateLanguage} onValueChange={chooseLanguage}>
								<SelectTrigger id="generate-language" className="w-full">
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
							<p className="text-xs text-muted-foreground">
								{m["vocabulary.addWord.outputLanguageHint"]()}
							</p>
						</div>
						<SubmitButton
							type="button"
							size="sm"
							pending={generate.isPending}
							onClick={confirmGenerate}
						>
							<Sparkles
								data-icon="inline-start"
								className="text-violet-500 dark:text-violet-400"
							/>
							{m["vocabulary.addWord.generate"]()}
						</SubmitButton>
					</PopoverContent>
				</Popover>
			</div>
			{generateError ? (
				<p className="text-xs font-bold text-destructive">{generateError}</p>
			) : null}
			<FieldGroup>
				<Field data-invalid={Boolean(error)}>
					<FieldLabel htmlFor="word-term">
						{m["vocabulary.addWord.termLabel"]()}
					</FieldLabel>
					<Input
						id="word-term"
						value={form.term}
						onChange={(event) =>
							setForm((current) => ({ ...current, term: event.target.value }))
						}
					/>
					{existingInLibrary ? (
						<p className="text-xs font-bold text-muted-foreground">
							{m["vocabulary.addWord.alreadyInLibrary"]()}
						</p>
					) : null}
					{wordLimitBlocked ? (
						<p className="text-xs font-bold text-muted-foreground">
							{m["vocabulary.addWord.limitReached"]({ max: MaxUserWords })}
						</p>
					) : null}
				</Field>
				<div className="grid gap-3 sm:grid-cols-2">
					<Field>
						<FieldLabel>{m["vocabulary.addWord.cefrLabel"]()}</FieldLabel>
						<Select
							value={form.cefr}
							onValueChange={(value) =>
								setForm((current) => ({ ...current, cefr: value as CEFR }))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue
									placeholder={m["vocabulary.addWord.cefrPlaceholder"]()}
								/>
							</SelectTrigger>
							<SelectContent>
								{CEFR_OPTIONS.map((option) => (
									<SelectItem key={option.value} value={option.value}>
										{option.label()}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
					<Field>
						<FieldLabel>{m["vocabulary.addWord.posLabel"]()}</FieldLabel>
						<Select
							value={form.partOfSpeech}
							onValueChange={(value) =>
								setForm((current) => ({
									...current,
									partOfSpeech: value as VocabularyPartOfSpeech,
								}))
							}
						>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								{POS_OPTIONS.map((option) => (
									<SelectItem key={option.value} value={option.value}>
										{option.label()}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
				</div>
				{deckId ? null : (
					<Field>
						<FieldLabel>{m["vocabulary.addWord.deckLabel"]()}</FieldLabel>
						<Select value={deckChoice} onValueChange={setDeckChoice}>
							<SelectTrigger className="w-full">
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="none">
									{m["vocabulary.addWord.noDeck"]()}
								</SelectItem>
								{(decks.data?.items ?? []).map((deck) => (
									<SelectItem key={deck.id} value={deck.id}>
										{deck.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</Field>
				)}
				<Field data-invalid={Boolean(error)}>
					<FieldLabel htmlFor="word-definition">
						{m["vocabulary.addWord.definitionLabel"]()}
					</FieldLabel>
					<Textarea
						id="word-definition"
						rows={2}
						value={form.definition}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								definition: event.target.value,
							}))
						}
					/>
				</Field>
				<Field>
					<FieldLabel htmlFor="word-example">
						{m["vocabulary.addWord.exampleLabel"]()}
					</FieldLabel>
					<Textarea
						id="word-example"
						rows={2}
						value={form.example}
						onChange={(event) =>
							setForm((current) => ({
								...current,
								example: event.target.value,
							}))
						}
					/>
				</Field>
				{form.details ? (
					<Collapsible>
						<CollapsibleTrigger asChild>
							<Button
								type="button"
								variant="ghost"
								size="sm"
								className="w-fit px-1 text-muted-foreground"
							>
								<ChevronDown data-icon="inline-end" />
								{m["vocabulary.addWord.detailsTitle"]()}
							</Button>
						</CollapsibleTrigger>
						<CollapsibleContent className="flex flex-col gap-2 rounded-xl border border-border bg-muted/40 p-3 text-xs">
							{form.details.ipa ? (
								<p className="font-mono">/{form.details.ipa}/</p>
							) : null}
							{form.details.syllables ? (
								<p>
									<span className="font-bold">
										{m["vocabulary.detail.syllables"]()}
									</span>{" "}
									{form.details.syllables}
								</p>
							) : null}
							{form.details.stressTip ? (
								<p>
									<span className="font-bold">
										{m["vocabulary.detail.stressTip"]()}
									</span>{" "}
									{form.details.stressTip}
								</p>
							) : null}
							{form.details.commonSlip ? (
								<p>
									<span className="font-bold">
										{m["vocabulary.detail.commonSlip"]()}
									</span>
									{form.details.commonSlip}
								</p>
							) : null}
							{form.details.longDefinition ? (
								<p>{form.details.longDefinition}</p>
							) : null}
							{form.details.etymology ? (
								<p>
									<span className="font-bold">
										{m["vocabulary.detail.origin"]()}
									</span>{" "}
									{form.details.etymology}
								</p>
							) : null}
							{form.details.contexts.length > 0 ? (
								<p>
									<span className="font-bold">
										{m["vocabulary.detail.contextsTitle"]()}
									</span>{" "}
									{form.details.contexts.map((c) => `“${c.quote}”`).join(" ")}
								</p>
							) : null}
							{form.details.collocations.length > 0 ? (
								<p>
									<span className="font-bold">
										{m["vocabulary.detail.collocationsTitle"]()}
									</span>{" "}
									{form.details.collocations.map((c) => c.phrase).join(" · ")}
								</p>
							) : null}
							{form.details.wordForms.length > 0 ? (
								<p>
									<span className="font-bold">
										{m["vocabulary.detail.morphologyTitle"]()}
									</span>{" "}
									{form.details.wordForms
										.map((f) => `${f.formType}: ${f.forms.join(", ")}`)
										.join(" · ")}
								</p>
							) : null}
						</CollapsibleContent>
					</Collapsible>
				) : null}
				{deckFull ? (
					<p className="text-xs font-bold text-muted-foreground">
						{m["vocabulary.deck.limitReached"]({ max: MaxDeckWords })}
					</p>
				) : null}
				{error ? <FieldError>{error}</FieldError> : null}
			</FieldGroup>
			<DialogFooter>
				<Button type="button" variant="ghost" onClick={onClose}>
					{m["common.actions.cancel"]()}
				</Button>
				<SubmitButton
					type="button"
					onClick={submit}
					pending={addWord.isPending}
					disabled={saveBlocked}
				>
					{m["vocabulary.addWord.save"]()}
				</SubmitButton>
			</DialogFooter>
		</>
	);
}
