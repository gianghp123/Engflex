import type { CEFR, VocabularyPartOfSpeech } from "@engflex/contracts";
import {
	CEFR_VALUES,
	VOCABULARY_PART_OF_SPEECH_VALUES,
} from "@engflex/contracts";
import { useState } from "react";
import { toast } from "sonner";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { DialogDescription, DialogFooter } from "#/components/ui/dialog";
import {
	Field,
	FieldError,
	FieldGroup,
	FieldLabel,
} from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "#/components/ui/select";
import { Textarea } from "#/components/ui/textarea";
import { POS_LABELS } from "#/features/vocabulary/components/pos-label";
import { useAddWord, useMyVocabulary } from "#/features/vocabulary/queries";
import { termKey } from "#/features/vocabulary/term-key";
import { enumOptions } from "#/lib/enum-options";
import { m } from "#/paraglide/messages";

const CEFR_OPTIONS = enumOptions(CEFR_VALUES);
const POS_OPTIONS = enumOptions(VOCABULARY_PART_OF_SPEECH_VALUES, POS_LABELS);

const EMPTY_FORM: {
	term: string;
	cefr: CEFR | "";
	partOfSpeech: VocabularyPartOfSpeech;
	definition: string;
	example: string;
} = {
	term: "",
	cefr: "",
	partOfSpeech: "noun",
	definition: "",
	example: "",
};

/**
 * The "add a word" mode: a custom word, saved to the reader's own list and
 * enrolled into `deckId` when the flow was opened from a deck.
 *
 * A typed term that already exists in the library is called out, because saving
 * it reuses that entry instead of creating a second one. The check is local and
 * covers the library page the app already holds; the system/global-dictionary
 * half is deferred (it needs `GET /vocabulary-items?search=`, debounced) and is
 * reported as no match rather than faked.
 */
export function AddWordForm({
	deckId,
	onClose,
}: {
	deckId?: string;
	onClose: () => void;
}) {
	const addWord = useAddWord();
	const library = useMyVocabulary();
	const [form, setForm] = useState(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);

	const typedKey = termKey(form.term);
	const existingInLibrary =
		typedKey.length > 0
			? (library.data?.items ?? []).some(
					(item) => termKey(item.term) === typedKey,
				)
			: false;

	function submit() {
		const term = form.term.trim();
		const definition = form.definition.trim();
		const cefr = form.cefr;
		if (!term || !definition || !cefr) {
			setError(m["vocabulary.addWord.required"]());
			return;
		}
		setError(null);
		const input = {
			term,
			cefr,
			partOfSpeech: form.partOfSpeech,
			definition,
			example: form.example.trim(),
		};
		addWord.mutate(deckId ? { ...input, deckId } : input, {
			onSuccess: () => {
				toast.success(m["common.toast.addedToVocabulary"]());
				setForm(EMPTY_FORM);
				setError(null);
				onClose();
			},
			onError: () => toast.error(m["vocabulary.addWord.failed"]()),
		});
	}

	return (
		<>
			<DialogDescription className="font-medium">
				{m["vocabulary.addWord.description"]()}
			</DialogDescription>
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
				>
					{m["vocabulary.addWord.save"]()}
				</SubmitButton>
			</DialogFooter>
		</>
	);
}
