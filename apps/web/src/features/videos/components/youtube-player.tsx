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
	getAvailablePlaybackRates: () => number[];
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
	getAvailableRates: () => number[];
}

export interface YouTubePlayerProps {
	videoId: string;
	activeSegment?: { start: number; end: number } | null;
	onReady?: () => void;
	onTimeUpdate?: (currentTime: number) => void;
	onSegmentEnd?: () => void;
	onPlayPastEnd?: () => void;
	onPlayingChange?: (playing: boolean) => void;
	onRateChange?: (rate: number) => void;
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
						onPlaybackRateChange?: (event: { data: number }) => void;
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
		{
			videoId,
			activeSegment,
			onReady,
			onTimeUpdate,
			onSegmentEnd,
			onPlayPastEnd,
			onPlayingChange,
			onRateChange,
			className,
		},
		ref,
	) {
		const wrapperRef = useRef<HTMLDivElement>(null);
		const playerInstanceRef = useRef<YTPlayerInstance | null>(null);
		const segmentStartRef = useRef<number | null>(null);
		const segmentEndRef = useRef<number | null>(null);
		const isSeekingRef = useRef<boolean>(false);
		const intervalRef = useRef<number | null>(null);

		const callbacksRef = useRef({
			onReady,
			onTimeUpdate,
			onSegmentEnd,
			onPlayPastEnd,
			onPlayingChange,
			onRateChange,
			activeSegment,
		});
		useEffect(() => {
			callbacksRef.current = {
				onReady,
				onTimeUpdate,
				onSegmentEnd,
				onPlayPastEnd,
				onPlayingChange,
				onRateChange,
				activeSegment,
			};
		});

		useImperativeHandle(
			ref,
			() => ({
				playSegment: (start: number, end?: number) => {
					segmentStartRef.current = start;
					segmentEndRef.current = end !== undefined ? end : null;
					isSeekingRef.current = true;
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
					segmentStartRef.current = null;
					isSeekingRef.current = false;
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
				getAvailableRates: () =>
					playerInstanceRef.current?.getAvailablePlaybackRates() ?? [],
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
					callbacksRef.current.onTimeUpdate?.(current);

					if (isSeekingRef.current && segmentStartRef.current !== null) {
						if (current >= segmentStartRef.current - 0.2) {
							isSeekingRef.current = false;
						}
					}

					const effectiveEnd =
						segmentEndRef.current ??
						callbacksRef.current.activeSegment?.end ??
						null;

					if (
						!isSeekingRef.current &&
						effectiveEnd !== null &&
						current >= effectiveEnd
					) {
						segmentEndRef.current = null;
						segmentStartRef.current = null;
						player.pauseVideo();
						callbacksRef.current.onSegmentEnd?.();
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
				if (!wrapperRef.current || !window.YT?.Player) return;
				playerInstanceRef.current?.destroy();
				playerInstanceRef.current = null;

				wrapperRef.current.innerHTML = "";
				const targetDiv = document.createElement("div");
				targetDiv.className = "size-full";
				wrapperRef.current.appendChild(targetDiv);

				const player = new window.YT.Player(targetDiv, {
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
							callbacksRef.current.onReady?.();
						},
						onStateChange: (event) => {
							if (!isMounted) return;
							callbacksRef.current.onPlayingChange?.(
								event.data === window.YT?.PlayerState?.PLAYING ||
									event.data === window.YT?.PlayerState?.BUFFERING,
							);
							if (event.data === window.YT?.PlayerState?.PLAYING) {
								const current =
									playerInstanceRef.current?.getCurrentTime() ?? 0;
								const seg = callbacksRef.current.activeSegment;
								if (seg) {
									if (!isSeekingRef.current && current >= seg.end - 0.3) {
										playerInstanceRef.current?.pauseVideo();
										callbacksRef.current.onPlayPastEnd?.();
										return;
									}
									if (
										segmentEndRef.current === null ||
										segmentEndRef.current > seg.end
									) {
										segmentEndRef.current = seg.end;
										segmentStartRef.current = seg.start;
									}
								}
								startTracking();
							} else {
								stopTracking();
							}
						},
						onPlaybackRateChange: (event) => {
							if (isMounted) callbacksRef.current.onRateChange?.(event.data);
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
		}, [videoId]);

		return (
			<div
				ref={wrapperRef}
				className={cn(
					"relative aspect-video w-full overflow-hidden bg-black",
					className,
				)}
			/>
		);
	},
);
