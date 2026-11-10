import { useState } from "react";
import { SubmitButton } from "#/components/common/submit-button";
import { m } from "#/paraglide/messages";

export function RetryButton() {
	const [retrying, setRetrying] = useState(false);

	return (
		<SubmitButton
			type="button"
			variant="outline"
			pending={retrying}
			onClick={() => {
				setRetrying(true);
				window.location.reload();
			}}
		>
			{m["common.actions.retry"]()}
		</SubmitButton>
	);
}
