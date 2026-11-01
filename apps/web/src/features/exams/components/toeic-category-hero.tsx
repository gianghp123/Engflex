import { BookOpen } from "lucide-react";
import { m } from "#/paraglide/messages";

export interface ToeicCategoryHeroProps {
	title: string;
	tags: string[];
	testsCount: number;
}

export function ToeicCategoryHero({
	title,
	tags,
	testsCount,
}: ToeicCategoryHeroProps) {
	return (
		<div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-blue-600 via-blue-500 to-sky-500 p-6 text-white shadow-md md:p-8">
			{/* Background watermark */}
			<BookOpen
				className="pointer-events-none absolute -bottom-6 right-2 size-36 text-white/10 md:-bottom-8 md:right-8 md:size-48"
				aria-hidden="true"
			/>

			<div className="relative z-10 flex flex-col gap-3 md:gap-4">
				<div className="flex flex-wrap items-center gap-2">
					{tags.map((tag) => (
						<span
							key={tag}
							className="inline-flex items-center rounded-md bg-white/20 px-2.5 py-0.5 text-xs font-semibold text-white backdrop-blur-sm"
						>
							{tag}
						</span>
					))}
				</div>

				<h1 className="text-2xl font-extrabold tracking-tight md:text-3xl lg:text-4xl">
					{title}
				</h1>

				<div className="flex flex-wrap items-center gap-4 text-xs font-medium text-white/90 md:text-sm">
					<span className="inline-flex items-center gap-1.5 rounded-full bg-white/15 px-3 py-1 backdrop-blur-sm">
						<BookOpen className="size-4" aria-hidden="true" />
						{m["exams.toeic.hub.testsCount"]({ count: testsCount })}
					</span>
				</div>
			</div>
		</div>
	);
}
