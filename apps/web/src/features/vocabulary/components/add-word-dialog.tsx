import type { CEFR } from "@engflex/contracts";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
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
import { useCreateCustomWord } from "#/features/vocabulary/queries";
import { m } from "#/paraglide/messages";

const CEFR_OPTIONS: Array<{ value: CEFR; label: () => string }> = [
	{ value: "B1", label: () => m["vocabulary.addWord.cefrOptions.b1"]() },
	{ value: "B2", label: () => m["vocabulary.addWord.cefrOptions.b2"]() },
	{ value: "C1", label: () => m["vocabulary.addWord.cefrOptions.c1"]() },
];

const EMPTY_FORM = {
	term: "",
	cefr: "B2" as CEFR,
	definition: "",
};

export function AddWordDialog() {
	const [open, setOpen] = useState(false);
	const [form, setForm] = useState(EMPTY_FORM);
	const [error, setError] = useState<string | null>(null);
	const createCustomWord = useCreateCustomWord();

	function submit() {
		if (!form.term.trim() || !form.definition.trim()) {
			setError(m["vocabulary.addWord.required"]());
			return;
		}
		createCustomWord.mutate(
			{
				term: form.term.trim(),
				cefr: form.cefr,
				definition: form.definition.trim(),
			},
			{
				onSuccess: () => {
					toast.success(m["common.toast.addedToVocabulary"]());
					setOpen(false);
					setForm(EMPTY_FORM);
					setError(null);
				},
				onError: () => toast.error(m["common.toast.error"]()),
			},
		);
	}

	return (
		<>
			<Button type="button" onClick={() => setOpen(true)}>
				<Plus data-icon="inline-start" />
				{m["vocabulary.addWord.button"]()}
			</Button>
			<Dialog open={open} onOpenChange={setOpen}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{m["vocabulary.addWord.title"]()}</DialogTitle>
						<DialogDescription>
							{m["vocabulary.addWord.description"]()}
						</DialogDescription>
					</DialogHeader>
					<FieldGroup>
						<Field data-invalid={Boolean(error)}>
							<FieldLabel htmlFor="word-term">
								{m["vocabulary.addWord.termLabel"]()}
							</FieldLabel>
							<Input
								id="word-term"
								value={form.term}
								onChange={(event) =>
									setForm((current) => ({
										...current,
										term: event.target.value,
									}))
								}
							/>
						</Field>
						<Field>
							<FieldLabel>{m["vocabulary.addWord.cefrLabel"]()}</FieldLabel>
							<Select
								value={form.cefr}
								onValueChange={(value) =>
									setForm((current) => ({ ...current, cefr: value as CEFR }))
								}
							>
								<SelectTrigger className="w-full">
									<SelectValue />
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
						{error ? <FieldError>{error}</FieldError> : null}
					</FieldGroup>
					<DialogFooter>
						<Button
							type="button"
							variant="ghost"
							onClick={() => setOpen(false)}
						>
							{m["common.actions.cancel"]()}
						</Button>
						<Button
							type="button"
							onClick={submit}
							disabled={createCustomWord.isPending}
						>
							{m["vocabulary.addWord.save"]()}
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	);
}
