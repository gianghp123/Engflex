import type {
	VideoCategoryResponse,
	VideoTranscriptResponse,
} from "@engflex/contracts";
import type { Video, VideoSection } from "./types";

const youtube = (id: string) => ({
	videoUrl: `https://www.youtube.com/watch?v=${id}`,
	thumbnailUrl: `https://i.ytimg.com/vi/${id}/hqdefault.jpg`,
	source: "youtube" as const,
});

const MOVIES: VideoCategoryResponse = {
	id: "8f6c1d2e-0a41-4c55-9d2b-1f0e7a3c5b01",
	slug: "movie-short-clip",
	name: "Movie short clip",
	createdAt: "2026-10-13T10:41:05+07:00",
	updatedAt: "2026-10-13T10:41:05+07:00",
};
const TED: VideoCategoryResponse = {
	id: "8f6c1d2e-0a41-4c55-9d2b-1f0e7a3c5b02",
	slug: "ted-talks",
	name: "TED Talks",
	createdAt: "2026-10-13T10:41:05+07:00",
	updatedAt: "2026-10-13T10:41:05+07:00",
};
const TALKS: VideoCategoryResponse = {
	id: "8f6c1d2e-0a41-4c55-9d2b-1f0e7a3c5b03",
	slug: "learning-tech-talks",
	name: "Learning & tech talks",
	createdAt: "2026-10-13T10:41:05+07:00",
	updatedAt: "2026-10-13T10:41:05+07:00",
};

function video(
	category: VideoCategoryResponse,
	fields: Omit<Video, "categoryId" | "category" | "description"> &
		Partial<Pick<Video, "description">>,
): Video {
	return {
		description: "",
		...fields,
		categoryId: category.id,
		category,
	};
}

