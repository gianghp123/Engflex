import type { ExamReviewContent } from "@engflex/contracts";
import { useId, useState } from "react";
import { Button } from "#/components/ui/button";
import { Field, FieldError, FieldLabel } from "#/components/ui/field";
import { Input } from "#/components/ui/input";
import { Textarea } from "#/components/ui/textarea";
import { ConfirmDeleteDialog } from "#/features/admin/components/confirm-delete-dialog";
import { ResourceFormDialog } from "#/features/admin/components/resource-form-dialog";
import { m } from "#/paraglide/messages";
import { type ExamDraft, validExamDraft } from "../store";

type Part = ExamReviewContent["parts"][number];
type Item = Part["items"][number];
type Question = Item["questions"][number];
type Block = Item["media"][number];

export function AuthorField({
	label,
	value,
	onChange,
	number = false,
	multiline = false,
	readOnly = false,
}: {
	label: string;
	value: string | number | null;
	onChange: (value: string) => void;
	number?: boolean;
	multiline?: boolean;
	readOnly?: boolean;
}) {
	const id = useId();
	return (
		<Field>
			<FieldLabel htmlFor={id}>{label}</FieldLabel>
			{multiline ? (
				<Textarea
					readOnly={readOnly}
					id={id}
					value={value ?? ""}
					onChange={(e) => onChange(e.target.value)}
				/>
			) : (
				<Input
					readOnly={readOnly}
					id={id}
					type={number ? "number" : "text"}
					step={number ? "any" : undefined}
					value={value ?? ""}
					onChange={(e) => onChange(e.target.value)}
				/>
			)}
		</Field>
	);
}
function Remove({ onRemove }: { onRemove: () => void }) {
	const [open, setOpen] = useState(false);
	return (
		<>
			<Button type="button" variant="ghost" onClick={() => setOpen(true)}>
				{m["admin.actions.delete"]()}
			</Button>
			<ConfirmDeleteDialog
				open={open}
				onOpenChange={setOpen}
				onConfirm={() => {
					onRemove();
					setOpen(false);
				}}
			/>
		</>
	);
}
function newQuestion(number: number): Question {
	return {
		id: crypto.randomUUID(),
		questionNumber: number,
		stem: "",
		options: ["A", "B", "C", "D"].map((key) => ({ key, text: "" })),
		correctKey: "A",
		explanation: null,
		vietnameseText: null,
		vocabulary: [],
		targetIds: [],
	};
}

