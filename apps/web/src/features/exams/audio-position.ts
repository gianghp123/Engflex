const key = (userId: string, attemptId: string) =>
	`engflex:toeic-audio:v1:${userId}:${attemptId}`;

export function readAudioPosition(
	userId: string,
	attemptId: string,
	sha256: string,
): number | null {
	try {
		const saved = JSON.parse(
			localStorage.getItem(key(userId, attemptId)) ?? "null",
		);
		return saved?.sha256 === sha256 &&
			typeof saved.position === "number" &&
			Number.isFinite(saved.position) &&
			saved.position >= 0
			? saved.position
			: null;
	} catch {
		return null;
	}
}

export function saveAudioPosition(
	userId: string,
	attemptId: string,
	sha256: string,
	position: number,
) {
	if (!Number.isFinite(position) || position < 0) return;
	try {
		localStorage.setItem(
			key(userId, attemptId),
			JSON.stringify({ sha256, position }),
		);
	} catch {
		/* Playback remains available when storage is blocked. */
	}
}

export function clearAudioPosition(userId: string, attemptId: string) {
	try {
		localStorage.removeItem(key(userId, attemptId));
	} catch {
		/* Storage is optional. */
	}
}
