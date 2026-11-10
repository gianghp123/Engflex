import { Keyboard, type LucideIcon, Settings, Zap } from "lucide-react";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";

export function ToolbarButton({
	icon: Icon,
	label,
	onClick,
}: {
	icon: LucideIcon;
	label: string;
	onClick?: () => void;
}) {
	return (
		<Button
			type="button"
			variant="ghost"
			size="icon"
			aria-label={label}
			title={label}
			onClick={onClick}
			disabled={!onClick}
		>
			<Icon />
		</Button>
	);
}

/** Speed / settings / shortcuts cluster shared by both practice modes. */
export function PlayerTools({
	speed,
	onSpeed,
	rates,
	onSelectSpeed,
	onSettings,
	onShortcuts,
}: {
	speed: number;
	onSpeed?: () => void;
	rates?: number[];
	onSelectSpeed?: (speed: number) => void;
	onSettings: () => void;
	onShortcuts: () => void;
}) {
	return (
		<div className="flex items-center gap-1">
			{rates ? (
				<select
					className="field h-8 w-auto"
					value={speed}
					aria-label={m["videos.practice.player.speed"]({ speed })}
					onChange={(event) => onSelectSpeed?.(Number(event.target.value))}
				>
					{rates.map((rate) => (
						<option key={rate} value={rate}>
							{rate}×
						</option>
					))}
				</select>
			) : (
				<Button
					type="button"
					variant="ghost"
					size="sm"
					onClick={onSpeed}
					aria-label={m["videos.practice.player.speed"]({ speed })}
					title={m["videos.practice.player.speed"]({ speed })}
					className="font-bold"
				>
					<Zap data-icon="inline-start" />
					{`${speed}x`}
				</Button>
			)}
			<ToolbarButton
				icon={Settings}
				label={m["videos.practice.player.settings"]()}
				onClick={onSettings}
			/>
			<ToolbarButton
				icon={Keyboard}
				label={m["videos.practice.player.shortcuts"]()}
				onClick={onShortcuts}
			/>
		</div>
	);
}
