import { UserButton, useUser } from "@clerk/tanstack-react-start";
import { Link, useRouterState } from "@tanstack/react-router";
import { useStore } from "@tanstack/react-store";
import { cn } from "cn";
import {
	AudioWaveform,
	BookmarkCheck,
	BookOpen,
	House,
	Info,
} from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { Logo } from "#/components/common/logo";
import { Avatar } from "#/components/ui/avatar";
import { Badge } from "#/components/ui/badge";
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
} from "#/components/ui/sidebar";
import { PROGRESS_SUMMARY } from "#/features/attempts/fixtures";
import { profileStore } from "#/features/profiles/store";
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
				to: APP_ROUTES.VOCABULARY.LIST,
				icon: BookmarkCheck,
				tile: "bg-nav-vocab",
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
							{group.items.map((item) => {
								const isActive =
									item.to === APP_ROUTES.HOME
										? pathname === APP_ROUTES.HOME
										: pathname.startsWith(item.to);
								return (
									<SidebarMenuItem key={item.title()}>
										<SidebarMenuButton
											asChild
											isActive={isActive}
											tooltip={item.title()}
										>
											<Link to={item.to} className="font-medium">
												<span
													className={cn(
														"flex size-6 shrink-0 items-center justify-center rounded-[7px] text-white",
														item.tile,
													)}
												>
													<item.icon className="size-3.5" />
												</span>
												<span className="text-sidebar-foreground">
													{item.title()}
												</span>
												{isActive ? (
													<span className="ml-auto size-1.5 rounded-full bg-sidebar-primary group-data-[collapsible=icon]:hidden" />
												) : null}
											</Link>
										</SidebarMenuButton>
									</SidebarMenuItem>
								);
							})}
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

function SidebarUserCard() {
	const { user } = useUser();
	const level = useStore(
		profileStore,
		(state) => state.profile.preferences.level,
	);

	const name = user?.fullName ?? user?.firstName ?? m["nav.user.learner"]();

	return (
		<div className="flex items-center gap-2.5 rounded-[16px] border-2 border-sidebar-border bg-card p-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:rounded-full group-data-[collapsible=icon]:border-0 group-data-[collapsible=icon]:bg-transparent group-data-[collapsible=icon]:p-0">
			<Avatar className="size-7 shrink-0">
				<UserButton />
			</Avatar>
			<div className="flex min-w-0 flex-col leading-tight group-data-[collapsible=icon]:hidden">
				<span className="truncate text-[13px] font-bold text-card-foreground">
					{name}
				</span>
				<span className="truncate text-[11px] font-medium text-muted-foreground">
					{m["progress.streak"]({ count: PROGRESS_SUMMARY.streakDays })}
				</span>
			</div>
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
		</div>
	);
}
