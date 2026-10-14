import { MoMascot } from "#/components/common/mo-mascot";
import {
	Empty,
	EmptyDescription,
	EmptyHeader,
	EmptyTitle,
} from "#/components/ui/empty";
import { m } from "#/paraglide/messages";

export function VideoHubEmpty() {
	return (
		<Empty className="surface-card items-center text-center">
			<MoMascot variant="confused" size={72} />
			<EmptyHeader>
				<EmptyTitle>{m["videos.hub.empty.title"]()}</EmptyTitle>
				<EmptyDescription>{m["videos.hub.empty.body"]()}</EmptyDescription>
			</EmptyHeader>
		</Empty>
	);
}
