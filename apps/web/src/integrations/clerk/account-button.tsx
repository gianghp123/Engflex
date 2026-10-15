import { UserButton } from "@clerk/tanstack-react-start";
import { UserRound } from "lucide-react";
import { useState } from "react";
import { ProfileDialog } from "#/features/profiles/components/profile-dialog";
import { m } from "#/paraglide/messages";

/**
 * Clerk `UserButton` with our own "Saved profile" menu item. The item runs a
 * click handler that opens {@link ProfileDialog} — our modal — instead of
 * Clerk's account modal. The default `manageAccount` / `signOut` items are
 * listed explicitly so they keep their order around ours.
 */
export function AccountButton() {
	const [profileOpen, setProfileOpen] = useState(false);
	return (
		<>
			<UserButton>
				<UserButton.MenuItems>
					<UserButton.Action
						label={m["profile.menu.label"]()}
						labelIcon={<UserRound />}
						onClick={() => setProfileOpen(true)}
					/>
					<UserButton.Action label="manageAccount" />
					<UserButton.Action label="signOut" />
				</UserButton.MenuItems>
			</UserButton>
			<ProfileDialog open={profileOpen} onOpenChange={setProfileOpen} />
		</>
	);
}
