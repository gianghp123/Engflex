import type {
	Collocation,
	VocabularyDetails,
	VocabularyItem,
	VocabularyStats,
	WordForm,
} from "@engflex/contracts";

const FORM_DEFAULT: WordForm[] = [{ formType: "Noun", forms: ["—"] }];

/** mock-only "today" so due-date comparisons stay deterministic (no clock). */
export const TODAY = "2026-09-23";

interface DetailsInput {
	contextLabel: string;
	contextQuote: string;
	longDefinition: string;
	syllables?: string;
	stressTip?: string;
	commonSlip?: string;
	etymology?: string;
	collocations?: Collocation[];
	wordForms?: WordForm[];
}

/** Fills the contract's required `VocabularyDetails` fields for hub-only cards. */
function details(input: DetailsInput): VocabularyDetails {
	return {
		audioUs: "",
		audioUk: "",
		syllables: input.syllables ?? "—",
		stressTip:
			input.stressTip ??
			"Keep the primary stress on the technical root syllable.",
		commonSlip:
			input.commonSlip ??
			"Avoid flattening the technical vowels when the term is used mid-sentence.",
		longDefinition: input.longDefinition,
		etymology: input.etymology ?? "Modern professional and technical English.",
		contexts: [{ label: input.contextLabel, quote: input.contextQuote }],
		collocations: input.collocations ?? [],
		wordForms: input.wordForms ?? FORM_DEFAULT,
	};
}

