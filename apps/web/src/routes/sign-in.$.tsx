import { SignIn } from "@clerk/tanstack-react-start";
import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";

export const Route = createFileRoute("/sign-in/$")({
	component: SignInPage,
});

function SignInPage() {
	return (
		<div className="flex min-h-svh items-center justify-center bg-dot-pattern p-6">
			<SignIn fallbackRedirectUrl={APP_ROUTES.HOME} />
		</div>
	);
}
