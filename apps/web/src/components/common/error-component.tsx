import { APP_ROUTES } from "#/app/app-route";
import { MoMascot, type MoVariant } from "#/components/common/mo-mascot";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";
import { Link } from "@tanstack/react-router";
import type { ReactNode } from "react";

/**
 * The one error surface. Every status page passes its own copy, so no
 * status-specific string is baked in here. Rendered bare — deliberately
 * outside AppShell — because /not-found and /error must also render for
 * signed-out visitors, who never get an AppShell.
 *
 * Mascots are existing Mo variants (no per-status artwork): 404 uses
 * "confused" (its "?" hook matches the lost-way story), 500 uses "cry".
 */
export function ErrorComponent({
  badge,
  mascot,
  caption,
  title,
  body,
  primaryLabel,
  extraAction,
}: {
  badge: string;
  mascot: MoVariant;
  caption: string;
  title: string;
  body: string;
  primaryLabel: string;
  /** Rendered beside the primary button (e.g. Retry on 500). */
  extraAction?: ReactNode;
}) {
  return (
    <div className="bg-dot-pattern flex min-h-svh flex-col items-center justify-center px-4 py-10">
      <div className="surface-card flex w-full max-w-xl flex-col items-center rounded-2xl p-6 text-center md:p-10">
        <p className="chip mb-6 px-3 py-1 text-xs">
          <span
            aria-hidden="true"
            className="inline-block size-2.5 rounded-full bg-secondary"
          />{" "}
          {badge}
        </p>
        <div className="relative mb-6 flex items-center justify-center">
          <span
            aria-hidden="true"
            className="absolute inset-0 rounded-full bg-secondary opacity-60"
          />
          <MoMascot variant={mascot} size={144} className="relative z-10" />
        </div>
        <p className="chip mb-6 px-3 py-1 text-xs">{caption}</p>
        <h1 className="mb-2 text-2xl font-bold tracking-tight text-foreground md:text-3xl">
          {title}
        </h1>
        <p className="mb-8 max-w-md text-[15px] font-medium text-muted-foreground">
          {body}
        </p>
        <div className="flex w-full flex-col items-center justify-center gap-3 sm:w-auto sm:flex-row">
          {extraAction}
          <Button asChild size="lg">
            <Link to={APP_ROUTES.HOME}>{primaryLabel}</Link>
          </Button>
        </div>
      </div>
      <p className="chip mt-6 px-3 py-1 text-xs text-muted-foreground">
        {m["errors.quote"]()}
      </p>
    </div>
  );
}
