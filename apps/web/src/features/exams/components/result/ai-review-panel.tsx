import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { CardShell } from "#/components/common/card-shell";
import { SubmitButton } from "#/components/common/submit-button";
import { Skeleton } from "#/components/ui/skeleton";
import {
	aiReviewQueryOptions,
	studyPlanQueryOptions,
} from "#/features/exams/queries";
import { generateAiReview, generateStudyPlan } from "#/features/exams/service";
import { OutOfCreditsModal } from "#/features/payments/components/out-of-credits-modal";
import { ApiError } from "#/lib/api";
import { m } from "#/paraglide/messages";

/** Local wire copies of the LLM review/plan documents (prompts_exam.go:
 *  ExamReviewOutput / StudyPlanOutput). The payload crosses as stored JSON,
 *  so every field is guarded before render. */
interface ExamMistake {
	questionId: string;
	selectedKey?: string;
	correctKey: string;
	explanation: string;
	stem?: string;
}

interface ExamReviewPayload {
	summary: string;
	mistakes: ExamMistake[];
}

interface StudyPlanFocusArea {
	area: string;
	detail: string;
}

interface StudyPlanItem {
	title: string;
	detail: string;
}

interface StudyPlanPayload {
	summary: string;
	focusAreas: StudyPlanFocusArea[];
	plan: StudyPlanItem[];
}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === "object" && value !== null;
}

function parseReview(payload: unknown): ExamReviewPayload | null {
	if (!isRecord(payload)) return null;
	if (typeof payload.summary !== "string") return null;
	if (!Array.isArray(payload.mistakes)) return null;
	const mistakes: ExamMistake[] = [];
	for (const item of payload.mistakes) {
		if (!isRecord(item)) return null;
		if (
			typeof item.questionId !== "string" ||
			typeof item.correctKey !== "string" ||
			typeof item.explanation !== "string"
		)
			return null;
		mistakes.push({
			questionId: item.questionId,
			selectedKey:
				typeof item.selectedKey === "string" ? item.selectedKey : undefined,
			correctKey: item.correctKey,
			explanation: item.explanation,
			stem: typeof item.stem === "string" ? item.stem : undefined,
		});
	}
	return { summary: payload.summary, mistakes };
}

function parsePlan(payload: unknown): StudyPlanPayload | null {
	if (!isRecord(payload)) return null;
	if (typeof payload.summary !== "string") return null;
	if (!Array.isArray(payload.focusAreas) || !Array.isArray(payload.plan))
		return null;
	const focusAreas: StudyPlanFocusArea[] = [];
	for (const item of payload.focusAreas) {
		if (!isRecord(item)) return null;
		if (typeof item.area !== "string" || typeof item.detail !== "string")
			return null;
		focusAreas.push({ area: item.area, detail: item.detail });
	}
	const plan: StudyPlanItem[] = [];
	for (const item of payload.plan) {
		if (!isRecord(item)) return null;
		if (typeof item.title !== "string" || typeof item.detail !== "string")
			return null;
		plan.push({ title: item.title, detail: item.detail });
	}
	return { summary: payload.summary, focusAreas, plan };
}

/** A quota denial never burns a retry: re-opening a cached review is free. */
function cachedRetry(count: number, error: unknown): boolean {
	return (
		!(error instanceof ApiError && [403, 404, 429].includes(error.status)) &&
		count < 2
	);
}

function isQuotaError(error: unknown): error is ApiError {
	return (
		error instanceof ApiError &&
		(error.code === "insufficient_credits" || error.code === "quota_exceeded")
	);
}

function isMissing(error: unknown): boolean {
	return error instanceof ApiError && error.status === 404;
}

