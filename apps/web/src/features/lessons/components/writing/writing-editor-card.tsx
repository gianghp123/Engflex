import { cn } from "cn";
import { Copy, Send, Trash2 } from "lucide-react";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { Textarea } from "#/components/ui/textarea";
import { m } from "#/paraglide/messages";

export function WritingEditorCard({
	value,
	onChange,
	maxWords,
	onClear,
	onDelete,
	onSubmit,
	onRevise,
	submitted = false,
	pending = false,
	deletePending = false,
	canDelete = false,
}: {
	value: string;
	onChange: (next: string) => void;
	maxWords: number;
	onClear: () => void;
	onDelete?: () => void;
	onSubmit: () => void;
	onRevise: () => void;
	submitted?: boolean;
	pending?: boolean;
	deletePending?: boolean;
	canDelete?: boolean;
}) {
	const words = value.trim() ? value.trim().split(/\s+/).length : 0;
	const overLimit = words > maxWords;

	return (
		<div className="surface-card flex h-full flex-col gap-3 p-5">
			<div className="flex items-center justify-between gap-2">
				<h2 className="text-lg font-bold text-foreground">
					{m["lessons.writing.responseTitle"]()}
				</h2>
				<Button
					type="button"
					variant="ghost"
					size="sm"
					className="btn btn-outline px-2"
					onClick={() => {
						void navigator.clipboard?.writeText(value).catch(() => {});
					}}
				>
					<Copy data-icon="inline-start" />
					{m["common.actions.copy"]()}
				</Button>
			</div>
			<Textarea
				value={value}
				onChange={(event) => onChange(event.target.value)}
				rows={7}
				placeholder={m["lessons.writing.placeholder"]()}
				className="field min-h-40 text-[15px]"
				disabled={pending}
			/>
			<div className="flex flex-wrap items-center gap-2">
				<span
					className={cn(
						"chip px-2 py-0.5 text-[11px]",
						overLimit && "border-ai-coral text-ai-coral",
					)}
				>
					{m["lessons.writing.wordsCount"]({ count: words, max: maxWords })}
				</span>
			</div>
			<div className="mt-auto flex items-center justify-between gap-2 pt-1">
				<div className="flex flex-wrap items-center gap-2">
					<Button
						type="button"
						variant="ghost"
						size="sm"
						className="btn btn-outline"
						onClick={onClear}
						disabled={pending}
					>
						{m["common.actions.clear"]()}
					</Button>
					{canDelete && onDelete ? (
						<SubmitButton
							type="button"
							variant="ghost"
							size="sm"
							className="btn btn-outline text-ai-coral"
							pending={deletePending}
							disabled={pending}
							onClick={onDelete}
						>
							<Trash2 data-icon="inline-start" />
							{m["lessons.writing.deleteText"]()}
						</SubmitButton>
					) : null}
				</div>
				{submitted && !pending ? (
					<Button
						type="button"
						size="sm"
						className="btn btn-primary"
						onClick={onRevise}
					>
						<Send data-icon="inline-start" />
						{m["lessons.writing.revise"]()}
					</Button>
				) : (
					<SubmitButton
						type="button"
						size="sm"
						className="btn btn-primary"
						onClick={onSubmit}
						disabled={words === 0}
						pending={pending}
					>
						<Send data-icon="inline-start" />
						{m["lessons.writing.submit"]()}
					</SubmitButton>
				)}
			</div>
		</div>
	);
}
