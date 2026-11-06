import { ArrowDown, Bookmark, HelpCircle } from "lucide-react";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import type {
	IeltsQuestion,
	IeltsQuestionGroup,
} from "#/features/exams/ielts/mock-data";
import { m } from "#/paraglide/messages";

export function IeltsQuestionPane({
	groups,
	answers,
	bookmarks,
	activeQuestionId,
	onAnswerChange,
	onToggleBookmark,
}: {
	groups: IeltsQuestionGroup[];
	answers: Record<string, string>;
	bookmarks: Set<string>;
	activeQuestionId: string | null;
	onAnswerChange: (questionId: string, answer: string) => void;
	onToggleBookmark: (questionId: string) => void;
}) {
	return (
		<div className="flex h-full flex-col overflow-hidden bg-background">
			<div className="flex h-10 shrink-0 items-center justify-between border-b border-border/70 px-4 bg-muted/30">
				<div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-muted-foreground">
					<HelpCircle className="size-3.5 text-sky-500" />
					<span>{m["exams.ielts.take.questionsTitle"]()}</span>
				</div>
			</div>

			<div className="flex-1 overflow-y-auto p-4 md:p-6 space-y-8">
				{groups.map((group) => (
					<div
						key={group.id}
						className="space-y-5 rounded-xl border border-border/70 bg-card p-4 md:p-6 shadow-xs"
					>
						<div className="rounded-lg border border-sky-200 bg-sky-50/60 p-3.5 text-xs md:text-sm text-sky-950 dark:border-sky-800/80 dark:bg-sky-950/30 dark:text-sky-200">
							<p className="font-semibold whitespace-pre-line leading-relaxed">
								{group.instruction}
							</p>
						</div>

						{group.isFlowchart ? (
							<IeltsFlowchartGroup
								questions={group.questions}
								answers={answers}
								bookmarks={bookmarks}
								activeQuestionId={activeQuestionId}
								onAnswerChange={onAnswerChange}
								onToggleBookmark={onToggleBookmark}
							/>
						) : (
							<div className="space-y-4">
								{group.questions.map((question) => (
									<IeltsQuestionCard
										key={question.id}
										question={question}
										answer={answers[question.id] ?? ""}
										isBookmarked={bookmarks.has(question.id)}
										isActive={activeQuestionId === question.id}
										onAnswerChange={(val) => onAnswerChange(question.id, val)}
										onToggleBookmark={() => onToggleBookmark(question.id)}
									/>
								))}
							</div>
						)}
					</div>
				))}
			</div>
		</div>
	);
}

