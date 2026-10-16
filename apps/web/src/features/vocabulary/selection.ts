/**
 * Selection utilities for the context dock. Pure DOM reads: everything is
 * captured at menu-open time, because opening a dialog later moves focus and
 * the browser selection with it.
 */

/** Longest context quote the example field accepts (characters). */
const MAX_CONTEXT = 280;

/** Trimmed selected text, only when the selection sits inside `region`. */
export function readSelection(region: HTMLElement): string {
	const sel = window.getSelection();
	if (!sel || sel.isCollapsed || sel.rangeCount === 0) return "";
	const anchor = sel.anchorNode;
	if (!anchor || !region.contains(anchor)) return "";
	return sel.toString().trim();
}

/**
 * The sentence (or transcript line) the selection sits in. Climbs to the
 * nearest element whose content is a single text run — a passage `<p>`, a
 * transcript line `<span>` — then widens from the selection to the nearest
 * sentence boundary. Returns null when nothing sensible can be derived.
 */
export function selectionContextSentence(region: HTMLElement): string | null {
	const sel = window.getSelection();
	if (!sel || sel.isCollapsed || sel.rangeCount === 0) return null;
	const anchor = sel.anchorNode;
	if (!anchor || !region.contains(anchor)) return null;
	const selected = sel.toString().trim();
	if (!selected) return null;
	const text = (textBlockOf(anchor, region)?.textContent ?? "").trim();
	if (!text) return null;
	const idx = text.indexOf(selected);
	if (idx < 0) return text.slice(0, MAX_CONTEXT);
	const before = text.slice(0, idx);
	const starts = [...before.matchAll(/[.!?…]["')\]]?\s+/g)];
	const lastStart = starts.at(-1);
	const start = lastStart ? lastStart.index + lastStart[0].length : 0;
	const after = text.slice(idx + selected.length);
	const endMatch = after.match(/[.!?…]["')\]]?(\s+|$)/);
	const end =
		endMatch && endMatch.index !== undefined
			? idx + selected.length + endMatch.index + endMatch[0].length
			: text.length;
	return text.slice(start, end).trim().slice(0, MAX_CONTEXT);
}

/** Nearest ancestor whose only child is one text node, else nearest block. */
function textBlockOf(node: Node, region: HTMLElement): HTMLElement | null {
	let el: HTMLElement | null =
		node.nodeType === Node.TEXT_NODE
			? node.parentElement
			: node instanceof HTMLElement
				? node
				: node.parentElement;
	while (el && el !== region) {
		if (
			el.childNodes.length === 1 &&
			el.firstChild?.nodeType === Node.TEXT_NODE
		)
			return el;
		el = el.parentElement;
	}
	return node.parentElement?.closest("p, li, blockquote") ?? null;
}
