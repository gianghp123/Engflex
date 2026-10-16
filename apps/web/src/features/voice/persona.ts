const TILE_TONES = [
	"bg-primary text-primary-foreground",
	"bg-secondary text-secondary-foreground",
	"bg-accent-violet text-primary-foreground",
] as const;

/** Initials for a partner's role tile ("Barista" -> "B", "Store clerk" -> "SC"). */
export function personaInitials(roleTitle: string): string {
	const words = roleTitle.split(/\s+/).filter(Boolean);
	const letters = words.map((word) =>
		word.replace(/^[^A-Za-z]*/, "").charAt(0),
	);
	return (letters.slice(0, 2).join("") || "?").toUpperCase();
}

/** Deterministic tile fill for a partner name (same rule the scenario card uses). */
export function tileTone(seed: string): (typeof TILE_TONES)[number] {
	return TILE_TONES[seed.length % TILE_TONES.length];
}
