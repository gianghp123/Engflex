import type { VideoTranscriptResponse } from "@engflex/contracts";
import type { checkDictation } from "#/features/videos/dictation";
import { m } from "#/paraglide/messages";

export function DictationNotes({
	line,
	translation,
	phonetic,
}: {
	line: VideoTranscriptResponse;
	translation: boolean;
	phonetic: boolean;
}) {
	if ((!translation || !line.vietnamese) && (!phonetic || !line.phonetic))
		return null;
	return (
		<div className="flex flex-col gap-1 text-sm">
			{translation && line.vietnamese ? (
				<p>
					<strong>{m["videos.dictation.preferences.translation"]()}</strong>{" "}
					<span lang="vi">{line.vietnamese}</span>
				</p>
			) : null}
			{phonetic && line.phonetic ? (
				<p>
					<strong>{m["videos.dictation.preferences.phonetic"]()}</strong>{" "}
					<span className="font-mono">{line.phonetic}</span>
				</p>
			) : null}
		</div>
	);
}

export function DictationWordResults({
	slots,
	revealed,
	scored,
}: {
	slots: ReturnType<typeof checkDictation>["slots"];
	revealed: number[];
	scored: boolean;
}) {
	return (
		<ul className="flex flex-wrap gap-2">
			{slots.map(({ word, blankIndex, position }) => (
				<li key={position} className="chip px-2 py-1 text-sm">
					<span lang="en">{word}</span>
					{blankIndex !== undefined && scored ? (
						<span className="ml-1 text-xs text-muted-foreground">
							{revealed.includes(position)
								? m["videos.dictation.feedback.revealed"]()
								: m["videos.dictation.feedback.correct"]()}
						</span>
					) : null}
				</li>
			))}
		</ul>
	);
}