export function ExamEditor({
	initial,
	onSave,
	onClose,
}: {
	initial: ExamDraft;
	onSave: (value: ExamDraft) => void;
	onClose: () => void;
}) {
	const [draft, setDraft] = useState(() => structuredClone(initial));
	const [assetKeys, setAssetKeys] = useState(() =>
		initial.content.assets.map(() => crypto.randomUUID()),
	);
	const [invalid, setInvalid] = useState(false);
	const content = draft.content;
	const setContent = (update: Partial<ExamReviewContent>) =>
		setDraft({ ...draft, content: { ...content, ...update } });
	return (
		<ResourceFormDialog
			open
			onOpenChange={(open) => {
				if (!open) onClose();
			}}
			title={m["admin.exams.editor.title"]()}
			onSubmit={() => {
				if (!validExamDraft(draft)) {
					setInvalid(true);
					return;
				}
				onSave(draft);
			}}
		>
			<AuthorField
				label={m["admin.exams.fields.title"]()}
				value={content.title}
				onChange={(title) => setContent({ title })}
			/>
			<AuthorField
				label={m["admin.exams.fields.slug"]()}
				value={draft.slug}
				onChange={(slug) => setDraft({ ...draft, slug })}
			/>
			<AuthorField
				label={m["admin.exams.fields.sourceKey"]()}
				value={draft.sourceKey}
				onChange={(sourceKey) => setDraft({ ...draft, sourceKey })}
			/>
			<AuthorField
				label={m["admin.exams.fields.collection"]()}
				value={content.collectionId}
				onChange={(collectionId) => setContent({ collectionId })}
			/>
			<AuthorField
				label={m["admin.exams.fields.testNumber"]()}
				value={content.testNumber}
				number
				onChange={(value) => setContent({ testNumber: Number(value) })}
			/>
			<AuthorField
				label={m["admin.exams.fields.audioAsset"]()}
				value={content.audioAssetId}
				onChange={(audioAssetId) => setContent({ audioAssetId })}
			/>
			<h3 className="font-semibold">{m["admin.exams.editor.assets"]()}</h3>
			{content.assets.map((asset, index) => (
				<div key={assetKeys[index]} className="rounded-lg border p-3">
					<AuthorField
						label={m["admin.exams.fields.assetId"]()}
						value={asset.sourceId}
						onChange={(sourceId) =>
							setContent({
								assets: content.assets.map((value, i) =>
									i === index ? { ...value, sourceId } : value,
								),
							})
						}
					/>
					<AuthorField
						label={m["admin.exams.fields.hash"]()}
						value={asset.sha256}
						onChange={(sha256) =>
							setContent({
								assets: content.assets.map((value, i) =>
									i === index ? { ...value, sha256 } : value,
								),
							})
						}
					/>
					<Remove
						onRemove={() => {
							setAssetKeys(assetKeys.filter((_, i) => i !== index));
							setContent({
								assets: content.assets.filter((_, i) => i !== index),
							});
						}}
					/>
				</div>
			))}
			<Button
				type="button"
				variant="outline"
				onClick={() => {
					setAssetKeys([...assetKeys, crypto.randomUUID()]);
					setContent({
						assets: [...content.assets, { sourceId: "", sha256: "" }],
					});
				}}
			>
				{m["admin.exams.editor.addAsset"]()}
			</Button>
			<h3 className="font-semibold">{m["admin.exams.editor.parts"]()}</h3>
			{content.parts.map((part, index) => (
				<details key={part.partNumber} className="rounded-lg border p-3" open>
					<summary>
						{m["admin.exams.editor.part"]({ number: part.partNumber })}
					</summary>
					<AuthorField
						label={m["admin.exams.fields.partNumber"]()}
						readOnly
						value={part.partNumber}
						number
						onChange={(value) => {
							const partNumber = Number(value);
							setContent({
								parts: content.parts.map((p, i) =>
									i === index
										? {
												...p,
												partNumber,
												section: partNumber <= 4 ? "listening" : "reading",
											}
										: p,
								),
							});
						}}
					/>
					<PartEditor
						part={part}
						onChange={(value) =>
							setContent({
								parts: content.parts.map((p, i) => (i === index ? value : p)),
							})
						}
					/>
					<Remove
						onRemove={() =>
							setContent({ parts: content.parts.filter((_, i) => i !== index) })
						}
					/>
				</details>
			))}
			<Button
				type="button"
				variant="outline"
				disabled={content.parts.length >= 7}
				onClick={() => {
					const partNumber = [1, 2, 3, 4, 5, 6, 7].find(
						(number) =>
							!content.parts.some((part) => part.partNumber === number),
					);
					if (!partNumber) return;
					setContent({
						parts: [
							...content.parts,
							{
								partNumber,
								section: partNumber <= 4 ? "listening" : "reading",
								items: [],
							},
						].sort((a, b) => a.partNumber - b.partNumber),
					});
				}}
			>
				{m["admin.exams.editor.addPart"]()}
			</Button>
			{invalid && <FieldError>{m["admin.exams.editor.invalid"]()}</FieldError>}
		</ResourceFormDialog>
	);
}

