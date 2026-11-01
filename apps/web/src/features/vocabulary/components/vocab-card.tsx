import type { VocabularyItem } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { cn } from "cn";
import { Check, CircleCheck } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { AudioButton } from "#/components/common/audio-button";
import { CefrBadge } from "#/components/common/cefr-badge";
import { PremiumBadge } from "#/components/common/premium-badge";
import { SourcePill } from "#/components/common/source-pill";
import { Button } from "#/components/ui/button";
import { useCanAccessPremium } from "#/features/payments/queries";
import { isDue } from "#/features/vocabulary/dates";
import { m } from "#/paraglide/messages";

export function VocabCard({
	item,
	onToggleMastered,
}: {
	item: VocabularyItem;
	onToggleMastered: () => void;
}) {
	const state = item.userState;
	const mastered = state?.mastered ?? false;
	const due = isDue(state?.srsDueAt);
	const context = item.details?.contexts?.[0];
	const canAccessPremium = useCanAccessPremium();
	const premiumBlocked = item.isPremium && !canAccessPremium;

	return (
		<article className="surface-card flex flex-col gap-3 p-5">
			<div className="flex items-start justify-between gap-2">
				<div className="min-w-0">
					<h3 className="text-base font-bold text-foreground">{item.term}</h3>
					<p className="font-mono text-xs text-muted-foreground">{item.ipa}</p>
				</div>
				{mastered ? (
					<span className="chip shrink-0 border-cefr-b1-border bg-cefr-b1-bg px-2 py-0.5 text-xs text-cefr-b1">
						<Check className="size-3.5" />
						{m["vocabulary.card.mastered"]()}
					</span>
				) : due ? (
					<button
						type="button"
						onClick={onToggleMastered}
						className="chip chip-accent shrink-0 px-2 py-0.5 text-xs"
					>
						<CircleCheck className="size-3.5" />
						{m["vocabulary.card.markMastered"]()}
					</button>
				) : (
					<button
						type="button"
						onClick={onToggleMastered}
						className="chip shrink-0 px-2 py-0.5 text-xs hover:border-primary"
					>
						<Check className="size-3.5" />
						{m["vocabulary.card.markMastered"]()}
					</button>
				)}
			</div>

			<div className="flex flex-wrap items-center gap-1.5">
				{item.isPremium ? <PremiumBadge /> : null}
				{item.partOfSpeech ? (
					<span className="chip bg-accent px-1.5 py-0 text-[11px]">
						{item.partOfSpeech}
					</span>
				) : null}
				{state ? <SourcePill source={state.sourceType} /> : null}
			</div>

			<p className="text-[15px] font-medium text-muted-foreground">
				{item.definition}
			</p>

			{context ? (
				<div className="bubble bubble-partner text-[15px]">
					“{context.quote}”
				</div>
			) : null}

			<div className="mt-auto flex items-center gap-2 pt-1">
				<AudioButton
					label={m["vocabulary.card.playTerm"]({ term: item.term })}
				/>
				<CefrBadge value={item.cefr} />
				{state?.sourceId ? (
					<span className="truncate text-[11px] font-medium text-muted-foreground">
						{state.sourceType === "lesson"
							? m["vocabulary.sourceLine.lesson"]({ id: state.sourceId })
							: m["vocabulary.sourceLine.note"]({ id: state.sourceId })}
					</span>
				) : null}
				<Button asChild size="sm" className="btn btn-primary ml-auto">
					{premiumBlocked ? (
						<Link to={APP_ROUTES.PAYMENTS.PRICING}>
							{m["common.premium.upgradeAction"]()}
						</Link>
					) : (
						<Link
							to={APP_ROUTES.VOCABULARY.DETAIL}
							params={{ itemId: item.id }}
						>
							{m["vocabulary.card.review"]()}
						</Link>
					)}
				</Button>
			</div>

			<span
				className={cn(
					"text-[11px] font-bold",
					mastered
						? "text-accuracy"
						: due
							? "text-ai-coral"
							: "text-muted-foreground",
				)}
			>
				{mastered
					? m["vocabulary.card.retained"]()
					: due
						? m["vocabulary.card.dueToday"]()
						: state?.srsDueAt?.slice(0, 10)}
			</span>
		</article>
	);
}
