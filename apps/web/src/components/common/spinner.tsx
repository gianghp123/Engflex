import { cn } from "cn";
import { Loader2Icon } from "lucide-react";

/**
 * The house spinner (`SubmitButton`, `CardShell`, `sonner`).
 * No `size-*` class on purpose: `buttonVariants` already sizes bare svgs per
 * size variant. An explicit `className` marks the centered-overlay use, which
 * is not an inline-start icon. Always decorative: the owning control or
 * container supplies the accessible name.
 */
export function Spinner({ className }: { className?: string }) {
	return (
		<Loader2Icon
			aria-hidden="true"
			className={cn("animate-spin", className)}
			data-icon={className ? undefined : "inline-start"}
		/>
	);
}
