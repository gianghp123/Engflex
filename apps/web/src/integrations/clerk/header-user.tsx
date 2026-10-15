import { Show, SignInButton } from "@clerk/tanstack-react-start";
import { AccountButton } from "#/integrations/clerk/account-button";

export default function HeaderUser() {
	return (
		<>
			<Show when="signed-in">
				<AccountButton />
			</Show>
			<Show when="signed-out">
				<SignInButton />
			</Show>
		</>
	);
}
