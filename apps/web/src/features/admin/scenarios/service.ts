import type {
	CreatePersona,
	CreateScenario,
	CreateScenarioTopic,
	Pagination,
	Persona,
	Scenario,
	ScenarioTopic,
	UpdatePersona,
	UpdateScenario,
	UpdateScenarioTopic,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api";

export function listScenarios(
	page: number,
	pageSize: number,
): Promise<{ items: Scenario[]; pagination: Pagination }> {
	return apiPage<Scenario>(API_ROUTES.SCENARIOS.LIST, undefined, {
		query: { page, pageSize },
	});
}

export function createScenario(payload: CreateScenario): Promise<Scenario> {
	return api<Scenario>(API_ROUTES.SCENARIOS.LIST, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updateScenario(
	id: string,
	payload: UpdateScenario,
): Promise<Scenario> {
	return api<Scenario>(API_ROUTES.SCENARIOS.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deleteScenario(id: string): Promise<void> {
	return api<void>(API_ROUTES.SCENARIOS.BY_ID(id), { method: "DELETE" });
}

/** The topics endpoint always returns each banner with its preview
 *  scenarios attached — the admin table only reads id/slug/name/shortName/
 *  position and ignores the nested scenarios. */
export function listScenarioTopics(): Promise<ScenarioTopic[]> {
	return api<ScenarioTopic[]>(API_ROUTES.SCENARIOS.TOPICS);
}

export function createScenarioTopic(
	payload: CreateScenarioTopic,
): Promise<ScenarioTopic> {
	return api<ScenarioTopic>(API_ROUTES.SCENARIOS.TOPICS, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updateScenarioTopic(
	id: string,
	payload: UpdateScenarioTopic,
): Promise<ScenarioTopic> {
	return api<ScenarioTopic>(`${API_ROUTES.SCENARIOS.TOPICS}/${id}`, {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deleteScenarioTopic(id: string): Promise<void> {
	return api<void>(`${API_ROUTES.SCENARIOS.TOPICS}/${id}`, {
		method: "DELETE",
	});
}

export function listPersonas(
	page: number,
	pageSize: number,
): Promise<{ items: Persona[]; pagination: Pagination }> {
	return apiPage<Persona>(API_ROUTES.PERSONAS.LIST, undefined, {
		query: { page, pageSize },
	});
}

export function createPersona(payload: CreatePersona): Promise<Persona> {
	return api<Persona>(API_ROUTES.PERSONAS.LIST, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updatePersona(
	id: string,
	payload: UpdatePersona,
): Promise<Persona> {
	return api<Persona>(API_ROUTES.PERSONAS.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deletePersona(id: string): Promise<void> {
	return api<void>(API_ROUTES.PERSONAS.BY_ID(id), { method: "DELETE" });
}
