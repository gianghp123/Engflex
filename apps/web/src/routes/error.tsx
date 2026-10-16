import { createFileRoute } from "@tanstack/react-router";
import { ErrorComponent } from "#/components/common/error-component";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/error")({
	component: () => (
		<ErrorComponent
			badge={m["errors.server.badge"]()}
			mascot="cry"
			caption={m["errors.server.caption"]()}
			title={m["errors.server.title"]()}
			body={m["errors.server.body"]()}
			primaryLabel={m["common.actions.backHome"]()}
		/>
	),
});
