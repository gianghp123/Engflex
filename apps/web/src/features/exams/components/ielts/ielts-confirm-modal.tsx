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

export function IeltsConfirmModal({
	action,
	answeredCount,
	totalCount,
	onClose,
	onConfirm,
	pending = false,
}: {
	action: "submit" | "exit" | "reset" | null;
	answeredCount: number;
	totalCount: number;
	onClose: () => void;
	onConfirm: () => void;
	pending?: boolean;
}) {
	if (!action) return null;

	const titles = {
		submit: m["exams.ielts.take.confirmSubmitTitle"](),
		exit: m["exams.ielts.take.confirmExitTitle"](),
		reset: m["exams.ielts.take.confirmResetTitle"](),
	};

	const descriptions = {
		submit: m["exams.ielts.take.confirmSubmitDesc"]({
			answered: answeredCount,
			total: totalCount,
		}),
		exit: m["exams.ielts.take.confirmExitDesc"](),
		reset: m["exams.ielts.take.confirmResetDesc"](),
	};

	const actionButtons = {
		submit: m["exams.ielts.take.confirmSubmitAction"](),
		exit: m["exams.ielts.take.confirmExitAction"](),
		reset: m["exams.ielts.take.confirmResetAction"](),
	};

	const cancelButtons = {
		submit: m["exams.ielts.take.confirmSubmitCancel"](),
		exit: m["exams.ielts.take.confirmExitCancel"](),
		reset: m["exams.ielts.take.confirmResetCancel"](),
	};

	return (
		<AlertDialog
			open={action !== null}
			onOpenChange={(open) => !open && !pending && onClose()}
		>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>{titles[action]}</AlertDialogTitle>
					<AlertDialogDescription>
						{descriptions[action]}
					</AlertDialogDescription>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel onClick={onClose} disabled={pending}>
						{cancelButtons[action]}
					</AlertDialogCancel>
					<SubmitButton
						pending={pending}
						onClick={onConfirm}
						className={
							action === "submit"
								? "bg-emerald-600 hover:bg-emerald-700 text-white"
								: "bg-rose-600 hover:bg-rose-700 text-white"
						}
					>
						{actionButtons[action]}
					</SubmitButton>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
}