function PartEditor({
	part,
	onChange,
}: {
	part: Part;
	onChange: (value: Part) => void;
}) {
	return (
		<div className="flex flex-col gap-3">
			{part.items.map((item, index) => (
				<details key={item.id} className="rounded border p-3">
					<summary>
						{m["admin.exams.editor.item"]({ number: index + 1 })}
					</summary>
					<ItemEditor
						item={item}
						onChange={(value) =>
							onChange({
								...part,
								items: part.items.map((old, i) => (i === index ? value : old)),
							})
						}
					/>
					<Remove
						onRemove={() =>
							onChange({
								...part,
								items: part.items.filter((_, i) => i !== index),
							})
						}
					/>
				</details>
			))}
			<Button
				type="button"
				variant="outline"
				onClick={() =>
					onChange({
						...part,
						items: [
							...part.items,
							{
								id: crypto.randomUUID(),
								type: [1, 2, 5].includes(part.partNumber) ? "single" : "group",
								questions: [],
								passages: [],
								transcript: [],
								media: [],
								itemAudio: null,
								stimulusAudio: null,
							},
						],
					})
				}
			>
				{m["admin.exams.editor.addItem"]()}
			</Button>
		</div>
	);
}
function ItemEditor({
	item,
	onChange,
}: {
	item: Item;
	onChange: (value: Item) => void;
}) {
	return (
		<div className="flex flex-col gap-3">
			<AuthorField
				label={m["admin.exams.fields.id"]()}
				readOnly
				value={item.id}
				onChange={(id) => onChange({ ...item, id })}
			/>
			<label>
				{m["admin.exams.fields.itemType"]()}
				<select
					className="field"
					value={item.type}
					onChange={(e) => onChange({ ...item, type: e.target.value })}
				>
					<option value="single">single</option>
					<option value="group">group</option>
				</select>
			</label>
			{(["itemAudio", "stimulusAudio"] as const).map((key) => (
				<div key={key} className="rounded border p-2">
					<h4>
						{key === "itemAudio"
							? m["admin.exams.fields.itemAudio"]()
							: m["admin.exams.fields.stimulusAudio"]()}
					</h4>
					{item[key] ? (
						<>
							<AuthorField
								label={m["admin.exams.fields.assetId"]()}
								value={item[key].assetId}
								onChange={(assetId) =>
									onChange({ ...item, [key]: { ...item[key], assetId } })
								}
							/>
							<AuthorField
								label={m["admin.exams.fields.start"]()}
								number
								value={item[key].startTime}
								onChange={(v) =>
									onChange({
										...item,
										[key]: { ...item[key], startTime: Number(v) },
									})
								}
							/>
							<AuthorField
								label={m["admin.exams.fields.end"]()}
								number
								value={item[key].endTime}
								onChange={(v) =>
									onChange({
										...item,
										[key]: { ...item[key], endTime: Number(v) },
									})
								}
							/>
							<Remove onRemove={() => onChange({ ...item, [key]: null })} />
						</>
					) : (
						<Button
							type="button"
							variant="outline"
							onClick={() =>
								onChange({
									...item,
									[key]: { assetId: "", startTime: 0, endTime: 1 },
								})
							}
						>
							{m["admin.actions.create"]()}
						</Button>
					)}
				</div>
			))}
			{item.questions.map((question, index) => (
				<details key={question.id} className="rounded border p-2">
					<summary>
						{m["admin.exams.editor.question"]({
							number: question.questionNumber,
						})}
					</summary>
					<QuestionEditor
						question={question}
						onChange={(value) =>
							onChange({
								...item,
								questions: item.questions.map((q, i) =>
									i === index ? value : q,
								),
							})
						}
					/>
					<Remove
						onRemove={() =>
							onChange({
								...item,
								questions: item.questions.filter((_, i) => i !== index),
							})
						}
					/>
				</details>
			))}
			<Button
				type="button"
				variant="outline"
				onClick={() =>
					onChange({
						...item,
						questions: [
							...item.questions,
							newQuestion(
								Math.max(0, ...item.questions.map((q) => q.questionNumber)) + 1,
							),
						],
					})
				}
			>
				{m["admin.exams.editor.addQuestion"]()}
			</Button>
			<h4>{m["admin.exams.editor.passages"]()}</h4>
			{item.passages.map((passage, index) => (
				<div key={passage.id} className="rounded border p-2">
					<AuthorField
						label={m["admin.exams.fields.id"]()}
						readOnly
						value={passage.id}
						onChange={(id) =>
							onChange({
								...item,
								passages: item.passages.map((p, i) =>
									i === index ? { ...p, id } : p,
								),
							})
						}
					/>
					<AuthorField
						label={m["admin.exams.fields.format"]()}
						value={passage.format}
						onChange={(format) =>
							onChange({
								...item,
								passages: item.passages.map((p, i) =>
									i === index ? { ...p, format } : p,
								),
							})
						}
					/>
					<BlocksEditor
						blocks={passage.blocks}
						onChange={(blocks) =>
							onChange({
								...item,
								passages: item.passages.map((p, i) =>
									i === index ? { ...p, blocks } : p,
								),
							})
						}
					/>
					<Remove
						onRemove={() =>
							onChange({
								...item,
								passages: item.passages.filter((_, i) => i !== index),
							})
						}
					/>
				</div>
			))}
			<Button
				type="button"
				variant="outline"
				onClick={() =>
					onChange({
						...item,
						passages: [
							...item.passages,
							{ id: crypto.randomUUID(), format: null, blocks: [] },
						],
					})
				}
			>
				{m["admin.exams.editor.addPassage"]()}
			</Button>
			<h4>{m["admin.exams.editor.media"]()}</h4>
			<BlocksEditor
				blocks={item.media}
				onChange={(media) => onChange({ ...item, media })}
			/>
			<h4>{m["admin.exams.editor.transcript"]()}</h4>
			{item.transcript.map((line, index) => (
				<div key={line.id} className="rounded border p-2">
					<AuthorField
						label={m["admin.exams.fields.role"]()}
						value={line.role}
						onChange={(role) =>
							onChange({
								...item,
								transcript: item.transcript.map((t, i) =>
									i === index ? { ...t, role } : t,
								),
							})
						}
					/>
					<AuthorField
						label={m["admin.exams.fields.speaker"]()}
						value={line.speaker}
						onChange={(speaker) =>
							onChange({
								...item,
								transcript: item.transcript.map((t, i) =>
									i === index ? { ...t, speaker } : t,
								),
							})
						}
					/>
					<AuthorField
						label={m["admin.exams.fields.text"]()}
						multiline
						value={line.text}
						onChange={(text) =>
							onChange({
								...item,
								transcript: item.transcript.map((t, i) =>
									i === index ? { ...t, text } : t,
								),
							})
						}
					/>
					<AuthorField
						label={m["admin.exams.fields.translation"]()}
						multiline
						value={line.vietnameseText}
						onChange={(vietnameseText) =>
							onChange({
								...item,
								transcript: item.transcript.map((t, i) =>
									i === index ? { ...t, vietnameseText } : t,
								),
							})
						}
					/>
					<Remove
						onRemove={() =>
							onChange({
								...item,
								transcript: item.transcript.filter((_, i) => i !== index),
							})
						}
					/>
				</div>
			))}
			<Button
				type="button"
				variant="outline"
				onClick={() =>
					onChange({
						...item,
						transcript: [
							...item.transcript,
							{
								id: crypto.randomUUID(),
								role: "speaker",
								speaker: null,
								optionKey: null,
								text: "",
								vietnameseText: null,
							},
						],
					})
				}
			>
				{m["admin.exams.editor.addTranscript"]()}
			</Button>
		</div>
	);
}
function QuestionEditor({
	question: q,
	onChange,
}: {
	question: Question;
	onChange: (value: Question) => void;
}) {
	return (
		<div className="flex flex-col gap-3">
			<AuthorField
				label={m["admin.exams.fields.id"]()}
				readOnly
				value={q.id}
				onChange={(id) => onChange({ ...q, id })}
			/>
			<AuthorField
				label={m["admin.exams.fields.questionNumber"]()}
				number
				value={q.questionNumber}
				onChange={(v) => onChange({ ...q, questionNumber: Number(v) })}
			/>
			<AuthorField
				label={m["admin.exams.fields.stem"]()}
				multiline
				value={q.stem}
				onChange={(stem) => onChange({ ...q, stem })}
			/>
			{q.options.map((option, index) => (
				<div key={option.key} className="rounded border p-2">
					<AuthorField
						label={m["admin.exams.fields.optionKey"]()}
						readOnly
						value={option.key}
						onChange={(key) =>
							onChange({
								...q,
								options: q.options.map((o, i) =>
									i === index ? { ...o, key } : o,
								),
							})
						}
					/>
					<AuthorField
						label={m["admin.exams.fields.text"]()}
						value={option.text}
						onChange={(text) =>
							onChange({
								...q,
								options: q.options.map((o, i) =>
									i === index ? { ...o, text } : o,
								),
							})
						}
					/>
					<Remove
						onRemove={() =>
							onChange({
								...q,
								options: q.options.filter((_, i) => i !== index),
							})
						}
					/>
				</div>
			))}
			<Button
				type="button"
				variant="outline"
				disabled={q.options.length >= 4}
				onClick={() => {
					const key = ["A", "B", "C", "D"].find(
						(value) => !q.options.some((option) => option.key === value),
					);
					if (key)
						onChange({
							...q,
							options: [...q.options, { key, text: "" }].sort((a, b) =>
								a.key.localeCompare(b.key),
							),
						});
				}}
			>
				{m["admin.exams.editor.addOption"]()}
			</Button>
			<AuthorField
				label={m["admin.exams.fields.correctKey"]()}
				value={q.correctKey}
				onChange={(correctKey) => onChange({ ...q, correctKey })}
			/>
			<AuthorField
				label={m["admin.exams.fields.explanation"]()}
				multiline
				value={q.explanation}
				onChange={(explanation) => onChange({ ...q, explanation })}
			/>
			<AuthorField
				label={m["admin.exams.fields.translation"]()}
				multiline
				value={q.vietnameseText}
				onChange={(vietnameseText) => onChange({ ...q, vietnameseText })}
			/>
			<AuthorField
				label={m["admin.exams.fields.targetIds"]()}
				value={q.targetIds.join(", ")}
				onChange={(value) =>
					onChange({
						...q,
						targetIds: value
							.split(",")
							.map((id) => id.trim())
							.filter(Boolean),
					})
				}
			/>
		</div>
	);
}
function BlocksEditor({
	blocks,
	onChange,
}: {
	blocks: Block[];
	onChange: (value: Block[]) => void;
}) {
	return (
		<div className="flex flex-col gap-2">
			{blocks.map((block, index) => {
				const set = (update: Partial<Block>) =>
					onChange(
						blocks.map((b, i) => (i === index ? { ...b, ...update } : b)),
					);
				return (
					<div key={block.id} className="rounded border p-2">
						<AuthorField
							label={m["admin.exams.fields.id"]()}
							readOnly
							value={block.id}
							onChange={(id) => set({ id })}
						/>
						<label>
							{m["admin.exams.fields.blockType"]()}
							<select
								className="field"
								value={block.type}
								onChange={(e) => set({ type: e.target.value })}
							>
								{[
									"paragraph",
									"heading",
									"emailHeader",
									"chatMessage",
									"table",
									"image",
								].map((type) => (
									<option key={type}>{type}</option>
								))}
							</select>
						</label>
						<AuthorField
							label={m["admin.exams.fields.text"]()}
							multiline
							value={block.text}
							onChange={(text) => set({ text })}
						/>
						<AuthorField
							label={m["admin.exams.fields.translation"]()}
							multiline
							value={block.vietnameseText}
							onChange={(vietnameseText) => set({ vietnameseText })}
						/>
						<AuthorField
							label={m["admin.exams.fields.assetId"]()}
							value={block.assetId}
							onChange={(assetId) => set({ assetId: assetId || null })}
						/>
						{block.type === "table" && (
							<AuthorField
								label={m["admin.exams.fields.rows"]()}
								multiline
								value={
									block.rows?.map((row) => row.join("\t")).join("\n") ?? ""
								}
								onChange={(text) =>
									set({ rows: text.split("\n").map((row) => row.split("\t")) })
								}
							/>
						)}
						{block.inline.map((node, n) => (
							<div key={node.id} className="rounded border p-2">
								<AuthorField
									label={m["admin.exams.fields.id"]()}
									readOnly
									value={node.id}
									onChange={(id) =>
										set({
											inline: block.inline.map((old, i) =>
												i === n ? { ...old, id } : old,
											),
										})
									}
								/>
								<label>
									{m["admin.exams.fields.inlineType"]()}
									<select
										className="field"
										value={node.type}
										onChange={(e) =>
											set({
												inline: block.inline.map((old, i) =>
													i === n ? { ...old, type: e.target.value } : old,
												),
											})
										}
									>
										{["text", "markedText", "blank", "insertionPoint"].map(
											(type) => (
												<option key={type}>{type}</option>
											),
										)}
									</select>
								</label>
								<AuthorField
									label={m["admin.exams.fields.text"]()}
									value={node.text}
									onChange={(text) =>
										set({
											inline: block.inline.map((old, i) =>
												i === n ? { ...old, text: text || null } : old,
											),
										})
									}
								/>
								<AuthorField
									label={m["admin.exams.fields.questionId"]()}
									value={node.questionId}
									onChange={(questionId) =>
										set({
											inline: block.inline.map((old, i) =>
												i === n
													? { ...old, questionId: questionId || null }
													: old,
											),
										})
									}
								/>
								<Remove
									onRemove={() =>
										set({ inline: block.inline.filter((_, i) => i !== n) })
									}
								/>
							</div>
						))}
						<Button
							type="button"
							variant="outline"
							onClick={() =>
								set({
									inline: [
										...block.inline,
										{
											id: crypto.randomUUID(),
											type: "text",
											text: "",
											questionId: null,
										},
									],
								})
							}
						>
							{m["admin.exams.editor.addInline"]()}
						</Button>
						<Remove
							onRemove={() => onChange(blocks.filter((_, i) => i !== index))}
						/>
					</div>
				);
			})}
			<Button
				type="button"
				variant="outline"
				onClick={() =>
					onChange([
						...blocks,
						{
							id: crypto.randomUUID(),
							type: "paragraph",
							text: "",
							vietnameseText: null,
							rows: null,
							inline: [],
							assetId: null,
						},
					])
				}
			>
				{m["admin.exams.editor.addBlock"]()}
			</Button>
		</div>
	);
}
