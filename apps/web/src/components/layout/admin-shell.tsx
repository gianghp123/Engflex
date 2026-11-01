import type { CSSProperties, ReactNode } from "react";
import { AdminSidebar } from "#/components/layout/admin-sidebar";
import { Topbar } from "#/components/layout/topbar";
import { SidebarInset, SidebarProvider } from "#/components/ui/sidebar";

export function AdminShell({ children }: { children: ReactNode }) {
	return (
		<SidebarProvider style={{ "--sidebar-width": "15rem" } as CSSProperties}>
			<AdminSidebar />
			<SidebarInset className="app-canvas">
				<Topbar />
				<div className="flex flex-1 flex-col">{children}</div>
			</SidebarInset>
		</SidebarProvider>
	);
}
