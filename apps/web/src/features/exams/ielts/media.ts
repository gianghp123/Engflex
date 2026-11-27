import { APP_ROUTES } from "#/app/app-route";

export function isIeltsMediaKey(key: string): boolean {
	const match =
		/^ielts\/([a-z0-9]+(?:-[a-z0-9]+)*)\/([a-z0-9]+(?:-[a-z0-9]+)*)\/(audio|images)\/([a-z0-9]+(?:-[a-z0-9]+)*)\.(mp3|jpg|png|webp|gif)$/.exec(
			key,
		);
	if (!match) return false;
	const [, , test, folder, name, extension] = match;
	return (
		(folder === "audio") === (extension === "mp3") &&
		name.startsWith(`${test}-`)
	);
}

export function ieltsMediaSource(
	identity: string | null | undefined,
): string | null {
	if (!identity || !isIeltsMediaKey(identity)) return null;
	return APP_ROUTES.MEDIA.DETAIL.replace("$", identity);
}
