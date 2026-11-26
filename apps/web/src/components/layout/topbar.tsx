import { useBreadcrumbItems } from "#/app/breadcrumbs";
import { PageCrumbs } from "#/components/common/page-crumbs";
import { LocaleSwitcher } from "#/components/layout/locale-switcher";
import { ThemeToggle } from "#/components/ThemeToggle";
import { Separator } from "#/components/ui/separator";
import { SidebarTrigger } from "#/components/ui/sidebar";
import { CreditsPill } from "#/features/payments/components/credits-pill";

/**
 * App header: sidebar trigger + current trail (pages publish it via
 * `useBreadcrumbs`) + locale/theme controls. The trail keeps its exact
 * position — only its source changes, from the route match chain to the
 * page-published slot.
 */
export function Topbar() {
	const items = useBreadcrumbItems();
	return (
		<header className="sticky top-0 z-10 flex h-14 shrink-0 items-center gap-2 border-b border-border bg-background px-4 text-foreground">
			<SidebarTrigger className="-ml-1" />
			<Separator
				orientation="vertical"
				className="mr-2 data-[orientation=vertical]:h-4"
			/>
			<PageCrumbs items={items} />
			<div className="ml-auto flex items-center gap-3">
				<CreditsPill />
				<LocaleSwitcher />
				<ThemeToggle />
			</div>
		</header>
	);
}
