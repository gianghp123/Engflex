import { APP_ROUTES } from "#/app/app-route";
import type { NoteItem } from "./types";

export const MOCK_NOTES: NoteItem[] = [
	{
		id: "note-1",
		kind: "vocab",
		title: "Resilience",
		subtitle: "/rɪˈzɪliəns/ • Noun",
		summary: "The capacity to recover quickly from difficulties; toughness.",
		personalNote:
			"Hay gặp trong đề thi TOEIC Reading Part 7 chủ đề quản lý và nhân sự. Collocation: 'demonstrate remarkable resilience'.",
		tags: ["toeic", "advanced", "part7"],
		createdAt: "2026-11-05",
		linkTo: APP_ROUTES.VOCABULARY.LIST,
		extraInfo: {
			level: "C1",
			badge: "High Frequency",
		},
	},
	{
		id: "note-2",
		kind: "lesson",
		title: "Mastering Conditionals Type 2 & 3",
		subtitle: "Grammar Core • Level B2",
		summary:
			"Learn unreal conditions in the present and past with clear contrast tables and real-world conversation examples.",
		personalNote:
			"Nhớ công thức Type 3: If + S + had V3, S + would have V3. Hay bị bẫy trong câu hỏi hoàn thành câu TOEIC Part 5!",
		tags: ["grammar", "conditionals", "b2"],
		createdAt: "2026-11-04",
		linkTo: APP_ROUTES.LESSONS.LIST,
		extraInfo: {
			level: "B2",
			badge: "Grammar",
		},
	},
	{
		id: "note-3",
		kind: "video",
		title: "How to Deliver an Impressive Self-Introduction",
		subtitle: "Business English • 08:42",
		summary:
			"Step-by-step breakdown of structure, vocabulary, and pitch control for international job interviews.",
		personalNote:
			"Luyện Shadowing đoạn 02:15 - 04:30 về cách tóm tắt kinh nghiệm làm việc bằng action verbs mạnh.",
		tags: ["interview", "speaking", "business"],
		createdAt: "2026-11-03",
		linkTo: APP_ROUTES.VIDEOS.LIST,
		extraInfo: {
			duration: "08:42",
			badge: "Shadowing",
		},
	},
	{
		id: "note-4",
		kind: "story",
		title: "The Clockmaker's Secret",
		subtitle: "CEFR B1 • Mystery Short Story",
		summary:
			"An apprentice uncovers an intricate brass automaton hidden in the clock tower attic during a stormy night.",
		personalNote:
			"Học được các tính từ miêu tả không gian rất hay: 'intricate', 'whirring', 'damp attic'. Đọc lại trước khi đi ngủ để ngấm ngữ điệu.",
		tags: ["story", "mystery", "b1"],
		createdAt: "2026-11-02",
		linkTo: APP_ROUTES.STORIES.LIST,
		extraInfo: {
			level: "B1",
			badge: "Audio Included",
		},
	},
	{
		id: "note-5",
		kind: "vocab",
		title: "Negotiate",
		subtitle: "/nɪˈɡoʊʃieɪt/ • Verb",
		summary: "To obtain or arrange by discussion, conference, or compromise.",
		personalNote:
			"Cấu trúc: negotiate with somebody for/over something. Danh từ: negotiation.",
		tags: ["collocation", "business", "toeic"],
		createdAt: "2026-10-30",
		linkTo: APP_ROUTES.VOCABULARY.LIST,
		extraInfo: {
			level: "B2",
			badge: "Business",
		},
	},
	{
		id: "note-6",
		kind: "video",
		title: "5 Linking Words That Make You Sound Like a Native",
		subtitle: "Pronunciation & Fluency • 06:15",
		summary:
			"Natural transitions including 'having said that', 'to put it another way', and 'furthermore'.",
		personalNote:
			"Thực hành áp dụng ngay cụm 'having said that' trong phòng luyện nói AI Voice tiếp theo.",
		tags: ["voice", "linking-words", "fluency"],
		createdAt: "2026-10-28",
		linkTo: APP_ROUTES.VIDEOS.LIST,
		extraInfo: {
			duration: "06:15",
			badge: "Speaking Tip",
		},
	},
	{
		id: "note-7",
		kind: "lesson",
		title: "TOEIC Listening Part 2: Indirect Responses",
		subtitle: "Exam Strategy • Target 650+",
		summary:
			"Decode conversational traps where answers do not use direct yes/no statements or repeat prompt keywords.",
		personalNote:
			"Cực kỳ cẩn thận với câu hỏi Wh-question mà đáp án là 'Check the schedule board' hoặc 'Ask Maria'!",
		tags: ["toeic", "part2", "listening"],
		createdAt: "2026-10-25",
		linkTo: APP_ROUTES.LESSONS.LIST,
		extraInfo: {
			level: "B1",
			badge: "Strategy",
		},
	},
	{
		id: "note-8",
		kind: "story",
		title: "A Cup of Coffee in Kyoto",
		subtitle: "CEFR A2 • Slice of Life",
		summary:
			"A quiet morning encounter in a small alleyway café near the Kamo river in the autumn breeze.",
		personalNote:
			"Bài đọc câu ngắn, từ vựng thân thiện, rất thích hợp để luyện phát âm trôi chảy.",
		tags: ["reading", "travel", "a2"],
		createdAt: "2026-10-20",
		linkTo: APP_ROUTES.STORIES.LIST,
		extraInfo: {
			level: "A2",
			badge: "Relaxing",
		},
	},
];
