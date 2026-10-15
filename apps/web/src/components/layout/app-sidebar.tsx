import { useUser } from "@clerk/tanstack-react-start";
import { Link, useRouterState } from "@tanstack/react-router";
import { cn } from "cn";
import {
	AudioWaveform,
	BookmarkCheck,
	BookOpen,
	ChevronRight,
	House,
	Info,
} from "lucide-react";
import { useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { Logo } from "#/components/common/logo";
import { Avatar } from "#/components/ui/avatar";
import { Badge } from "#/components/ui/badge";
import {
	Collapsible,
	CollapsibleContent,
	CollapsibleTrigger,
} from "#/components/ui/collapsible";
import {
	Sidebar,
	SidebarContent,
	SidebarFooter,
	SidebarGroup,
	SidebarGroupLabel,
	SidebarHeader,
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
	SidebarMenuSub,
	SidebarMenuSubButton,
	SidebarMenuSubItem,
	useSidebar,
} from "#/components/ui/sidebar";
import { useProfile } from "#/features/profiles/queries";
import { AccountButton } from "#/integrations/clerk/account-button";
import { m } from "#/paraglide/messages";

/**
 * Nav items render as a filled rounded tile with a knocked-out white glyph, so
 * `tile` is a solid fill (not a text colour) and every one is dark enough for a
 * white glyph to clear WCAG 1.4.11 non-text contrast (>=3:1):
 * primary 5.31:1, orange 3.56:1, blue 5.17:1, emerald 3.77:1.
 */
const NAV_GROUPS = [
	{
		label: () => m["nav.group.practice"](),
		items: [
			{
				title: () => m["nav.item.home"](),
				to: APP_ROUTES.HOME,
				icon: House,
				tile: "bg-nav-home",
			},
			{
				title: () => m["nav.item.lessons"](),
				to: APP_ROUTES.LESSONS.LIST,
				icon: BookOpen,
				tile: "bg-nav-lessons",
			},
			{
				title: () => m["nav.item.voice"](),
				to: APP_ROUTES.VOICE.LIST,
				icon: AudioWaveform,
				tile: "bg-nav-voice",
			},
		],
	},
	{
		label: () => m["nav.group.knowledge"](),
		items: [
			{
				title: () => m["nav.item.vocabulary"](),
				to: APP_ROUTES.VOCABULARY.DECKS.LIST,
				icon: BookmarkCheck,
				tile: "bg-nav-vocab",
				// The section's list-level pages, which is all that earns a nav
				// entry. Every other /vocabulary path is a drill-down — a deck or a
				// word belongs to one of these two — so it lights the parent instead.
				// Copy is the page's own title/breadcrumb, not a parallel nav string,
				// so the menu and the page cannot drift apart.
				items: [
					{
						title: () => m["vocabulary.decks.title"](),
						to: APP_ROUTES.VOCABULARY.DECKS.LIST,
					},
					{
						title: () => m["vocabulary.decks.allWords"](),
						to: APP_ROUTES.VOCABULARY.WORDS.LIST,
					},
				],
			},
		],
	},
] as const;

const LEVEL_LABELS = {
	beginner: () => m["nav.user.level.beginner"](),
	intermediate: () => m["nav.user.level.intermediate"](),
	advanced: () => m["nav.user.level.advanced"](),
} as const;

export function AppSidebar() {
	const pathname = useRouterState({
		select: (state) => state.location.pathname,
	});

	return (
		<Sidebar collapsible="icon">
			{/* A plain Link, not a SidebarMenuButton: the primitive forces every
			    collapsed menu button to `size-8!` (32px), which is narrower than
			    the 36px mascot bubble, and overriding it needs an `!important`
			    that loses the cascade to the primitive's own `size-8!`. Dropping
			    the header's horizontal padding when collapsed gives the bubble
			    the full rail width to sit centred in. */}
			<SidebarHeader className="group-data-[collapsible=icon]:px-0">
				<Link
					to={APP_ROUTES.HOME}
					className="flex h-12 items-center rounded-md px-2 hover:bg-transparent group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:p-0"
				>
					{/* One mark, not two: `hidden` on the inner wordmark span drops
					    just the wordmark, so the bubble keeps its 36px in both states.
					    Toggling the Logo root itself would collide with its own
					    `inline-flex` and let both variants render at once. */}
					<Logo
						variant="full"
						className="text-sidebar-foreground group-data-[collapsible=icon]:[&>span:last-child]:hidden"
					/>
				</Link>
			</SidebarHeader>
			<SidebarContent className="group-data-[collapsible=icon]:gap-1">
				{NAV_GROUPS.map((group) => (
					<SidebarGroup
						key={group.label()}
						className="group-data-[collapsible=icon]:p-0"
					>
						<SidebarGroupLabel>{group.label()}</SidebarGroupLabel>
						<SidebarMenu>
							{group.items.map((item) => (
								<NavMenuItem
									key={item.title()}
									item={item}
									pathname={pathname}
								/>
							))}
						</SidebarMenu>
					</SidebarGroup>
				))}
			</SidebarContent>
			<SidebarFooter>
				<SidebarMenu>
					<SidebarMenuItem>
						<SidebarMenuButton asChild tooltip={m["about.crumb"]()}>
							<Link to={APP_ROUTES.ABOUT} className="font-medium">
								<Info data-icon="inline-start" />
								<span>{m["about.crumb"]()}</span>
							</Link>
						</SidebarMenuButton>
					</SidebarMenuItem>
				</SidebarMenu>
				<SidebarUserCard />
			</SidebarFooter>
		</Sidebar>
	);
}

/** A nav entry with `items` renders as the collapsible branch; the union member
 *  is narrowed by the `in` check below, which keeps the literal `to` types that
 *  `<Link>` needs to resolve. */
type NavItem = (typeof NAV_GROUPS)[number]["items"][number];

/**
 * One nav entry. Without sub-destinations it is a link to its page; with them it
 * becomes a collapsible section, and the section's list-level pages sit inside.
 */
function NavMenuItem({ item, pathname }: { item: NavItem; pathname: string }) {
	const { state, setOpen } = useSidebar();
	const inSection = pathname.startsWith(item.to);

	// Radix is uncontrolled unless `open` is passed, and `defaultOpen` is read
	// once at mount — a client navigation does not remount the sidebar, so a
	// client-side hop into a sub-page would leave the section shut. Deriving the
	// state during render (never a syncing effect, per rules/react-effects.md):
	// with no explicit choice we follow the route, and a choice is stored beside
	// the section-ness it was made in, so leaving the section resets it.
	const [choice, setChoice] = useState<{
		inSection: boolean;
		open: boolean;
	} | null>(null);
	const open = choice?.inSection === inSection ? choice.open : inSection;

	const tile = (
		<span
			className={cn(
				"flex size-6 shrink-0 items-center justify-center rounded-[7px] text-white",
				item.tile,
			)}
		>
			<item.icon className="size-3.5" />
		</span>
	);
	const label = <span className="text-sidebar-foreground">{item.title()}</span>;

	if (!("items" in item)) {
		const isActive =
			item.to === APP_ROUTES.HOME ? pathname === APP_ROUTES.HOME : inSection;
		return (
			<SidebarMenuItem>
				<SidebarMenuButton asChild isActive={isActive} tooltip={item.title()}>
					<Link to={item.to} className="font-medium">
						{tile}
						{label}
						{isActive ? (
							<span className="ml-auto size-1.5 rounded-full bg-sidebar-primary group-data-[collapsible=icon]:hidden" />
						) : null}
					</Link>
				</SidebarMenuButton>
			</SidebarMenuItem>
		);
	}

	return (
		<Collapsible
			asChild
			open={open}
			className="group/collapsible"
			onOpenChange={(next) => {
				// Collapsed, the sub-menu is `hidden`, so the toggle would flip a
				// group nobody can see. Expanding the rail is the action the click
				// should have taken instead.
				if (state === "collapsed") {
					setOpen(true);
					return;
				}
				setChoice({ inSection, open: next });
			}}
		>
			<SidebarMenuItem>
				<CollapsibleTrigger asChild>
					{/* No `isActive` here: the parent stays a neutral header for the
					    section, and the open chevron plus the list below show where you
					    are. The mark belongs to the lit child entry. */}
					<SidebarMenuButton tooltip={item.title()}>
						{tile}
						{label}
						<ChevronRight className="ml-auto size-3.5 shrink-0 text-sidebar-foreground/60 transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90 group-data-[collapsible=icon]:hidden" />
					</SidebarMenuButton>
				</CollapsibleTrigger>
				<CollapsibleContent>
					<SidebarMenuSub>
						{item.items.map((sub) => {
							const isActive = isSubActive(sub.to, item.items, pathname);
							return (
								<SidebarMenuSubItem key={sub.title()}>
									<SidebarMenuSubButton asChild isActive={isActive}>
										<Link to={sub.to}>
											<span>{sub.title()}</span>
											{isActive ? (
												<span className="ml-auto size-1.5 rounded-full bg-sidebar-primary group-data-[collapsible=icon]:hidden" />
											) : null}
										</Link>
									</SidebarMenuSubButton>
								</SidebarMenuSubItem>
							);
						})}
					</SidebarMenuSub>
				</CollapsibleContent>
			</SidebarMenuItem>
		</Collapsible>
	);
}

/**
 * A section's landing route is also the bare prefix of its sibling's route
 * (`DECKS.LIST` versus `WORDS.LIST`), so a plain `startsWith` on both would
 * light *both* sub-items while the reader is on the words page — the decks
 * landing is not an ancestor of the words list, it just shares its prefix.
 * Longest match wins: an entry yields the highlight to a sibling whose route is
 * the deeper match for the same pathname.
 */
function isSubActive(
	to: string,
	siblings: readonly { to: string }[],
	pathname: string,
) {
	return (
		pathname.startsWith(to) &&
		!siblings.some(
			(other) =>
				other.to !== to &&
				other.to.startsWith(to) &&
				pathname.startsWith(other.to),
		)
	);
}

function SidebarUserCard() {
	const { user } = useUser();
	const level = useProfile().data?.preferences.level;

	const name = user?.fullName ?? user?.firstName ?? m["nav.user.learner"]();

	return (
		<div className="flex items-center gap-2.5 rounded-[16px] border-2 border-sidebar-border bg-card p-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:rounded-full group-data-[collapsible=icon]:border-0 group-data-[collapsible=icon]:bg-transparent group-data-[collapsible=icon]:p-0">
			<Avatar className="size-7 shrink-0">
				<AccountButton />
			</Avatar>
			<div className="flex min-w-0 flex-col leading-tight group-data-[collapsible=icon]:hidden">
				<span className="truncate text-[13px] font-bold text-card-foreground">
					{name}
				</span>
			</div>
			{level ? (
				<Badge
					variant="secondary"
					className={cn(
						"chip chip-accent ml-auto px-2 py-0.5 text-[11px]",
						// The wrapper centres on the avatar alone, so a `shrink-0` badge
						// would re-introduce the off-centre gap once both are hidden.
						"group-data-[collapsible=icon]:hidden",
					)}
				>
					{LEVEL_LABELS[level]()}
				</Badge>
			) : null}
		</div>
	);
}
