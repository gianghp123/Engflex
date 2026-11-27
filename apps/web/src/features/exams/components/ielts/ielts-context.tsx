import type { IeltsContextBlock } from "@engflex/contracts";
import { createElement } from "react";
import { ieltsMediaSource } from "#/features/exams/ielts/media";
import { m } from "#/paraglide/messages";

const TAGS = {
	div: "div",
	p: "p",
	span: "span",
	b: "b",
	strong: "strong",
	i: "i",
	em: "em",
	u: "u",
	br: "br",
	ul: "ul",
	ol: "ol",
	li: "li",
	table: "table",
	thead: "thead",
	tbody: "tbody",
	tfoot: "tfoot",
	tr: "tr",
	td: "td",
	th: "th",
	h1: "h1",
	h2: "h2",
	h3: "h3",
	h4: "h4",
	h5: "h5",
	h6: "h6",
	sup: "sup",
	sub: "sub",
} as const;

function ContextNode({ block }: { block: IeltsContextBlock }) {
	if (block.tag === "text") return <>{block.text} </>;
	if (block.tag === "img") {
		const src = ieltsMediaSource(block.assetId);
		return src ? (
			<img
				src={src}
				alt={block.alt || m["exams.ielts.media.diagram"]()}
				className="my-3 h-auto max-w-full rounded-lg"
			/>
		) : (
			<p role="alert">{m["exams.ielts.media.failed"]()}</p>
		);
	}
	const tag = TAGS[block.tag as keyof typeof TAGS];
	if (!tag) return null;
	const children = block.children?.map((child, index) => (
		// biome-ignore lint/suspicious/noArrayIndexKey: Source nodes have no IDs; their validated order is immutable.
		<ContextNode key={`${child.tag}-${index}`} block={child} />
	));
	const element = createElement(
		tag,
		{
			...(tag === "td" || tag === "th"
				? {
						colSpan: block.colSpan || undefined,
						rowSpan: block.rowSpan || undefined,
					}
				: {}),
		},
		tag === "br" ? undefined : children,
	);
	return tag === "table" ? (
		<div className="my-3 overflow-x-auto">{element}</div>
	) : (
		element
	);
}

export function IeltsContext({
	blocks,
}: {
	blocks: IeltsContextBlock[] | null | undefined;
}) {
	return (
		<div className="text-sm leading-relaxed [&_p]:my-2 [&_table]:w-full [&_table]:border-collapse [&_td]:border [&_td]:border-border [&_td]:p-2 [&_th]:border [&_th]:border-border [&_th]:p-2 [&_ul]:list-inside [&_ul]:list-disc [&_ol]:list-inside [&_ol]:list-decimal">
			{blocks?.map((block, index) => (
				// biome-ignore lint/suspicious/noArrayIndexKey: Source nodes have no IDs; their validated order is immutable.
				<ContextNode key={`${block.tag}-${index}`} block={block} />
			))}
		</div>
	);
}
