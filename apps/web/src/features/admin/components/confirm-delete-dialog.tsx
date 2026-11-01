import { SubmitButton } from "#/components/common/submit-button";
import {
	AlertDialog,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "#/components/ui/alert-dialog";
import { m } from "#/paraglide/messages";

export function ConfirmDeleteDialog({
	open,
	onOpenChange,
	onConfirm,
	pending = false,
	title,
	description,
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onConfirm: () => void;
	pending?: boolean;
	title?: string;
	description?: string;
}) {
	return (
		<AlertDialog open={open} onOpenChange={onOpenChange}>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>
						{title ?? m["admin.confirm.deleteTitle"]()}
					</AlertDialogTitle>
					<AlertDialogDescription>
						{description ?? m["admin.confirm.deleteDescription"]()}
					</AlertDialogDescription>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel>{m["admin.actions.cancel"]()}</AlertDialogCancel>
					<SubmitButton
						type="button"
						variant="destructive"
						pending={pending}
						onClick={onConfirm}
					>
						{m["admin.actions.delete"]()}
					</SubmitButton>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
}
