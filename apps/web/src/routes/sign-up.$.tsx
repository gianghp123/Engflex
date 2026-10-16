import { SignUp } from "@clerk/tanstack-react-start";
import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";

export const Route = createFileRoute("/sign-up/$")({
	component: SignUpPage,
});

function SignUpPage() {
	return (
		<div className="flex min-h-svh items-center justify-center bg-background p-6">
			<SignUp fallbackRedirectUrl={APP_ROUTES.HOME} />
		</div>
	);
}
