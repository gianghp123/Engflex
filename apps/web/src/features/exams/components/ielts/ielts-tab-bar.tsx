import { BookOpen } from "lucide-react";

export interface IeltsSectionTabMeta {
	sectionNumber: number;
	label: string;
	totalQuestions: number;
	answeredCount: number;
}

export function IeltsTabBar({
	sections,
	activeSection,
	onSelectSection,
}: {
	sections: IeltsSectionTabMeta[];
	activeSection: number;
	onSelectSection: (sectionNumber: number) => void;
}) {
	return (
		<div className="flex h-11 w-full shrink-0 items-center border-b border-border/80 bg-muted/20 px-4 overflow-x-auto select-none scrollbar-none">
			<div className="flex items-center gap-1.5">
				<span className="flex size-7 shrink-0 items-center justify-center text-muted-foreground mr-1">
					<BookOpen className="size-4" aria-hidden="true" />
				</span>

				{sections.map((sec) => {
					const isActive = activeSection === sec.sectionNumber;
					return (
						<button
							key={sec.sectionNumber}
							type="button"
							onClick={() => onSelectSection(sec.sectionNumber)}
							className={`inline-flex items-center gap-2 rounded-md px-3.5 py-1.5 text-xs font-bold transition-all cursor-pointer ${
								isActive
									? "bg-[#1e293b] text-white shadow-xs dark:bg-slate-800"
									: "text-muted-foreground hover:bg-muted/70 hover:text-foreground"
							}`}
						>
							<span>{sec.label}</span>
							<span
								className={`rounded px-1.5 py-0.5 text-[10px] font-bold ${
									isActive
										? "bg-white/20 text-white"
										: sec.answeredCount > 0
											? "bg-sky-500/20 text-sky-600 dark:text-sky-400"
											: "bg-muted text-muted-foreground"
								}`}
							>
								{sec.answeredCount}/{sec.totalQuestions}
							</span>
						</button>
					);
				})}
			</div>
		</div>
	);
}
