/** Current sentence as words with their IPA underneath. */
export function WordIpaRow({
	words,
}: {
	words: { word: string; ipa?: string }[];
}) {
	return (
		<ul className="flex flex-wrap items-start justify-center gap-x-5 gap-y-3 px-2">
			{words.map(({ word, ipa }, index) => (
				// biome-ignore lint/suspicious/noArrayIndexKey: words repeat within a sentence; position is the identity.
				<li key={index} className="flex flex-col items-center gap-1">
					<span className="border-b-2 border-border px-0.5 text-xl leading-8 font-bold text-foreground md:text-2xl">
						{word}
					</span>
					{ipa ? (
						<span className="font-mono text-xs text-muted-foreground">
							/{ipa}/
						</span>
					) : null}
				</li>
			))}
		</ul>
	);
}
