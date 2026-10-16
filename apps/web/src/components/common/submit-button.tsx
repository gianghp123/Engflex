import type { VariantProps } from "class-variance-authority";
import { cn } from "cn";
import type * as React from "react";
import { Spinner } from "#/components/common/spinner";
import { Button, type buttonVariants } from "#/components/ui/button";

/** Icon-only sizes have no label, so the spinner takes the icon's place.
 *  Every other size uses the centered overlay (see `SubmitButton`). */
const ICON_SIZE = /^icon(-|$)/;

/**
 * `Button` with a pending state, for anything driven by a mutation.
 *
 * Takes the same props as `Button` (it reuses `buttonVariants`' type, so
 * every variant, size, `asChild` and native button prop stays valid) plus:
 *
 * - `pending` — native `disabled` + `aria-busy`, and a centered spinner
 *   overlay. The label stays in flow at `opacity-0` (width frozen, and the
 *   accessible name survives — `visibility:hidden` would drop it), so the
 *   button never changes size mid-flight. `disabled:opacity-50` keeps the
 *   disabled look. Icon-only sizes swap icon for spinner instead (same box,
 *   no shift possible).
 * - `loadingLabel` — optional visible label while pending. Supplying it also
 *   hands the whole pending presentation to this component: the spinner
 *   replaces the entire children (icon included) so a leading icon and the
 *   spinner never appear side by side. Omit it for the centered overlay
 *   (label held invisible, spinner on top), which needs no message key.
 *
 * Pass `pending` from the mutation (`mutation.isPending`) rather than local
 * state: a local flag would clear the spinner before the request resolves.
 *
 * With `asChild` no spinner is injected — `Slot` requires a single child, so
 * adding one would break every `asChild` usage. `aria-busy`/`data-pending`
 * still land on the rendered child for the consumer to style.
 */
function SubmitButton({
	pending = false,
	loadingLabel,
	size,
	className,
	disabled,
	children,
	...props
}: React.ComponentProps<"button"> &
	VariantProps<typeof buttonVariants> & {
		asChild?: boolean;
		pending?: boolean;
		loadingLabel?: string;
	}) {
	// Centered-spinner overlay, not a label swap: the invisible label holds
	// the intrinsic width, and the svg stays a direct child so the
	// `has-[>svg]` padding rules keep matching (losing either would shift
	// the width). The wrapper replays the button's own gap per size so the
	// held width is byte-identical to the idle one.
	// `disabled` is destructured out (not left in `...props`) so a caller
	// guard like `disabled={words === 0}` can never override the pending
	// state back to enabled — the `{...props}` spread below would win.
	const iconSize = ICON_SIZE.test(size ?? "");
	const overlay = pending && !props.asChild && !loadingLabel && !iconSize;
	return (
		<Button
			size={size}
			className={overlay ? cn("relative", className) : className}
			disabled={pending || disabled}
			aria-busy={pending || undefined}
			data-pending={pending || undefined}
			{...props}
		>
			{pending && !props.asChild ? (
				loadingLabel ? (
					<>
						<Spinner />
						{loadingLabel}
					</>
				) : iconSize ? (
					<Spinner />
				) : (
					<>
						<span
							className={cn(
								"flex items-center opacity-0",
								size === "xs" ? "gap-1" : size === "sm" ? "gap-1.5" : "gap-2",
							)}
						>
							{children}
						</span>
						<Spinner className="absolute inset-0 m-auto" />
					</>
				)
			) : (
				children
			)}
		</Button>
	);
}

export { SubmitButton };
