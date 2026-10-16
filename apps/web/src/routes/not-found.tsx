import { createFileRoute } from "@tanstack/react-router";
import { ErrorComponent } from "#/components/common/error-component";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/not-found")({
	component: () => (
		<ErrorComponent
			badge={m["errors.notFound.badge"]()}
			mascot="confused"
			caption={m["errors.notFound.caption"]()}
			title={m["errors.notFound.title"]()}
			body={m["errors.notFound.body"]()}
			primaryLabel={m["errors.notFound.primary"]()}
		/>
	),
});