/** Card/definition/context copy transcribed from the vocabulary-hub mock. */
export const VOCABULARY_ITEMS: VocabularyItem[] = [
	{
		id: "idempotent",
		term: "idempotent",
		ipa: "/ˌaɪ.dɛmˈpoʊ.tənt/",
		partOfSpeech: "adj",
		cefr: "C1",
		status: "published",
		domain: "backend_db",
		definition:
			"Denoting an operation or API request that can be applied multiple times without changing the result beyond the initial execution.",
		details: {
			audioUs: "",
			audioUk: "",
			syllables: "eye · dem · Poh · tent",
			stressTip:
				"Primary stress sits firmly on the third syllable /poʊ/, not the second syllable.",
			commonSlip:
				'Placing stress on "-dem-" (*eye-DEM-puh-tent). Ensure the diphthong /oʊ/ remains audible and crisp.',
			longDefinition:
				"Denoting an operation, API request, or mathematical function that can be applied multiple consecutive times without changing the result beyond the initial application. In HTTP specifications, methods like GET, PUT, and DELETE are guaranteed to be idempotent.",
			etymology:
				"Formed from Latin idem (the same) + potens (having power). Coined in mathematics by Benjamin Peirce (1870).",
			contexts: [
				{
					label: "Webhook delivery & payment processing",
					quote:
						"Ensure that our payment webhook handlers are strictly idempotent using unique deduplication keys so network retries do not trigger duplicate charges for customer orders.",
				},
				{
					label: "REST API RFC & interface specifications",
					quote:
						"While POST requests are inherently non-idempotent, we must enforce an Idempotency-Key header on this checkout endpoint to guarantee replay safety.",
				},
				{
					label: "Database upserts & transactional outbox",
					quote:
						"Making the consumer worker's state transition idempotent via an upsert pattern completely eliminated our race condition during cluster failovers.",
				},
			],
			collocations: [
				{
					phrase: "Strictly idempotent",
					pattern: "adv + adj",
					example:
						"The ingestion pipeline must remain strictly idempotent to handle message queue re-deliveries.",
				},
				{
					phrase: "Guarantee idempotency",
					pattern: "verb + noun",
					example:
						"We use Redis distributed locks to guarantee idempotency across concurrent client requests.",
				},
				{
					phrase: "Idempotent consumer",
					pattern: "adj + noun",
					example:
						"Designing idempotent consumers is a fundamental pattern in Kafka-based distributed pipelines.",
				},
				{
					phrase: "Inherent idempotency",
					pattern: "adj + noun",
					example:
						"HTTP PUT carries inherent idempotency because replacing a resource yields an identical state.",
				},
			],
			wordForms: [
				{ formType: "Noun (abstract)", forms: ["idempotence", "idempotency"] },
				{ formType: "Adverb", forms: ["idempotently"] },
				{ formType: "Antonym", forms: ["non-idempotent"] },
			],
		},
		userState: {
			itemId: "idempotent",
			sourceType: "lesson",
			sourceId: "Distributed Systems Architecture",
			mastered: false,
			srsDueAt: "2026-09-24T09:00:00Z",
			createdAt: "2026-05-18T09:00:00Z",
			note: "Remember to highlight this term when explaining message broker retry logic to the mobile team in next week's architecture sync.",
		},
	},
	{
		id: "backpressure",
		term: "backpressure",
		ipa: "/ˈbækˌprɛʃ.ər/",
		partOfSpeech: "noun",
		cefr: "B2",
		status: "published",
		domain: "distributed_systems",
		definition:
			"The resistance or opposing force exerted opposite to the desired flow of data in a software pipeline when a downstream consumer cannot handle the ingress rate.",
		details: details({
			contextLabel: "Engineering context",
			contextQuote:
				"When the queue worker buffer saturated, the reactive stream applied backpressure to throttle the upstream producer before memory spikes crashed the container.",
			longDefinition:
				"The resistance or opposing force exerted opposite to the desired flow of data in a software pipeline when a downstream consumer cannot handle the ingress rate.",
			syllables: "back · presh · er",
		}),
		userState: {
			itemId: "backpressure",
			sourceType: "conversation",
			mastered: false,
			srsDueAt: "2026-09-23T09:00:00Z",
			createdAt: "2026-06-02T09:00:00Z",
		},
	},
	{
		id: "bottleneck",
		term: "bottleneck",
		ipa: "/ˈbɑː.t̬əl.nek/",
		partOfSpeech: "noun",
		cefr: "B1",
		status: "published",
		domain: "backend_db",
		definition:
			"A point of congestion in a production system that occurs when workloads arrive too quickly for the designated computing process to handle.",
		details: details({
			contextLabel: "Engineering context",
			contextQuote:
				"Our latency profiling identified the unindexed Postgres foreign key query as the primary execution bottleneck during peak checkout spikes.",
			longDefinition:
				"A point of congestion in a production system that occurs when workloads arrive too quickly for the designated computing process to handle.",
			syllables: "bot · tl · neck",
		}),
		userState: {
			itemId: "bottleneck",
			sourceType: "manual",
			mastered: true,
			createdAt: "2026-06-10T09:00:00Z",
			note: "Added during team retrospective",
		},
	},
	{
		id: "re-entrancy",
		term: "re-entrancy",
		ipa: "/riːˈɛn.trən.si/",
		partOfSpeech: "noun",
		cefr: "C1",
		status: "published",
		domain: "distributed_systems",
		definition:
			"The property of a computer program or routine allowing it to be safely executed concurrently or interrupted and called again before its previous invocation finishes.",
		details: details({
			contextLabel: "Engineering context",
			contextQuote:
				"We mitigated the classical re-entrancy vulnerability by enforcing the Checks-Effects-Interactions design pattern and locking state mutators.",
			longDefinition:
				"The property of a computer program or routine allowing it to be safely executed concurrently or interrupted and called again before its previous invocation finishes.",
			syllables: "re · en · tran · cy",
		}),
		userState: {
			itemId: "re-entrancy",
			sourceType: "lesson",
			sourceId: "Smart Contract & Concurrency Pitfalls",
			mastered: false,
			srsDueAt: "2026-09-25T09:00:00Z",
			createdAt: "2026-06-14T09:00:00Z",
		},
	},
	{
		id: "trade-off",
		term: "trade-off",
		ipa: "/ˈtreɪd.ɑːf/",
		partOfSpeech: "noun",
		cefr: "B1",
		status: "published",
		domain: "frontend_ui",
		definition:
			"A situational decision that involves diminishing or losing one desirable quality or aspect of a software system in return for gaining qualities in another aspect.",
		details: details({
			contextLabel: "Engineering context",
			contextQuote:
				"Choosing an eventually consistent NoSQL cluster is a deliberate trade-off prioritizing global read throughput over immediate serializability.",
			longDefinition:
				"A situational decision that involves diminishing or losing one desirable quality or aspect of a software system in return for gaining qualities in another aspect.",
			syllables: "trade · off",
		}),
		userState: {
			itemId: "trade-off",
			sourceType: "manual",
			mastered: false,
			createdAt: "2026-06-20T09:00:00Z",
			note: "Technical leadership vocabulary",
		},
	},
	{
		id: "concurrency",
		term: "concurrency",
		ipa: "/kənˈkɜː.ən.si/",
		partOfSpeech: "noun",
		cefr: "B2",
		status: "published",
		domain: "distributed_systems",
		definition:
			"Not parallelism. The composition of independently executing computations.",
		details: details({
			contextLabel: "Engineering context",
			contextQuote:
				"Locking discipline is what keeps concurrency safe when several workers mutate the same aggregate.",
			longDefinition:
				"The composition of independently executing computations — distinct from parallelism, which requires simultaneous execution.",
			syllables: "con · cur · ren · cy",
		}),
		userState: {
			itemId: "concurrency",
			sourceType: "manual",
			mastered: false,
			createdAt: "2026-09-23T10:14:00Z",
			note: "Added today, 10:14 AM",
		},
	},
	{
		id: "heuristics",
		term: "heuristics",
		ipa: "/hjʊˈrɪs.tɪks/",
		partOfSpeech: "noun",
		cefr: "C1",
		status: "published",
		domain: "ai_ml",
		definition:
			"Practical rule-of-thumb method not guaranteed to be optimal, but sufficient for immediate goals.",
		details: details({
			contextLabel: "Engineering context",
			contextQuote:
				"We replaced the hand-tuned heuristics with a learned ranking model once the offline evaluation caught up.",
			longDefinition:
				"Practical rule-of-thumb methods that are not guaranteed to be optimal but are sufficient for immediate goals.",
			syllables: "heu · ris · tics",
		}),
		userState: {
			itemId: "heuristics",
			sourceType: "manual",
			mastered: false,
			createdAt: "2026-09-22T09:00:00Z",
			note: "Added yesterday",
		},
	},
	{
		id: "fault-tolerance",
		term: "fault tolerance",
		ipa: "/fɔːlt ˈtɒl.ər.əns/",
		partOfSpeech: "noun",
		cefr: "B1",
		status: "published",
		domain: "devops_cloud",
		definition:
			"Enabling a system to continue operating properly in the event of the failure of one or more components.",
		details: details({
			contextLabel: "Engineering context",
			contextQuote:
				"Fault tolerance here means draining a zone without dropping in-flight checkout sessions.",
			longDefinition:
				"Enabling a system to continue operating properly in the event of the failure of one or more of its components.",
			syllables: "fault · tol · er · ance",
		}),
		userState: {
			itemId: "fault-tolerance",
			sourceType: "manual",
			mastered: false,
			createdAt: "2026-09-20T09:00:00Z",
			note: "Added 3 days ago",
		},
	},
];

/** Aggregated hub stats, transcribed from the mock. */
export const VOCABULARY_STATS: VocabularyStats = {
	totalSaved: 284,
	mastered: 192,
	needsReview: 32,
	customAdditions: 46,
	dueToday: 10,
	retentionPct: 68,
};
