import { MoMascot, type MoVariant } from "#/components/common/mo-mascot";
import { cn } from "cn";
import type { ReactNode } from "react";

/**
 * Shared Mo tip card. Fill + ink come from the PageHero bubble tokens
 * (`--tip-bg` / `--tip-ink`), framed as a chunky card. Callers pass
 * already-localized strings — no message keys live here.
 */
export function TipCard({
  title,
  body,
  variant = "nice",
  className,
}: {
  title: string;
  body?: ReactNode;
  variant?: MoVariant;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex items-center gap-3 rounded-[24px] border-2 border-border bg-tip-bg p-4 shadow-edge",
        className,
      )}
    >
      <MoMascot variant={variant} size={40} className="hidden sm:inline-flex" />
      <div className="min-w-0">
        <p className="text-[11px] font-bold tracking-wide text-tip-ink">
          {title}
        </p>
        <p className="text-[15px] font-bold text-tip-ink">{body}</p>
      </div>
    </div>
  );
}
