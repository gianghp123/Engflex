import { Bookmark, ChevronDown, ChevronUp } from "lucide-react";
import { m } from "#/paraglide/messages";

export interface IeltsPaletteGroup {
	sectionNumber: number;
	title: string;
	questions: { id: string; questionNumber: number }[];
}

export function IeltsQuestionPalette({
	isCollapsed,
	onToggleCollapse,
	groups,
	activeQuestionId,
	answers,
	bookmarks,
	onSelectQuestion,
}: {
	isCollapsed: boolean;
	onToggleCollapse: () => void;
	groups: IeltsPaletteGroup[];
	activeQuestionId: string | null;
	answers: Record<string, string>;
	bookmarks: Set<string>;
	onSelectQuestion: (questionId: string, sectionNumber: number) => void;
}) {
	const totalQuestions = groups.reduce((acc, g) => acc + g.questions.length, 0);
	const answeredCount = Object.values(answers).filter(
		(v) => v != null && v.trim() !== "",
	).length;
	const pct =
		totalQuestions > 0 ? Math.round((answeredCount / totalQuestions) * 100) : 0;

	return (
		<div className="bg-card">
			<button
				type="button"
				onClick={onToggleCollapse}
				className="flex w-full items-center gap-3 px-4 py-2 text-xs font-semibold hover:bg-muted"
				aria-expanded={!isCollapsed}
				aria-label={
					isCollapsed
						? m["exams.ielts.take.expandPalette"]()
						: m["exams.ielts.take.collapsePalette"]()
				}
			>
				<span>{m["exams.ielts.take.progress"]()}</span>
				<span className="font-mono text-muted-foreground">
					{answeredCount}/{totalQuestions} ({pct}%)
				</span>
				<span className="h-1.5 max-w-40 flex-1 overflow-hidden rounded-full bg-muted">
					<span
						className="block h-full bg-emerald-600 transition-all"
						style={{ width: `${pct}%` }}
					/>
				</span>
				{isCollapsed ? (
					<ChevronUp className="ml-auto size-4" />
				) : (
					<ChevronDown className="ml-auto size-4" />
				)}
			</button>
			{!isCollapsed && (
				<div className="flex gap-5 overflow-x-auto px-4 pb-3">
					{groups.map((group) => (
						<div
							key={group.sectionNumber}
							className="min-w-44 flex-1 space-y-2"
						>
							<p className="text-xs font-semibold text-muted-foreground">
								{group.title}
							</p>
							<div className="grid grid-cols-5 gap-1.5">
								{group.questions.map(({ id, questionNumber }) => {
									const answered = Boolean(answers[id]?.trim());
									const active = activeQuestionId === id;
									return (
										<button
											key={id}
											type="button"
											aria-current={active ? "step" : undefined}
											onClick={() => onSelectQuestion(id, group.sectionNumber)}
											className={`relative flex h-8 items-center justify-center rounded-md border text-xs font-semibold transition-colors ${answered ? "border-sky-600 bg-sky-600 text-white" : "border-border bg-background hover:bg-muted"} ${active ? "ring-2 ring-foreground ring-offset-1 ring-offset-background" : ""}`}
										>
											{questionNumber}
											{bookmarks.has(id) && (
												<Bookmark className="absolute -right-1 -top-1 size-3 fill-amber-400 text-amber-700" />
											)}
										</button>
									);
								})}
							</div>
						</div>
					))}
				</div>
			)}
		</div>
	);
}
