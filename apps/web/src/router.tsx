import { RouteErrorFallback } from "#/components/common/error-pages";
import * as Sentry from "@sentry/tanstackstart-react";
import { createRouter as createTanStackRouter } from "@tanstack/react-router";
import { setupRouterSsrQueryIntegration } from "@tanstack/react-router-ssr-query";
import { getContext } from "./integrations/tanstack-query/root-provider";
import { routeTree } from "./routeTree.gen";

export function getRouter() {
  const context = getContext();

  const router = createTanStackRouter({
    routeTree,
    context,
    scrollRestoration: true,
    defaultPreload: "intent",
    defaultPreloadStaleTime: 0,
    defaultPreloadDelay: 250,
    // Router-wide backstop: without this, any match the per-route chain
    // doesn't cover renders the framework's default "Something went
    // wrong!" UI. Every boundary now shows the app error page.
    defaultErrorComponent: RouteErrorFallback,
  });

  setupRouterSsrQueryIntegration({ router, queryClient: context.queryClient });

  if (!router.isServer) {
    Sentry.addIntegration(
      Sentry.tanstackRouterBrowserTracingIntegration(router),
    );
  }

  return router;
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof getRouter>;
  }
}
