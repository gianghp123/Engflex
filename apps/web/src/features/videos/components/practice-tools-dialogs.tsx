import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { Field, FieldLabel } from "#/components/ui/field";
import { Switch } from "#/components/ui/switch";
import { VIDEO_SPEEDS, type VideoPreferences } from "#/features/videos/store";
import { m } from "#/paraglide/messages";

export function PracticeToolsDialogs({
	mode,
	open,
	onOpenChange,
	settings,
	onChange,
	onReset,
}: {
	mode: "dictation" | "shadowing";
	open: "settings" | "shortcuts" | null;
	onOpenChange: (open: "settings" | "shortcuts" | null) => void;
	settings: VideoPreferences;
	onChange: (update: Partial<VideoPreferences>) => void;
	onReset: () => void;
}) {
	const toggles =
		mode === "dictation"
			? [
					{
						key: "showTranscript" as const,
						label: m["videos.practice.preferences.transcript"](),
					},
					{
						key: "showMedia" as const,
						label: m["videos.practice.preferences.video"](),
					},
				]
			: [
					{
						key: "showTranscript" as const,
						label: m["videos.practice.preferences.transcript"](),
					},
					{
						key: "autoPause" as const,
						label: m["videos.shadowing.controls.autoPause"](),
					},
					{
						key: "largeVideo" as const,
						label: m["videos.practice.preferences.largeVideo"](),
					},
					{
						key: "showPhonetic" as const,
						label: m["videos.shadowing.transcript.phonetic"](),
					},
					{
						key: "showTranslation" as const,
						label: m["videos.shadowing.transcript.translation"](),
					},
				];
	return (
		<Dialog
			open={open !== null}
			onOpenChange={(value) => {
				if (!value) onOpenChange(null);
			}}
		>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>
						{open === "settings"
							? m["videos.practice.player.settings"]()
							: m["videos.practice.player.shortcuts"]()}
					</DialogTitle>
					<DialogDescription>
						{open === "settings"
							? m["videos.practice.preferences.description"]()
							: m["videos.practice.shortcuts.focus"]()}
					</DialogDescription>
				</DialogHeader>
				{open === "settings" ? (
					<div className="flex flex-col gap-4">
						<Field>
							<FieldLabel htmlFor="practice-speed">
								{m["videos.practice.preferences.speed"]()}
							</FieldLabel>
							<select
								id="practice-speed"
								className="field"
								value={settings.speed}
								onChange={(e) => onChange({ speed: Number(e.target.value) })}
							>
								{VIDEO_SPEEDS.map((speed) => (
									<option key={speed} value={speed}>
										{speed}x
									</option>
								))}
							</select>
						</Field>
						{toggles.map(({ key, label }) => (
							<label
								key={key}
								htmlFor={`practice-${key}`}
								className="flex justify-between gap-3"
							>
								{label}
								<Switch
									id={`practice-${key}`}
									checked={settings[key]}
									onCheckedChange={(value) => onChange({ [key]: value })}
								/>
							</label>
						))}
						<Button variant="outline" onClick={onReset}>
							{m["videos.practice.preferences.reset"]()}
						</Button>
					</div>
				) : (
					<dl className="grid grid-cols-[auto_1fr] gap-3">
						<dt>
							<kbd>Tab</kbd>
						</dt>
						<dd>{m["videos.practice.shortcuts.tab"]()}</dd>
						<dt>
							<kbd>Enter / Space</kbd>
						</dt>
						<dd>{m["videos.practice.shortcuts.activate"]()}</dd>
						{mode === "dictation" && (
							<>
								<dt>
									<kbd>Space / Enter</kbd>
								</dt>
								<dd>{m["videos.practice.shortcuts.submit"]()}</dd>
								<dt>
									<kbd>Backspace</kbd>
								</dt>
								<dd>{m["videos.practice.shortcuts.backspace"]()}</dd>
							</>
						)}
						<dt>
							<kbd>Esc</kbd>
						</dt>
						<dd>{m["videos.practice.shortcuts.close"]()}</dd>
					</dl>
				)}
			</DialogContent>
		</Dialog>
	);
}
