import { Settings } from "lucide-react";
import { Button } from "#/components/ui/button";
import { Field, FieldLabel } from "#/components/ui/field";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "#/components/ui/popover";
import { Switch } from "#/components/ui/switch";
import { readerPreferences } from "#/features/stories/store";
import { m } from "#/paraglide/messages";

export function StoryReaderSettings() {
	const [settings, set] = readerPreferences.usePreferences();
	return (
		<Popover>
			<PopoverTrigger asChild>
				<Button
					variant="outline"
					size="icon"
					aria-label={m["stories.reader.settings.title"]()}
				>
					<Settings className="size-4" />
				</Button>
			</PopoverTrigger>
			<PopoverContent className="flex w-72 flex-col gap-4" align="end">
				<h2 className="font-semibold">
					{m["stories.reader.settings.title"]()}
				</h2>
				<Field>
					<FieldLabel htmlFor="reader-font-size">
						{m["stories.reader.settings.fontSize"]()}
					</FieldLabel>
					<select
						id="reader-font-size"
						className="field"
						value={settings.fontSize}
						onChange={(e) => set({ fontSize: Number(e.target.value) })}
					>
						{[16, 18, 20, 24].map((size) => (
							<option key={size} value={size}>
								{size}px
							</option>
						))}
					</select>
				</Field>
				<Field>
					<FieldLabel htmlFor="reader-font">
						{m["stories.reader.settings.fontFamily"]()}
					</FieldLabel>
					<select
						id="reader-font"
						className="field"
						value={settings.fontFamily}
						onChange={(e) => set({ fontFamily: e.target.value })}
					>
						<option value="sans">{m["stories.reader.settings.sans"]()}</option>
						<option value="serif">
							{m["stories.reader.settings.serif"]()}
						</option>
					</select>
				</Field>
				<Field>
					<FieldLabel htmlFor="reader-line-height">
						{m["stories.reader.settings.lineHeight"]()}
					</FieldLabel>
					<select
						id="reader-line-height"
						className="field"
						value={settings.lineHeight}
						onChange={(e) => set({ lineHeight: Number(e.target.value) })}
					>
						{[1.5, 1.75, 2].map((height) => (
							<option key={height} value={height}>
								{height}
							</option>
						))}
					</select>
				</Field>
				<label
					className="flex items-center justify-between gap-3"
					htmlFor="reader-translation"
				>
					{m["stories.reader.showTranslation"]()}
					<Switch
						id="reader-translation"
						checked={settings.showTranslation}
						onCheckedChange={(showTranslation) => set({ showTranslation })}
					/>
				</label>
				<Button variant="outline" onClick={readerPreferences.reset}>
					{m["stories.reader.settings.reset"]()}
				</Button>
			</PopoverContent>
		</Popover>
	);
}
