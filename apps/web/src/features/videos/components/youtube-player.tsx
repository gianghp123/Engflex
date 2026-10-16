import { cn } from "cn";
import { forwardRef, useEffect, useImperativeHandle, useRef } from "react";

export interface YTPlayerInstance {
	playVideo: () => void;
	pauseVideo: () => void;
	seekTo: (seconds: number, allowSeekAhead?: boolean) => void;
	getCurrentTime: () => number;
	getDuration: () => number;
	setPlaybackRate: (suggestedRate: number) => void;
	getPlaybackRate: () => number;
	getPlayerState: () => number;
	destroy: () => void;
}

export interface YouTubePlayerRef {
	playSegment: (start: number, end?: number) => void;
	play: () => void;
	pause: () => void;
	seekTo: (seconds: number) => void;
	setSpeed: (speed: number) => void;
	getCurrentTime: () => number;
}

export interface YouTubePlayerProps {
	videoId: string;
	onReady?: () => void;
	onTimeUpdate?: (currentTime: number) => void;
	onSegmentEnd?: () => void;
	className?: string;
}

declare global {
	interface Window {
		YT?: {
			Player: new (
				elementId: HTMLElement,
				config: {
					videoId?: string;
					playerVars?: {
						autoplay?: number;
						controls?: number;
						disablekb?: number;
						enablejsapi?: number;
						fs?: number;
						modestbranding?: number;
						playsinline?: number;
						rel?: number;
						origin?: string;
					};
					events?: {
						onReady?: (event: { target: YTPlayerInstance }) => void;
						onStateChange?: (event: { data: number }) => void;
						onError?: (event: { data: number }) => void;
					};
				},
			) => YTPlayerInstance;
			PlayerState?: {
				UNSTARTED: number;
				ENDED: number;
				PLAYING: number;
				PAUSED: number;
				BUFFERING: number;
				CUED: number;
			};
		};
		onYouTubeIframeAPIReady?: () => void;
	}
}

export const YouTubePlayer = forwardRef<YouTubePlayerRef, YouTubePlayerProps>(
	function YouTubePlayer(
		{ videoId, onReady, onTimeUpdate, onSegmentEnd, className },
		ref,
	) {
		const containerRef = useRef<HTMLDivElement>(null);
		const playerInstanceRef = useRef<YTPlayerInstance | null>(null);
		const segmentEndRef = useRef<number | null>(null);
		const intervalRef = useRef<number | null>(null);

		useImperativeHandle(
			ref,
			() => ({
				playSegment: (start: number, end?: number) => {
					segmentEndRef.current = end !== undefined ? end : null;
					const player = playerInstanceRef.current;
					if (player) {
						player.seekTo(start, true);
						player.playVideo();
					}
				},
				play: () => {
					playerInstanceRef.current?.playVideo();
				},
				pause: () => {
					segmentEndRef.current = null;
					playerInstanceRef.current?.pauseVideo();
				},
				seekTo: (seconds: number) => {
					playerInstanceRef.current?.seekTo(seconds, true);
				},
				setSpeed: (speed: number) => {
					playerInstanceRef.current?.setPlaybackRate(speed);
				},
				getCurrentTime: () => {
					return playerInstanceRef.current?.getCurrentTime() ?? 0;
				},
			}),
			[],
		);

		useEffect(() => {
			let isMounted = true;

			const startTracking = () => {
				if (intervalRef.current !== null) return;
				intervalRef.current = window.setInterval(() => {
					const player = playerInstanceRef.current;
					if (!player) return;
					const current = player.getCurrentTime();
					onTimeUpdate?.(current);
					if (
						segmentEndRef.current !== null &&
						current >= segmentEndRef.current
					) {
						segmentEndRef.current = null;
						player.pauseVideo();
						onSegmentEnd?.();
					}
				}, 100);
			};

			const stopTracking = () => {
				if (intervalRef.current !== null) {
					window.clearInterval(intervalRef.current);
					intervalRef.current = null;
				}
			};

			const initPlayer = () => {
				if (!containerRef.current || !window.YT?.Player) return;
				playerInstanceRef.current?.destroy();

				const player = new window.YT.Player(containerRef.current, {
					videoId,
					playerVars: {
						autoplay: 0,
						controls: 1,
						enablejsapi: 1,
						playsinline: 1,
						rel: 0,
						origin: window.location.origin,
					},
					events: {
						onReady: () => {
							if (!isMounted) return;
							onReady?.();
						},
						onStateChange: (event) => {
							if (!isMounted) return;
							if (event.data === window.YT?.PlayerState?.PLAYING) {
								startTracking();
							} else {
								stopTracking();
							}
						},
					},
				});
				playerInstanceRef.current = player;
			};

			if (window.YT?.Player) {
				initPlayer();
			} else {
				const existingScript = document.getElementById("yt-iframe-api");
				if (!existingScript) {
					const script = document.createElement("script");
					script.id = "yt-iframe-api";
					script.src = "https://www.youtube.com/iframe_api";
					document.body.appendChild(script);
				}

				const previousReady = window.onYouTubeIframeAPIReady;
				window.onYouTubeIframeAPIReady = () => {
					previousReady?.();
					if (isMounted) {
						initPlayer();
					}
				};
			}

			return () => {
				isMounted = false;
				stopTracking();
				playerInstanceRef.current?.destroy();
				playerInstanceRef.current = null;
			};
		}, [videoId, onReady, onTimeUpdate, onSegmentEnd]);

		return (
			<div
				className={cn(
					"relative aspect-video w-full overflow-hidden bg-black",
					className,
				)}
			>
				<div ref={containerRef} className="size-full" />
			</div>
		);
	},
);
