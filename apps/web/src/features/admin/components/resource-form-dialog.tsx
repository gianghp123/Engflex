import type { ReactNode } from "react";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { FieldGroup } from "#/components/ui/field";
import { m } from "#/paraglide/messages";

/**
 * Create/edit shell shared by every admin resource form: dialog chrome +
 * Cancel/Save footer only. Each resource owns its own fields (`Field` /
 * `FieldLabel` / `FieldError` from `#/components/ui/field`) and submit
 * handling — this component stays field-agnostic on purpose, mirroring
 * `features/vocabulary/components/add-word-dialog.tsx`.
 */
export function ResourceFormDialog({
	open,
	onOpenChange,
	title,
	description,
	onSubmit,
	pending = false,
	children,
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title: string;
	description?: string;
	onSubmit: () => void;
	pending?: boolean;
	children: ReactNode;
}) {
	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent
				className="max-h-[85vh] overflow-y-auto"
				onOpenAutoFocus={(event) => event.preventDefault()}
			>
				<DialogHeader>
					<DialogTitle>{title}</DialogTitle>
					{description ? (
						<DialogDescription>{description}</DialogDescription>
					) : null}
				</DialogHeader>
				<form
					onSubmit={(event) => {
						event.preventDefault();
						onSubmit();
					}}
				>
					<FieldGroup>{children}</FieldGroup>
					<DialogFooter className="mt-6">
						<Button
							type="button"
							variant="ghost"
							onClick={() => onOpenChange(false)}
						>
							{m["admin.actions.cancel"]()}
						</Button>
						<SubmitButton type="submit" pending={pending}>
							{m["admin.actions.save"]()}
						</SubmitButton>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
