// Bucket-qualified identity; the media forward strips the fixed bucket prefix.
// Full exam JSON is deliberately excluded from public media requests.
export function isToeicMediaKey(key: string): boolean {
	const match =
		/^toeic\/([a-z0-9]+(?:-[a-z0-9]+)*)\/([a-z0-9]+(?:-[a-z0-9]+)*)\/(images|audio)\/([a-z0-9]+(?:-[a-z0-9]+)*)\.(jpg|png|webp|gif|mp3)$/.exec(
			key,
		);
	if (!match) return false;
	const [, , test, folder, name, extension] = match;
	if ((folder === "audio") !== (extension === "mp3")) return false;
	return name.startsWith(`${test}-`) || /^[a-f0-9]{64}$/.test(name);
}
