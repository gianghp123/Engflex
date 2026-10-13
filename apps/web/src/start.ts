import { clerkMiddleware } from "@clerk/tanstack-react-start/server";
import {
	sentryGlobalFunctionMiddleware,
	sentryGlobalRequestMiddleware,
	wrapMiddlewaresWithSentry,
} from "@sentry/tanstackstart-react";
import { createCsrfMiddleware, createStart } from "@tanstack/react-start";
import { paraglideRequestMiddleware } from "./lib/paraglide-middleware";

const csrfMiddleware = createCsrfMiddleware({
	filter: (context) => context.handlerType === "serverFn",
});

const [
	tracedParaglideRequestMiddleware,
	tracedCsrfMiddleware,
	tracedClerkMiddleware,
] = wrapMiddlewaresWithSentry({
	paraglideRequestMiddleware,
	csrfMiddleware,
	clerkMiddleware: clerkMiddleware(),
});

export const startInstance = createStart(() => ({
	requestMiddleware: [
		sentryGlobalRequestMiddleware,
		tracedParaglideRequestMiddleware,
		tracedCsrfMiddleware,
		tracedClerkMiddleware,
	],
	functionMiddleware: [sentryGlobalFunctionMiddleware],
}));