function ReviewBody({ payload }: { payload: ExamReviewPayload }) {
	return (
		<div className="flex flex-col gap-4">
			<div>
				<h3 className="text-sm font-semibold">
					{m["exams.result.aiReview.summary"]()}
				</h3>
				<p className="mt-1 text-sm text-muted-foreground">{payload.summary}</p>
			</div>
			<div>
				<h3 className="text-sm font-semibold">
					{m["exams.result.aiReview.mistakesTitle"]()}
				</h3>
				{payload.mistakes.length === 0 ? (
					<p className="mt-1 text-sm text-muted-foreground">
						{m["exams.result.aiReview.empty"]()}
					</p>
				) : (
					<ul className="mt-2 flex flex-col gap-3">
						{payload.mistakes.map((mistake) => (
							<li
								key={mistake.questionId}
								className="rounded-lg border border-border/60 p-3"
							>
								<p className="text-sm font-semibold">
									{mistake.stem || mistake.questionId}
								</p>
								<p className="mt-1 text-xs text-muted-foreground">
									{m["exams.result.aiReview.yourAnswer"]({
										answer: mistake.selectedKey ?? "—",
									})}{" "}
									·{" "}
									{m["exams.result.aiReview.correctAnswer"]({
										answer: mistake.correctKey,
									})}
								</p>
								<p className="mt-1 text-sm">{mistake.explanation}</p>
							</li>
						))}
					</ul>
				)}
			</div>
		</div>
	);
}

function PlanBody({ payload }: { payload: StudyPlanPayload }) {
	return (
		<div className="flex flex-col gap-4">
			<div>
				<h3 className="text-sm font-semibold">
					{m["exams.result.studyPlan.summary"]()}
				</h3>
				<p className="mt-1 text-sm text-muted-foreground">{payload.summary}</p>
			</div>
			<div>
				<h3 className="text-sm font-semibold">
					{m["exams.result.studyPlan.focusTitle"]()}
				</h3>
				{payload.focusAreas.length === 0 ? (
					<p className="mt-1 text-sm text-muted-foreground">
						{m["exams.result.studyPlan.empty"]()}
					</p>
				) : (
					<ul className="mt-2 flex flex-col gap-2">
						{payload.focusAreas.map((focus) => (
							<li
								key={focus.area}
								className="rounded-lg border border-border/60 p-3"
							>
								<p className="text-sm font-semibold">{focus.area}</p>
								<p className="mt-1 text-sm text-muted-foreground">
									{focus.detail}
								</p>
							</li>
						))}
					</ul>
				)}
			</div>
			<div>
				<h3 className="text-sm font-semibold">
					{m["exams.result.studyPlan.stepsTitle"]()}
				</h3>
				<ol className="mt-2 flex list-decimal flex-col gap-2 pl-5">
					{payload.plan.map((step) => (
						<li key={step.title} className="text-sm">
							<span className="font-semibold">{step.title}</span>
							<span className="text-muted-foreground"> — {step.detail}</span>
						</li>
					))}
				</ol>
			</div>
		</div>
	);
}

function ReviewSection({
	attemptId,
	examReviewCost,
}: {
	attemptId: string;
	examReviewCost: number;
}) {
	const queryClient = useQueryClient();
	const [quotaError, setQuotaError] = useState<ApiError | null>(null);
	const cached = useQuery({
		...aiReviewQueryOptions(attemptId),
		retry: cachedRetry,
	});
	const generate = useMutation({
		mutationFn: () => generateAiReview(attemptId),
		onSuccess: (data) => {
			queryClient.setQueryData(aiReviewQueryOptions(attemptId).queryKey, data);
		},
		onError: (error) => {
			if (isQuotaError(error)) setQuotaError(error);
		},
	});

	const parsed = cached.data ? parseReview(cached.data.payload) : null;

	return (
		<CardShell
			pending={cached.isPending}
			skeleton={<Skeleton className="h-48 w-full" />}
			className="p-4 sm:p-6"
		>
			<div className="flex flex-col gap-3">
				<h2 className="text-base font-bold">
					{m["exams.result.aiReview.title"]()}
				</h2>
				{cached.data && parsed ? (
					<ReviewBody payload={parsed} />
				) : (cached.data && !parsed) ||
					(cached.isError && !isMissing(cached.error)) ? (
					<>
						<p className="text-sm text-destructive">
							{m["exams.result.aiReview.failed"]()}
						</p>
						<div>
							<SubmitButton
								type="button"
								pending={cached.isFetching}
								onClick={() => void cached.refetch()}
							>
								{m["exams.result.aiReview.cta"]({ credits: examReviewCost })}
							</SubmitButton>
						</div>
					</>
				) : (
					<>
						<p className="text-sm text-muted-foreground">
							{m["exams.result.aiReview.hint"]({ credits: examReviewCost })}
						</p>
						{generate.isError && !isQuotaError(generate.error) ? (
							<p className="text-sm text-destructive">
								{m["exams.result.aiReview.failed"]()}
							</p>
						) : null}
						<div>
							<SubmitButton
								type="button"
								pending={generate.isPending}
								loadingLabel={m["exams.result.aiReview.generating"]()}
								onClick={() => generate.mutate()}
							>
								{m["exams.result.aiReview.cta"]({ credits: examReviewCost })}
							</SubmitButton>
						</div>
					</>
				)}
			</div>
			<OutOfCreditsModal
				error={quotaError}
				onClose={() => setQuotaError(null)}
			/>
		</CardShell>
	);
}

