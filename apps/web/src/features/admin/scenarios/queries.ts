import type {
	CreatePersona,
	CreateScenario,
	CreateScenarioTopic,
	UpdatePersona,
	UpdateScenario,
	UpdateScenarioTopic,
} from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	createPersona,
	createScenario,
	createScenarioTopic,
	deletePersona,
	deleteScenario,
	deleteScenarioTopic,
	listPersonas,
	listScenarios,
	listScenarioTopics,
	updatePersona,
	updateScenario,
	updateScenarioTopic,
} from "#/features/admin/scenarios/service";

const keys = {
	scenarios: (page: number, pageSize: number) =>
		["admin", "scenarios", page, pageSize] as const,
	topics: ["admin", "scenario-topics"] as const,
	personas: (page: number, pageSize: number) =>
		["admin", "personas", page, pageSize] as const,
};

function useInvalidate(key: readonly unknown[]) {
	const queryClient = useQueryClient();
	return () => queryClient.invalidateQueries({ queryKey: key });
}

export function useAdminScenarios(page: number, pageSize: number) {
	return useQuery({
		queryKey: keys.scenarios(page, pageSize),
		queryFn: () => listScenarios(page, pageSize),
	});
}

export function useCreateScenario() {
	const invalidate = useInvalidate(["admin", "scenarios"]);
	return useMutation({
		mutationFn: (payload: CreateScenario) => createScenario(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateScenario() {
	const invalidate = useInvalidate(["admin", "scenarios"]);
	return useMutation({
		mutationFn: ({ id, payload }: { id: string; payload: UpdateScenario }) =>
			updateScenario(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteScenario() {
	const invalidate = useInvalidate(["admin", "scenarios"]);
	return useMutation({
		mutationFn: (id: string) => deleteScenario(id),
		onSuccess: invalidate,
	});
}

export function useAdminScenarioTopics() {
	return useQuery({
		queryKey: keys.topics,
		queryFn: () => listScenarioTopics(),
	});
}

export function useCreateScenarioTopic() {
	const invalidate = useInvalidate(keys.topics);
	return useMutation({
		mutationFn: (payload: CreateScenarioTopic) => createScenarioTopic(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateScenarioTopic() {
	const invalidate = useInvalidate(keys.topics);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateScenarioTopic;
		}) => updateScenarioTopic(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteScenarioTopic() {
	const invalidate = useInvalidate(keys.topics);
	return useMutation({
		mutationFn: (id: string) => deleteScenarioTopic(id),
		onSuccess: invalidate,
	});
}

export function useAdminPersonas(page: number, pageSize: number) {
	return useQuery({
		queryKey: keys.personas(page, pageSize),
		queryFn: () => listPersonas(page, pageSize),
	});
}

export function useCreatePersona() {
	const invalidate = useInvalidate(["admin", "personas"]);
	return useMutation({
		mutationFn: (payload: CreatePersona) => createPersona(payload),
		onSuccess: invalidate,
	});
}

export function useUpdatePersona() {
	const invalidate = useInvalidate(["admin", "personas"]);
	return useMutation({
		mutationFn: ({ id, payload }: { id: string; payload: UpdatePersona }) =>
			updatePersona(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeletePersona() {
	const invalidate = useInvalidate(["admin", "personas"]);
	return useMutation({
		mutationFn: (id: string) => deletePersona(id),
		onSuccess: invalidate,
	});
}
