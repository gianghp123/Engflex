import type { ExamMode, ExamTestDetail } from "@engflex/contracts";
import { createFileRoute, notFound, useNavigate } from "@tanstack/react-router";
import { Clock, FileText, Layers } from "lucide-react";
import { useRef, useState } from "react";
import { z } from "zod";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import {
	examTestQueryOptions,
	useStartExamAttempt,
} from "#/features/exams/queries";
import { loadOr404 } from "#/lib/router/route-loader";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/exams/ielts/$categoryId/$testId/")({
	loader: async ({
		context: { queryClient },
		params: { categoryId, testId },
	}) => {
		if (
			!z.uuid().safeParse(categoryId).success ||
			!z.uuid().safeParse(testId).success
		)
			throw notFound();
		const test = await loadOr404(queryClient, examTestQueryOptions(testId));
		if (test.categoryId !== categoryId || test.category?.type !== "ielts")
			throw notFound();
		return test;
	},
	pendingComponent: TestPending,
	component: IeltsTestDetailPage,
});

function TestPending() {
	return (
		<PageLayout>
			<Skeleton className="h-9 w-3/4" />
			<Skeleton className="h-4 w-1/2" />
			<div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
				<CardShell
					pending
					className="p-5 lg:col-span-7"
					skeleton={<Skeleton className="h-96 w-full rounded-xl" />}
				/>
				<CardShell
					pending
					className="p-5 lg:col-span-5"
					skeleton={<Skeleton className="h-96 w-full rounded-xl" />}
				/>
			</div>
		</PageLayout>
	);
}

function IeltsTestDetailPage() {
	const test = Route.useLoaderData();
	return <IeltsTestConfiguration key={test.id} test={test} />;
}

