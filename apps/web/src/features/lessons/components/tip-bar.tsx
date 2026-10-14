import { TipCard } from "#/components/common/tip-card";
import { m } from "#/paraglide/messages";

/** Back-compat alias: the lessons detail page renders its calm quote via the shared card. */
export function TipBar({ tip }: { tip: string }) {
	return <TipCard title={m["lessons.detail.calmTip"]()} body={`“${tip}”`} />;
}