function IeltsQuestionCard({
	question,
	answer,
	isBookmarked,
	isActive,
	onAnswerChange,
	onToggleBookmark,
}: {
	question: IeltsQuestion;
	answer: string;
	isBookmarked: boolean;
	isActive: boolean;
	onAnswerChange: (val: string) => void;
	onToggleBookmark: () => void;
}) {
	return (
		<div
			id={question.id}
			className={`scroll-mt-6 rounded-lg border p-4 transition-all duration-200 ${
				isActive
					? "border-sky-500 bg-sky-50/20 shadow-xs dark:bg-sky-950/20"
					: "border-border/60 bg-background hover:border-border"
			}`}
		>
			<div className="flex items-start justify-between gap-3">
				<div className="flex items-center gap-2">
					<span className="flex size-6 shrink-0 items-center justify-center rounded-md bg-muted font-mono text-xs font-bold text-foreground">
						{question.number}
					</span>
					<span className="text-xs font-bold text-muted-foreground uppercase">
						{m["exams.ielts.take.question"]({ number: question.number })}
					</span>
				</div>

				<Button
					variant="ghost"
					size="sm"
					onClick={onToggleBookmark}
					className={`h-7 px-2 text-xs gap-1 ${
						isBookmarked
							? "text-amber-500 hover:text-amber-600 bg-amber-50 dark:bg-amber-950/40"
							: "text-muted-foreground hover:text-foreground"
					}`}
					title={
						isBookmarked
							? m["exams.ielts.take.bookmarked"]()
							: m["exams.ielts.take.bookmark"]()
					}
				>
					<Bookmark
						className={`size-3.5 ${isBookmarked ? "fill-amber-400" : ""}`}
					/>
					<span className="hidden sm:inline">
						{isBookmarked
							? m["exams.ielts.take.bookmarked"]()
							: m["exams.ielts.take.bookmark"]()}
					</span>
				</Button>
			</div>

			{question.prompt && (
				<p className="mt-2 text-sm md:text-[15px] font-medium leading-relaxed text-foreground">
					{question.prompt}
				</p>
			)}

			<div className="mt-3">
				{question.type === "true_false_not_given" && question.options && (
					<div className="flex flex-wrap gap-2">
						{question.options.map((opt) => {
							const isSelected = answer === opt;
							return (
								<button
									key={opt}
									type="button"
									onClick={() => onAnswerChange(opt)}
									className={`cursor-pointer rounded-lg border px-3.5 py-1.5 text-xs font-semibold transition-all ${
										isSelected
											? "border-sky-500 bg-sky-500 text-white shadow-xs"
											: "border-border/70 bg-card text-foreground hover:bg-muted"
									}`}
								>
									{opt}
								</button>
							);
						})}
					</div>
				)}

				{question.type === "matching" && question.options && (
					<div className="flex items-center gap-2 flex-wrap">
						<span className="text-xs font-medium text-muted-foreground mr-1">
							Chọn bảo tàng:
						</span>
						{question.options.map((letter) => {
							const isSelected = answer === letter;
							return (
								<button
									key={letter}
									type="button"
									onClick={() => onAnswerChange(letter)}
									className={`cursor-pointer size-8 rounded-lg border text-xs font-bold transition-all ${
										isSelected
											? "border-sky-500 bg-sky-500 text-white shadow-xs"
											: "border-border/70 bg-card text-foreground hover:bg-muted"
									}`}
								>
									{letter}
								</button>
							);
						})}
					</div>
				)}

				{question.type === "fill_blank" && (
					<div className="max-w-md">
						<Input
							type="text"
							value={answer}
							onChange={(e) => onAnswerChange(e.target.value)}
							placeholder={m["exams.ielts.take.inputPlaceholder"]()}
							className="h-9 text-sm"
						/>
					</div>
				)}
			</div>
		</div>
	);
}

