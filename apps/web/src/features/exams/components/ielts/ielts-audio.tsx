import { useEffect, useRef, useState } from "react";
import { ieltsMediaSource } from "#/features/exams/ielts/media";
import { m } from "#/paraglide/messages";

export function IeltsAudio({ identity }: { identity: string }) {
	const ref = useRef<HTMLAudioElement>(null);
	const [failed, setFailed] = useState(false);
	const src = ieltsMediaSource(identity);
	useEffect(() => {
		const audio = ref.current;
		return () => {
			audio?.pause();
		};
	}, []);
	return (
		<div className="space-y-2 rounded-xl border border-border bg-card p-3">
			<p className="text-sm font-semibold">{m["exams.ielts.media.audio"]()}</p>
			{src && !failed ? (
				// biome-ignore lint/a11y/useMediaCaption: Listening transcripts are available in review; the test must withhold them while answering.
				<audio
					ref={ref}
					src={src}
					controls
					preload="metadata"
					className="w-full"
					onError={() => setFailed(true)}
					aria-label={m["exams.ielts.media.audio"]()}
				/>
			) : (
				<p role="alert" className="text-sm text-destructive">
					{m["exams.ielts.media.failed"]()}
				</p>
			)}
		</div>
	);
}
