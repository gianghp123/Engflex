import { cn } from "cn";
import type { ReactNode } from "react";
import { Spinner } from "#/components/common/spinner";

/**
 * A card frame with a `pending` presentation, in the spirit of
 * `SubmitButton`:
 *
 * - `pending` + `skeleton` -> the shape-matched skeleton renders in the frame.
 * - `pending` alone        -> the real children are held `invisible` and a
 *   centered spinner overlays them, so the frame keeps the size the loaded
 *   content will occupy (no shift). The caller owns the children's size.
 * - otherwise              -> the children render.
 *
 * `surface` defaults to `surface-card`; a card with a different surface (the
 * active-lesson hero uses `surface-hero`) overrides it. Positional + layout
 * classes stay in `className`.
 */
export function CardShell({
	pending = false,
	skeleton,
	surface = "surface-card",
	className,
	children,
}: {
	pending?: boolean;
	skeleton?: ReactNode;
	surface?: string;
	className?: string;
	children?: ReactNode;
}) {
	if (pending && skeleton) {
		return (
			<div
				className={cn("relative", surface, className)}
				aria-busy
				data-pending
			>
				{skeleton}
			</div>
		);
	}
	if (pending) {
		return (
			<div
				className={cn("relative", surface, className)}
				aria-busy
				data-pending
			>
				<div className="invisible" aria-hidden="true">
					{children}
				</div>
				<Spinner className="absolute inset-0 m-auto" />
			</div>
		);
	}
	return <div className={cn(surface, className)}>{children}</div>;
}