function IeltsFlowchartGroup({
	questions,
	answers,
	bookmarks,
	activeQuestionId,
	onAnswerChange,
	onToggleBookmark,
}: {
	questions: IeltsQuestion[];
	answers: Record<string, string>;
	bookmarks: Set<string>;
	activeQuestionId: string | null;
	onAnswerChange: (questionId: string, val: string) => void;
	onToggleBookmark: (questionId: string) => void;
}) {
	const q22 = questions.find((q) => q.number === 22);
	const q23 = questions.find((q) => q.number === 23);
	const q24 = questions.find((q) => q.number === 24);
	const q25 = questions.find((q) => q.number === 25);
	const q26 = questions.find((q) => q.number === 26);
	const q27 = questions.find((q) => q.number === 27);

	return (
		<div className="space-y-4 select-none">
			<div className="text-center font-bold text-sm tracking-wide text-muted-foreground uppercase mb-2">
				Sending Office Meeting Invites Flowchart
			</div>

			<div className="rounded-xl border border-sky-200/80 bg-sky-50/30 p-4 space-y-3 dark:border-sky-900/60 dark:bg-sky-950/20">
				<h4 className="text-xs font-bold uppercase tracking-wider text-sky-700 dark:text-sky-300">
					1. Checking Before Sending
				</h4>
				<div className="space-y-2">
					{q22 && (
						<FlowchartItem
							question={q22}
							label="Skip meetings that could be resolved by a"
							postLabel="or a simple question."
							answer={answers[q22.id] ?? ""}
							isBookmarked={bookmarks.has(q22.id)}
							isActive={activeQuestionId === q22.id}
							onAnswerChange={(val) => onAnswerChange(q22.id, val)}
							onToggleBookmark={() => onToggleBookmark(q22.id)}
						/>
					)}
					{q23 && (
						<FlowchartItem
							question={q23}
							label="Check content: Ensure there is a clear"
							postLabel="with items in sensible order."
							answer={answers[q23.id] ?? ""}
							isBookmarked={bookmarks.has(q23.id)}
							isActive={activeQuestionId === q23.id}
							onAnswerChange={(val) => onAnswerChange(q23.id, val)}
							onToggleBookmark={() => onToggleBookmark(q23.id)}
						/>
					)}
					{q24 && (
						<FlowchartItem
							question={q24}
							label="Check meeting time: Use scheduling tool to spot any"
							postLabel="with key attendees."
							answer={answers[q24.id] ?? ""}
							isBookmarked={bookmarks.has(q24.id)}
							isActive={activeQuestionId === q24.id}
							onAnswerChange={(val) => onAnswerChange(q24.id, val)}
							onToggleBookmark={() => onToggleBookmark(q24.id)}
						/>
					)}
				</div>
			</div>

			<div className="flex justify-center text-muted-foreground">
				<ArrowDown className="size-5 animate-bounce" />
			</div>

			<div className="rounded-xl border border-indigo-200/80 bg-indigo-50/30 p-4 space-y-3 dark:border-indigo-900/60 dark:bg-indigo-950/20">
				<h4 className="text-xs font-bold uppercase tracking-wider text-indigo-700 dark:text-indigo-300">
					2. Sending in Two Steps (Larger / Formal Meetings)
				</h4>
				<div className="space-y-2">
					{q25 && (
						<FlowchartItem
							question={q25}
							label="Step A: Send a meeting"
							postLabel="early to secure date and time for stakeholders."
							answer={answers[q25.id] ?? ""}
							isBookmarked={bookmarks.has(q25.id)}
							isActive={activeQuestionId === q25.id}
							onAnswerChange={(val) => onAnswerChange(q25.id, val)}
							onToggleBookmark={() => onToggleBookmark(q25.id)}
						/>
					)}
					{q26 && (
						<FlowchartItem
							question={q26}
							label="Step B: Agree a firm"
							postLabel="at the outset for the full details."
							answer={answers[q26.id] ?? ""}
							isBookmarked={bookmarks.has(q26.id)}
							isActive={activeQuestionId === q26.id}
							onAnswerChange={(val) => onAnswerChange(q26.id, val)}
							onToggleBookmark={() => onToggleBookmark(q26.id)}
						/>
					)}
					{q27 && (
						<FlowchartItem
							question={q27}
							label="Final step: Update calendar invite and ensure"
							postLabel="notification is triggered."
							answer={answers[q27.id] ?? ""}
							isBookmarked={bookmarks.has(q27.id)}
							isActive={activeQuestionId === q27.id}
							onAnswerChange={(val) => onAnswerChange(q27.id, val)}
							onToggleBookmark={() => onToggleBookmark(q27.id)}
						/>
					)}
				</div>
			</div>
		</div>
	);
}

function FlowchartItem({
	question,
	label,
	postLabel,
	answer,
	isBookmarked,
	isActive,
	onAnswerChange,
	onToggleBookmark,
}: {
	question: IeltsQuestion;
	label: string;
	postLabel?: string;
	answer: string;
	isBookmarked: boolean;
	isActive: boolean;
	onAnswerChange: (val: string) => void;
	onToggleBookmark: () => void;
}) {
	return (
		<div
			id={question.id}
			className={`scroll-mt-6 rounded-lg border p-3 bg-background transition-all ${
				isActive
					? "border-sky-500 shadow-xs ring-1 ring-sky-500"
					: "border-border/60 hover:border-border"
			}`}
		>
			<div className="flex items-center justify-between gap-2 mb-2">
				<div className="flex items-center gap-1.5">
					<span className="flex size-5 shrink-0 items-center justify-center rounded bg-muted font-mono text-[11px] font-bold text-foreground">
						{question.number}
					</span>
					<span className="text-[11px] font-bold text-muted-foreground uppercase">
						{m["exams.ielts.take.question"]({ number: question.number })}
					</span>
				</div>

				<button
					type="button"
					onClick={onToggleBookmark}
					className={`p-1 rounded hover:bg-muted text-xs cursor-pointer ${
						isBookmarked ? "text-amber-500" : "text-muted-foreground"
					}`}
				>
					<Bookmark
						className={`size-3.5 ${isBookmarked ? "fill-amber-400" : ""}`}
					/>
				</button>
			</div>

			<p className="text-xs text-foreground leading-relaxed mb-2">
				<span>{label} </span>
				<span className="font-bold underline">({question.number})</span>
				{postLabel && <span> {postLabel}</span>}
			</p>

			<Input
				type="text"
				value={answer}
				onChange={(e) => onAnswerChange(e.target.value)}
				placeholder={m["exams.ielts.take.inputPlaceholder"]()}
				className="h-8 text-xs max-w-sm"
			/>
		</div>
	);
}