export const VIDEO_SECTIONS: VideoSection[] = [
	{
		category: MOVIES,
		total: 157,
		videos: [
			video(MOVIES, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000101",
				title: "Princess Mononoke | Official Trailer",
				...youtube("4OiMOHRDs14"),
				cefrLevel: "A2",
				duration: 64,
				views: 63026,
				isPro: true,
				practice: { dictation: true, shadowing: false },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
			video(MOVIES, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000102",
				title: "Kiki's Delivery Service | Official Trailer",
				...youtube("4bG17OYs-GA"),
				cefrLevel: "B1",
				duration: 50,
				views: 97991,
				practice: { dictation: false, shadowing: false },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
			video(MOVIES, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000103",
				title: "Toy Story 4 | Official Trailer",
				...youtube("wmiIUN-7qhE"),
				cefrLevel: "B2",
				duration: 113,
				views: 46444,
				practice: { dictation: true, shadowing: true },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
			video(MOVIES, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000104",
				title: "Zootopia | Official US Trailer #2",
				...youtube("jWM0ct-OLsM"),
				cefrLevel: "B1",
				duration: 175,
				views: 22605,
				isPro: true,
				practice: { dictation: false, shadowing: false },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
		],
	},
	{
		category: TED,
		total: 148,
		videos: [
			video(TED, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000201",
				title: "How to speak so that people want to listen",
				...youtube("eIho2S0ZahI"),
				cefrLevel: "B1",
				duration: 598,
				views: 140846,
				practice: { dictation: false, shadowing: true },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
			video(TED, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000202",
				title: "Inside the mind of a master procrastinator",
				...youtube("arj7oStGLkU"),
				cefrLevel: "B2",
				duration: 844,
				views: 49923,
				practice: { dictation: false, shadowing: false },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
			video(TED, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000203",
				title: "Your body language may shape who you are",
				...youtube("Ks-_Mh1QhMc"),
				cefrLevel: "C1",
				duration: 1263,
				views: 23413,
				isPro: true,
				practice: { dictation: true, shadowing: false },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
			video(TED, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000204",
				title: "Do schools kill creativity?",
				...youtube("iG9CE55wbtY"),
				cefrLevel: "B2",
				duration: 1204,
				views: 39404,
				practice: { dictation: false, shadowing: false },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
		],
	},
	{
		category: TALKS,
		total: 36,
		videos: [
			video(TALKS, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000301",
				title: "Learn English in 30 minutes: all the basics you need",
				...youtube("juKd26qkNAw"),
				cefrLevel: "A1",
				duration: 1843,
				views: 312877,
				practice: { dictation: true, shadowing: true },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
			video(TALKS, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000302",
				title: "What the heck is the event loop anyway?",
				...youtube("8aGhZQkoFbQ"),
				cefrLevel: "C1",
				duration: 1570,
				views: 18230,
				isPro: true,
				practice: { dictation: false, shadowing: false },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
			video(TALKS, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000303",
				title: "Steve Jobs' 2005 Stanford commencement address",
				...youtube("UF8uR6Z6KLc"),
				cefrLevel: "B2",
				duration: 905,
				views: 87512,
				practice: { dictation: false, shadowing: true },
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
			video(TALKS, {
				id: "c0a8e1f4-1b2c-4d3e-8f90-000000000304",
				title: "Stand-up meeting: giving a clear status update",
				videoUrl: "",
				thumbnailUrl: "",
				cefrLevel: "B2",
				duration: 142,
				createdAt: "2026-10-13T10:41:05+07:00",
				updatedAt: "2026-10-13T10:41:05+07:00",
			}),
		],
	},
];

const LINES = [
	{
		content: "I'm Kiki, and this is my cat, Jiji.",
		phonetic: "aɪm ˈkiːki ænd ðɪs ɪz maɪ kæt ˈdʒiːdʒi",
		vietnamese: "Mình là Kiki, còn đây là mèo của mình, Jiji.",
	},
	{
		content: "We just flew in from the countryside last night.",
		phonetic: "wiː dʒʌst fluː ɪn frɒm ðə ˈkʌntrisaɪd lɑːst naɪt",
		vietnamese: "Tụi mình vừa bay từ vùng quê tới đây tối qua.",
	},
	{
		content: "I need to find a town to live in for a year.",
		phonetic: "aɪ niːd tə faɪnd ə taʊn tə lɪv ɪn fər ə jɪə",
		vietnamese: "Mình cần tìm một thị trấn để sống trong một năm.",
	},
	{
		content: "Is there a place where a young witch could stay?",
		phonetic: "ɪz ðeər ə pleɪs weər ə jʌŋ wɪtʃ kʊd steɪ",
		vietnamese: "Có chỗ nào cho một phù thủy nhỏ ở lại không?",
	},
	{
		content: "You can stay in the room above the bakery.",
		phonetic: "juː kən steɪ ɪn ðə ruːm əˈbʌv ðə ˈbeɪkəri",
		vietnamese: "Cháu có thể ở căn phòng phía trên tiệm bánh.",
	},
	{
		content: "I'll start a delivery service with my broom.",
		phonetic: "aɪl stɑːt ə dɪˈlɪvəri ˈsɜːvɪs wɪð maɪ bruːm",
		vietnamese: "Mình sẽ mở dịch vụ giao hàng bằng cây chổi của mình.",
	},
	{
		content: "Please be careful with this package, it's a birthday present.",
		phonetic: "pliːz bi ˈkeəfəl wɪð ðɪs ˈpækɪdʒ ɪts ə ˈbɜːθdeɪ ˈpreznt",
		vietnamese: "Cẩn thận với gói hàng này nhé, đó là quà sinh nhật.",
	},
	{
		content: "Thank you so much, I'll deliver it right away!",
		phonetic: "θæŋk juː səʊ mʌtʃ aɪl dɪˈlɪvər ɪt raɪt əˈweɪ",
		vietnamese: "Cảm ơn nhiều lắm, mình sẽ giao ngay đây!",
	},
];

/** Mock transcript shared by every fixture video until the API serves real lines. */
export const DICTATION_TRANSCRIPTS: Omit<
	VideoTranscriptResponse,
	"videoExerciseId"
>[] = LINES.map((line, index) => ({
	id: `5d1e7c40-2f3a-4b6c-9e8d-${String(index + 1).padStart(12, "0")}`,
	sequence: index + 1,
	...line,
	startTimestamp: index * 4,
	endTimestamp: index * 4 + 3.5,
	createdAt: "2026-10-13T10:41:05+07:00",
	updatedAt: "2026-10-13T10:41:05+07:00",
}));

export function getVideoById(id: string): Video | undefined {
	const found = VIDEO_SECTIONS.flatMap((section) => section.videos).find(
		(item) => item.id === id,
	);
	if (!found) return undefined;
	return {
		...found,
		transcripts: DICTATION_TRANSCRIPTS.map((line) => ({
			...line,
			videoExerciseId: found.id,
		})),
	};
}