function IeltsTestConfiguration({ test }: { test: ExamTestDetail }) {
	const navigate = useNavigate();
	const start = useStartExamAttempt();
	const startLock = useRef(false);

	const sections = test.ieltsContent?.sections ?? [
		{ sectionNumber: 1, questionCount: 13 },
		{ sectionNumber: 2, questionCount: 13 },
		{ sectionNumber: 3, questionCount: 14 },
	];
	const allSections = sections.map((s) => s.sectionNumber);

	const [selectedMode, setSelectedMode] = useState<ExamMode>("exam");
	const [practiceSections, setPracticeSections] =
		useState<number[]>(allSections);

	const selectedParts =
		selectedMode === "exam" ? allSections : practiceSections;

	const durationMinutes =
		test.skill === "listening" ? 40 : test.skill === "writing" ? 60 : 60;
	const totalQuestions = sections.reduce((sum, s) => sum + s.questionCount, 0);

	useBreadcrumbs([
		{ label: m["exams.ielts.hub.title"](), to: APP_ROUTES.EXAMS.IELTS.LIST },
		{
			label: test.category?.title ?? "",
			to: APP_ROUTES.EXAMS.IELTS.CATEGORY,
			params: { categoryId: test.categoryId },
		},
		{ label: test.title },
	]);

	const toggleSection = (sectionNumber: number) => {
		if (practiceSections.includes(sectionNumber)) {
			if (practiceSections.length === 1) return;
			setPracticeSections(practiceSections.filter((s) => s !== sectionNumber));
		} else {
			setPracticeSections(
				[...practiceSections, sectionNumber].sort((a, b) => a - b),
			);
		}
	};

	const handleStart = async () => {
		if (startLock.current || start.isPending || selectedParts.length === 0)
			return;
		startLock.current = true;
		try {
			const attempt = await start.mutateAsync({
				testId: test.id,
				req: {
					mode: selectedMode,
					selectedParts,
				},
			});
			void navigate({
				to: APP_ROUTES.EXAMS.IELTS.SESSION,
				params: { id: attempt.id },
			});
		} catch {
			startLock.current = false;
		}
	};

	return (
		<PageLayout>
			<div className="flex flex-col gap-6">
				<div>
					<BackButton
						to={APP_ROUTES.EXAMS.IELTS.CATEGORY}
						params={{ categoryId: test.categoryId }}
						label={m["exams.ielts.test.back"]()}
					/>
					<div className="mt-3 flex flex-col gap-2">
						<div className="flex flex-wrap items-center gap-2">
							<h1 className="text-2xl font-bold tracking-tight text-foreground md:text-3xl">
								{test.title}
							</h1>
							{test.skill && (
								<span className="rounded-md bg-sky-500/10 px-2 py-0.5 text-xs font-semibold text-sky-600 dark:text-sky-400 capitalize">
									{m["exams.ielts.test.skill"]({ skill: test.skill })}
								</span>
							)}
							{test.subType && (
								<span className="rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary uppercase">
									{m["exams.ielts.test.subType"]({ subType: test.subType })}
								</span>
							)}
						</div>
						<div className="flex flex-wrap items-center gap-4 text-xs font-medium text-muted-foreground md:text-sm">
							<span className="inline-flex items-center gap-1.5">
								<Clock className="size-4" />
								{m["exams.ielts.test.duration"]({ count: durationMinutes })}
							</span>
							<span className="inline-flex items-center gap-1.5">
								<FileText className="size-4" />
								{m["exams.ielts.test.questionsCount"]({
									count: totalQuestions,
								})}
							</span>
							<span className="inline-flex items-center gap-1.5">
								<Layers className="size-4" />
								{m["exams.ielts.test.sectionsCount"]({
									count: sections.length,
								})}
							</span>
						</div>
					</div>
				</div>

				<div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
					<div className="flex flex-col gap-4 lg:col-span-7">
						<button
							type="button"
							onClick={() => setSelectedMode("exam")}
							className={`flex w-full cursor-pointer flex-col gap-2 rounded-2xl border p-5 text-left transition-all ${
								selectedMode === "exam"
									? "border-primary bg-primary/5 shadow-sm"
									: "border-border/80 bg-card hover:border-border"
							}`}
						>
							<div className="flex items-center justify-between">
								<h3 className="text-base font-bold text-foreground">
									{m["exams.ielts.test.modeExamTitle"]()}
								</h3>
								<div
									className={`size-4 rounded-full border-2 transition-all ${
										selectedMode === "exam"
											? "border-primary bg-primary"
											: "border-muted-foreground"
									}`}
								/>
							</div>
							<p className="text-xs text-muted-foreground md:text-sm">
								{m["exams.ielts.test.modeExamDesc"]()}
							</p>
						</button>

						<button
							type="button"
							onClick={() => setSelectedMode("practice")}
							className={`flex w-full cursor-pointer flex-col gap-2 rounded-2xl border p-5 text-left transition-all ${
								selectedMode === "practice"
									? "border-primary bg-primary/5 shadow-sm"
									: "border-border/80 bg-card hover:border-border"
							}`}
						>
							<div className="flex items-center justify-between">
								<h3 className="text-base font-bold text-foreground">
									{m["exams.ielts.test.modePracticeTitle"]()}
								</h3>
								<div
									className={`size-4 rounded-full border-2 transition-all ${
										selectedMode === "practice"
											? "border-primary bg-primary"
											: "border-muted-foreground"
									}`}
								/>
							</div>
							<p className="text-xs text-muted-foreground md:text-sm">
								{m["exams.ielts.test.modePracticeDesc"]()}
							</p>

							{selectedMode === "practice" && (
								<div className="mt-3 flex flex-col gap-2 border-t border-border/60 pt-3">
									<span className="text-xs font-semibold text-foreground">
										{m["exams.ielts.test.selectSections"]()}
									</span>
									<div className="flex flex-wrap gap-2">
										{sections.map((s) => {
											const selected = practiceSections.includes(
												s.sectionNumber,
											);
											return (
												<button
													key={s.sectionNumber}
													type="button"
													onClick={(e) => {
														e.stopPropagation();
														toggleSection(s.sectionNumber);
													}}
													className={`rounded-xl border px-3 py-1.5 text-xs font-bold transition-all ${
														selected
															? "border-primary bg-primary text-primary-foreground"
															: "border-border/80 bg-background text-muted-foreground hover:bg-muted"
													}`}
												>
													{m["exams.ielts.test.sectionItem"]({
														number: s.sectionNumber,
													})}
													<span className="ml-1 opacity-70">
														({s.questionCount}q)
													</span>
												</button>
											);
										})}
									</div>
								</div>
							)}
						</button>
					</div>

					<div className="flex flex-col gap-4 lg:col-span-5">
						<div className="flex flex-col gap-4 rounded-2xl border border-border/80 bg-card p-6 shadow-sm">
							<h3 className="text-sm font-bold uppercase tracking-wider text-muted-foreground">
								Thông tin bài thi
							</h3>
							<div className="flex flex-col divide-y divide-border/60 text-sm">
								<div className="flex justify-between py-2">
									<span className="text-muted-foreground">Chế độ</span>
									<span className="font-semibold text-foreground">
										{selectedMode === "exam"
											? m["exams.ielts.test.modeExamTitle"]()
											: m["exams.ielts.test.modePracticeTitle"]()}
									</span>
								</div>
								<div className="flex justify-between py-2">
									<span className="text-muted-foreground">Section đã chọn</span>
									<span className="font-semibold text-foreground">
										{selectedParts.join(", ")}
									</span>
								</div>
								<div className="flex justify-between py-2">
									<span className="text-muted-foreground">Thời gian</span>
									<span className="font-semibold text-foreground">
										{selectedMode === "exam"
											? m["exams.ielts.test.duration"]({
													count: durationMinutes,
												})
											: "Không giới hạn"}
									</span>
								</div>
							</div>

							<Button
								size="lg"
								onClick={handleStart}
								disabled={start.isPending || selectedParts.length === 0}
								className="mt-2 w-full rounded-xl font-bold shadow-md"
							>
								{selectedMode === "exam"
									? m["exams.ielts.test.startExam"]()
									: m["exams.ielts.test.startPractice"]()}
							</Button>
						</div>
					</div>
				</div>
			</div>
		</PageLayout>
	);
}
