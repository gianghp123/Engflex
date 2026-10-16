export function PassagePanel({
	kicker,
	title,
	passage,
}: {
	kicker: string;
	title: string;
	passage: string;
}) {
	const paragraphs = passage
		.split(/\n{2,}/)
		.map((paragraph) => paragraph.trim())
		.filter(Boolean);

	return (
		<article className="surface-card flex h-full flex-col gap-3 p-5">
			<div className="flex flex-wrap items-center justify-between gap-2">
				<span className="chip bg-accent px-2 py-0.5 text-[11px] font-bold">
					{kicker}
				</span>
			</div>
			<h2 className="text-xl font-bold text-foreground">{title}</h2>
			<div className="flex flex-col gap-3 text-[15px] leading-[24px] font-medium text-foreground">
				{paragraphs.map((paragraph) => (
					<p key={paragraph.slice(0, 40)}>{paragraph}</p>
				))}
			</div>
		</article>
	);
}