function PlanSection({
	attemptId,
	studyPlanCost,
}: {
	attemptId: string;
	studyPlanCost: number;
}) {
	const queryClient = useQueryClient();
	const [quotaError, setQuotaError] = useState<ApiError | null>(null);
	const cached = useQuery({
		...studyPlanQueryOptions(attemptId),
		retry: cachedRetry,
	});
	const generate = useMutation({
		mutationFn: () => generateStudyPlan(attemptId),
		onSuccess: (data) => {
			queryClient.setQueryData(studyPlanQueryOptions(attemptId).queryKey, data);
		},
		onError: (error) => {
			if (isQuotaError(error)) setQuotaError(error);
		},
	});

	const parsed = cached.data ? parsePlan(cached.data.payload) : null;

	return (
		<CardShell
			pending={cached.isPending}
			skeleton={<Skeleton className="h-48 w-full" />}
			className="p-4 sm:p-6"
		>
			<div className="flex flex-col gap-3">
				<h2 className="text-base font-bold">
					{m["exams.result.studyPlan.title"]()}
				</h2>
				{cached.data && parsed ? (
					<PlanBody payload={parsed} />
				) : (cached.data && !parsed) ||
					(cached.isError && !isMissing(cached.error)) ? (
					<>
						<p className="text-sm text-destructive">
							{m["exams.result.studyPlan.failed"]()}
						</p>
						<div>
							<SubmitButton
								type="button"
								pending={cached.isFetching}
								onClick={() => void cached.refetch()}
							>
								{m["exams.result.studyPlan.cta"]({ credits: studyPlanCost })}
							</SubmitButton>
						</div>
					</>
				) : (
					<>
						<p className="text-sm text-muted-foreground">
							{m["exams.result.studyPlan.hint"]({ credits: studyPlanCost })}
						</p>
						{generate.isError && !isQuotaError(generate.error) ? (
							<p className="text-sm text-destructive">
								{m["exams.result.studyPlan.failed"]()}
							</p>
						) : null}
						<div>
							<SubmitButton
								type="button"
								pending={generate.isPending}
								loadingLabel={m["exams.result.studyPlan.generating"]()}
								onClick={() => generate.mutate()}
							>
								{m["exams.result.studyPlan.cta"]({ credits: studyPlanCost })}
							</SubmitButton>
						</div>
					</>
				)}
			</div>
			<OutOfCreditsModal
				error={quotaError}
				onClose={() => setQuotaError(null)}
			/>
		</CardShell>
	);
}

/** Paid AI analysis for a completed attempt: the mistake review and the study
 *  plan, each charged at the per-attempt cost passed in via props. A
 *  previously generated document loads via the free cached GET on mount —
 *  no CTA, no credit spent. */
export function AiReviewPanel({
	attemptId,
	examReviewCost,
	studyPlanCost,
}: {
	attemptId: string;
	examReviewCost: number;
	studyPlanCost: number;
}) {
	return (
		<div className="flex flex-col gap-4">
			<ReviewSection attemptId={attemptId} examReviewCost={examReviewCost} />
			<PlanSection attemptId={attemptId} studyPlanCost={studyPlanCost} />
		</div>
	);
}
