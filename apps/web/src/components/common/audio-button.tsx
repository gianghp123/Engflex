import { cn } from "cn";
import { Loader2Icon, Volume2 } from "lucide-react";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";

/** Simulated playback: no audio files exist — a fixed 800ms playing state. */
export function AudioButton({
	label = m["common.audio.playPronunciation"](),
	className,
}: {
	label?: string;
	className?: string;
}) {
	const [playing, setPlaying] = useState(false);
	return (
		<Button
			type="button"
			variant="ghost"
			size="icon-sm"
			aria-label={label}
			className={cn("text-muted-foreground hover:text-primary", className)}
			onClick={() => {
				setPlaying(true);
				window.setTimeout(() => setPlaying(false), 800);
			}}
		>
			{playing ? <Loader2Icon className="animate-spin" /> : <Volume2 />}
		</Button>
	);
}
