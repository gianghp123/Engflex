import type {
	Conversation,
	ConversationSession,
	ConversationSummary,
	Feedback,
	Pagination,
	Persona,
	Scenario,
	ScenarioTopic,
	StartConversation,
	TranscribeResult,
	TranscriptResult,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api";

/** The three actions of the correction modal, in the engine's vocabulary. */
export type TranscriptAction = "review" | "send" | "dismiss";

export function createConversation(
	input: StartConversation,
): Promise<Conversation> {
	return api<Conversation>(
		API_ROUTES.CONVERSATIONS.CREATE,
		{
			method: "POST",
			body: JSON.stringify(input),
		},
		{ withCredentials: true },
	);
}

export function getConversation(id: string): Promise<Conversation> {
	return api<Conversation>(API_ROUTES.CONVERSATIONS.BY_ID(id), undefined, {
		withCredentials: true,
	});
}

export function endConversation(id: string): Promise<Conversation> {
	return api<Conversation>(
		API_ROUTES.CONVERSATIONS.END(id),
		{ method: "POST" },
		{ withCredentials: true },
	);
}

export type { Conversation, ConversationSession };

export function analyzeTurn(
	conversationId: string,
	position: number,
): Promise<Feedback> {
	return api<Feedback>(
		API_ROUTES.CONVERSATIONS.ANALYZE_TURN(conversationId, position),
		{ method: "POST" },
		{ withCredentials: true },
	);
}

/**
 * The correction modal's one command. The engine owns the reviewed turn and
 * resolves which turn is corrected, so nothing here names a turn.
 */
export function transcriptCommand(
	conversationId: string,
	action: TranscriptAction,
	text?: string,
): Promise<TranscriptResult> {
	return api<TranscriptResult>(
		API_ROUTES.CONVERSATIONS.TRANSCRIPT(conversationId),
		{ method: "POST", body: JSON.stringify({ action, text: text ?? "" }) },
		{ withCredentials: true },
	);
}

/**
 * A recorded re-speak for the correction modal. Multipart, so the browser
 * sets its own boundary; the sentence comes back for the field and nothing
 * is committed.
 */
export function transcribeAudio(
	conversationId: string,
	audio: Blob,
	filename: string,
): Promise<TranscribeResult> {
	const form = new FormData();
	form.append("audio", audio, filename);
	return api<TranscribeResult>(
		API_ROUTES.CONVERSATIONS.TRANSCRIBE(conversationId),
		{ method: "POST", body: form },
		{ withCredentials: true },
	);
}

export function listScenarios(params?: {
	topicId?: string;
}): Promise<{ items: Scenario[]; pagination: Pagination }> {
	const query: Record<string, string> = {};
	if (params?.topicId) query.topicId = params.topicId;
	return apiPage<Scenario>(
		API_ROUTES.SCENARIOS.LIST,
		{ method: "GET" },
		{ withCredentials: true, query },
	);
}

/** Every topic in position order, each carrying up to `previewK` eligible
 * scenarios with their partner persona — one banner per topic. */
export function listTopicsWithPreview(params?: {
	previewK?: number;
	difficulty?: string;
	search?: string;
}): Promise<ScenarioTopic[]> {
	const query: Record<string, string> = {};
	query.previewK = String(params?.previewK ?? 3);
	if (params?.difficulty) query.difficulty = params.difficulty;
	if (params?.search?.trim()) query.search = params.search.trim();
	return api<ScenarioTopic[]>(
		API_ROUTES.SCENARIOS.TOPICS,
		{ method: "GET" },
		{ withCredentials: true, query },
	);
}

export function getScenarioById(id: string): Promise<Scenario> {
	return api<Scenario>(API_ROUTES.SCENARIOS.BY_ID(id), undefined, {
		withCredentials: true,
	});
}

export function listPersonas(): Promise<{
	items: Persona[];
	pagination: Pagination;
}> {
	return apiPage<Persona>(
		API_ROUTES.PERSONAS.LIST,
		{ method: "GET" },
		{ withCredentials: true },
	);
}

/** One page of the caller's finished conversations, newest first. */
export function listConversations(params: {
	scenarioId?: string;
	page?: number;
	pageSize?: number;
}): Promise<{ items: ConversationSummary[]; pagination: Pagination }> {
	const query: Record<string, string> = {};
	if (params.scenarioId) query.scenarioId = params.scenarioId;
	if (params.page) query.page = String(params.page);
	if (params.pageSize) query.pageSize = String(params.pageSize);
	return apiPage<ConversationSummary>(
		API_ROUTES.CONVERSATIONS.LIST,
		{ method: "GET" },
		{ withCredentials: true, query },
	);
}
