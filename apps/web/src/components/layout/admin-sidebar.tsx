import { UserButton } from "@clerk/tanstack-react-start";
import { Link, useRouterState } from "@tanstack/react-router";
import { cn } from "cn";
import {
	ArrowLeft,
	BookmarkCheck,
	GraduationCap,
	LayoutDashboard,
	MessageSquare,
	Users,
	Video,
} from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { Logo } from "#/components/common/logo";
import { Avatar } from "#/components/ui/avatar";
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
import { m } from "#/paraglide/messages";

const NAV_ITEMS = [
	{
		title: () => m["admin.nav.overview"](),
		to: APP_ROUTES.ADMIN.OVERVIEW,
		icon: LayoutDashboard,
		tile: "bg-[#6c4df0]",
	},
	{
		title: () => m["admin.nav.vocabulary"](),
		to: APP_ROUTES.ADMIN.VOCABULARY,
		icon: BookmarkCheck,
		tile: "bg-[#059669]",
	},
	{
		title: () => m["admin.nav.videos"](),
		to: APP_ROUTES.ADMIN.VIDEOS,
		icon: Video,
		tile: "bg-[#2563eb]",
	},
	{
		title: () => m["admin.nav.scenarios"](),
		to: APP_ROUTES.ADMIN.SCENARIOS,
		icon: MessageSquare,
		tile: "bg-[#ea580c]",
	},
	{
		title: () => m["admin.nav.lessons"](),
		to: APP_ROUTES.ADMIN.LESSONS,
		icon: GraduationCap,
		tile: "bg-[#dc2626]",
	},
	{
		title: () => m["admin.nav.users"](),
		to: APP_ROUTES.ADMIN.USERS,
		icon: Users,
		tile: "bg-[#0891b2]",
	},
] as const;

export function AdminSidebar() {
	const pathname = useRouterState({
		select: (state) => state.location.pathname,
	});

	return (
		<Sidebar collapsible="icon">
			<SidebarHeader className="group-data-[collapsible=icon]:px-0">
				<Link
					to={APP_ROUTES.ADMIN.OVERVIEW}
					className="flex h-12 items-center rounded-md px-2 hover:bg-transparent group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:p-0"
				>
					<Logo
						variant="full"
						className="text-sidebar-foreground group-data-[collapsible=icon]:[&>span:last-child]:hidden"
					/>
				</Link>
			</SidebarHeader>
			<SidebarContent className="group-data-[collapsible=icon]:gap-1">
				<SidebarGroup className="group-data-[collapsible=icon]:p-0">
					<SidebarGroupLabel>{m["admin.overview.title"]()}</SidebarGroupLabel>
					<SidebarMenu>
						{NAV_ITEMS.map((item) => {
							const isActive =
								item.to === APP_ROUTES.ADMIN.OVERVIEW
									? pathname === APP_ROUTES.ADMIN.OVERVIEW
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
			</SidebarContent>
			<SidebarFooter>
				<SidebarMenu>
					<SidebarMenuItem>
						<SidebarMenuButton asChild tooltip={m["nav.item.home"]()}>
							<Link to={APP_ROUTES.HOME} className="font-medium">
								<ArrowLeft data-icon="inline-start" />
								<span>{m["nav.item.home"]()}</span>
							</Link>
						</SidebarMenuButton>
					</SidebarMenuItem>
				</SidebarMenu>
				<div className="flex items-center gap-2.5 rounded-[16px] border-2 border-sidebar-border bg-card p-2 group-data-[collapsible=icon]:justify-center group-data-[collapsible=icon]:rounded-full group-data-[collapsible=icon]:border-0 group-data-[collapsible=icon]:bg-transparent group-data-[collapsible=icon]:p-0">
					<Avatar className="size-7 shrink-0">
						<UserButton />
					</Avatar>
				</div>
			</SidebarFooter>
		</Sidebar>
	);
}
