import { Logo } from "#/components/common/logo";
import { m } from "#/paraglide/messages";

export function FlowHeader() {
	return (
		<div className="flex items-center gap-3 p-4">
			<Logo variant="mark" />
			<div className="flex flex-col leading-tight">
				<span className="text-sm font-bold tracking-tight">
					{m["common.brand.appName"]()}
				</span>
				<span className="text-[11px] text-muted-foreground">
					{m["onboarding.tagline"]()}
				</span>
			</div>
		</div>
	);
}
