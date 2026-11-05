export type IeltsQuestionType =
	| "true_false_not_given"
	| "matching"
	| "fill_blank";

export interface IeltsQuestion {
	id: string;
	number: number;
	type: IeltsQuestionType;
	prompt?: string;
	options?: string[];
	correctAnswer: string | string[];
	explanation: string;
}

export interface IeltsQuestionGroup {
	id: string;
	title?: string;
	instruction: string;
	passageRefId: string;
	isFlowchart?: boolean;
	questions: IeltsQuestion[];
}

export interface IeltsPassage {
	id: string;
	title: string;
	subtitle?: string;
	paragraphs: string[];
}

export interface IeltsSection {
	sectionNumber: number;
	title: string;
	passages: IeltsPassage[];
	groups: IeltsQuestionGroup[];
}

export interface IeltsExamTest {
	id: string;
	title: string;
	durationSeconds: number;
	totalQuestions: number;
	sections: IeltsSection[];
}

export function calculateIeltsBandScore(correct: number): number {
	if (correct >= 39) return 9.0;
	if (correct >= 37) return 8.5;
	if (correct >= 36) return 8.0;
	if (correct >= 34) return 7.5;
	if (correct >= 32) return 7.0;
	if (correct >= 30) return 6.5;
	if (correct >= 27) return 6.0;
	if (correct >= 23) return 5.5;
	if (correct >= 19) return 5.0;
	if (correct >= 15) return 4.5;
	if (correct >= 12) return 4.0;
	if (correct >= 9) return 3.5;
	if (correct >= 6) return 3.0;
	if (correct >= 3) return 2.5;
	return 2.0;
}

export function isAnswerCorrect(
	userAns: string | undefined | null,
	correctAns: string | string[],
): boolean {
	if (!userAns) return false;
	const trimmed = userAns.trim().toLowerCase();
	if (Array.isArray(correctAns)) {
		return correctAns.some((c) => c.trim().toLowerCase() === trimmed);
	}
	return correctAns.trim().toLowerCase() === trimmed;
}

export const GT_IELTS_TEST_1: IeltsExamTest = {
	id: "gt-ielts-collection-1-reading-test-1",
	title: "GT IELTS Collection 1 Reading Test 1",
	durationSeconds: 3600,
	totalQuestions: 40,
	sections: [
		{
			sectionNumber: 1,
			title: "Section 1",
			passages: [
				{
					id: "s1-p1",
					title: "Harry’s Car Rental",
					paragraphs: [
						"Harry’s Car Rental is a well-established independent company centrally located in the city, providing safe and reliable vehicles for all your travel requirements. Whether you need a short-term hire or a long-term rental, our competitive rates are tailored to your specific schedule and choice of vehicle.",
						"To ensure your peace of mind, basic insurance cover and roadside assistance are always included in our listed prices. We guarantee that every car is thoroughly cleaned and mechanically checked before your arrival, ensuring you drive away in a well-maintained vehicle that meets your exact needs. Whether you are starting an early-morning road trip or arriving late at night, our secure handover system allows you to pick up and return vehicles at our designated rental point at any time of day or night, guaranteeing a quick and easy process that fits your schedule.",
						"Our personal car rental service is ideal for families or individuals who need reliable everyday transport. Whether you are heading out on holiday, planning a cross-country road trip, or simply need a temporary ride while your own vehicle is being repaired, we have a solution to fit your lifestyle.",
						"For our corporate clients, business car rental is an excellent way to reduce the cost of maintaining a permanent company fleet. We offer a wide range of vehicle categories perfectly suited for client visits, staff travel, temporary projects, or replacement transport. Flexible rental periods allow businesses to hire vehicles for a single day, several weeks, or for regular use throughout the year. To help your company manage travel costs more efficiently, we also provide clear invoicing and streamlined booking arrangements.",
						"Please note that all drivers must be at least 18 years of age, possess a valid driver's license, and provide a standard security deposit at the time of booking. To ensure transparent pricing, all rentals come with a generous daily mileage allowance of 150 miles. We operate a fair full-to-full fuel policy, meaning you will receive the vehicle with a full tank of gas and simply need to return it full to avoid any extra refueling charges. To reserve your vehicle today, please contact us at 0829730101 or visit our website at harryrental.net.",
					],
				},
				{
					id: "s1-p2",
					title: "Local Galleries",
					paragraphs: [
						"A. Riverside Museum\nHoused in the Old Ferry Warehouse—originally built to store ropes and sails for river traders—the museum features seven galleries. Exhibits cover model boats, navigation tools, dock workers’ clothing, and the history of local shipbuilders.",
						"B. Clara Moseley’s House\nBuilt in 1822, this is the former home of pioneer photographer Clara Moseley. Active throughout the late 19th and early 20th centuries, Moseley was a trailblazer in documentary photography, women’s education, and heritage preservation. Visitors are taken on a 45-minute guided tour of the home and are then free to explore the small courtyard garden.",
						"C. Bramford Museum\nThe old mill town of Bramford grew rapidly after two large weaving sheds were built in 1848 and 1876. Today, the museum’s exhibits illustrate traditional spinning, dyeing, and hand weaving methods of local people, as well as the grueling daily lives and early labor union struggles of the Victorian mill workers. The museum has a touch-screen kiosk which contains over 8,000 historic photographs and texts, reproductions of which can be made on request.",
						"D. Eastgate Museum\nLocated in the middle of the market town of Eastgate, this museum is housed in a former primary school built in 1864. Its wide-ranging collections feature immersive reconstructions of an early 20th-century classroom, a nursery, and a children’s bedroom. Alongside an interesting display of handmade toys and school uniforms, visitors can read the fascinating original diary of pupil Arthur Bell, written during the evacuation of local children in 1941 to protect them from bombing raids.",
						"E. Northmere Museum\nSet in a beautifully preserved 19th-century building, the museum showcases regional scientific history, highlighting rare early electrical equipment, as well as the era’s railway boom. Its historical library contains the original notebooks of Victorian inventor Henry Vale and is open to the public on the first Saturday of each month. For deeper dives, the Research Room is available by appointment for studying local scientific records.",
					],
				},
			],
			groups: [
				{
					id: "g1",
					passageRefId: "s1-p1",
					instruction:
						"Do the following statements agree with the information given in the text? In boxes 1-6, choose:\nTRUE if the statement agrees with the information\nFALSE if the statement contradicts the information\nNOT GIVEN if there is no information on this",
					questions: [
						{
							id: "q1",
							number: 1,
							type: "true_false_not_given",
							prompt: "Customers have to pay separately for insurance cover.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "FALSE",
							explanation:
								"The text states: 'basic insurance cover and roadside assistance are always included in our listed prices.' Thus, customers do not pay separately.",
						},
						{
							id: "q2",
							number: 2,
							type: "true_false_not_given",
							prompt:
								"Customers are given a written report showing that the car has been inspected.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "NOT GIVEN",
							explanation:
								"The text mentions every car is mechanically checked before arrival, but there is no mention of a written inspection report being provided to customers.",
						},
						{
							id: "q3",
							number: 3,
							type: "true_false_not_given",
							prompt:
								"Customers can collect a car outside normal office hours.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "TRUE",
							explanation:
								"The text states: 'our secure handover system allows you to pick up and return vehicles... at any time of day or night.'",
						},
						{
							id: "q4",
							number: 4,
							type: "true_false_not_given",
							prompt: "Businesses are not limited to one fixed rental length.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "TRUE",
							explanation:
								"The text states: 'Flexible rental periods allow businesses to hire vehicles for a single day, several weeks, or for regular use throughout the year.'",
						},
						{
							id: "q5",
							number: 5,
							type: "true_false_not_given",
							prompt:
								"Drivers can return the car with less fuel without paying more.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "FALSE",
							explanation:
								"The text states: 'simply need to return it full to avoid any extra refueling charges.'",
						},
						{
							id: "q6",
							number: 6,
							type: "true_false_not_given",
							prompt: "The deposit made at the time of booking is refundable.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "NOT GIVEN",
							explanation:
								"The passage states that drivers must 'provide a standard security deposit at the time of booking', but it does not specify whether the deposit is refundable.",
						},
					],
				},
				{
					id: "g2",
					passageRefId: "s1-p2",
					instruction:
						"Look at the five descriptions of the museums A-E in one area of England. For which museum are the following statements true? NB You may use any letter more than once.",
					questions: [
						{
							id: "q7",
							number: 7,
							type: "matching",
							prompt:
								"The displays show how cloth was traditionally produced locally.",
							options: ["A", "B", "C", "D", "E"],
							correctAnswer: "C",
							explanation:
								"Bramford Museum (C): 'illustrate traditional spinning, dyeing, and hand weaving methods of local people'.",
						},
						{
							id: "q8",
							number: 8,
							type: "matching",
							prompt: "Sailing equipment used to be kept in this building.",
							options: ["A", "B", "C", "D", "E"],
							correctAnswer: "A",
							explanation:
								"Riverside Museum (A): 'Housed in the Old Ferry Warehouse—originally built to store ropes and sails for river traders'.",
						},
						{
							id: "q9",
							number: 9,
							type: "matching",
							prompt:
								"You can find information on the rise of one type of transport.",
							options: ["A", "B", "C", "D", "E"],
							correctAnswer: "E",
							explanation:
								"Northmere Museum (E): 'highlighting... the era’s railway boom'.",
						},
						{
							id: "q10",
							number: 10,
							type: "matching",
							prompt: "There are things to see both inside and outside.",
							options: ["A", "B", "C", "D", "E"],
							correctAnswer: "B",
							explanation:
								"Clara Moseley’s House (B): 'guided tour of the home and are then free to explore the small courtyard garden'.",
						},
						{
							id: "q11",
							number: 11,
							type: "matching",
							prompt:
								"Visitors can ask for copies of old pictures and documents.",
							options: ["A", "B", "C", "D", "E"],
							correctAnswer: "C",
							explanation:
								"Bramford Museum (C): 'touch-screen kiosk which contains over 8,000 historic photographs and texts, reproductions of which can be made on request'.",
						},
						{
							id: "q12",
							number: 12,
							type: "matching",
							prompt: "One section has a monthly public opening time.",
							options: ["A", "B", "C", "D", "E"],
							correctAnswer: "E",
							explanation:
								"Northmere Museum (E): 'Its historical library... is open to the public on the first Saturday of each month'.",
						},
						{
							id: "q13",
							number: 13,
							type: "matching",
							prompt:
								"Someone interested in protecting historic places lived here for a time.",
							options: ["A", "B", "C", "D", "E"],
							correctAnswer: "B",
							explanation:
								"Clara Moseley’s House (B): Clara Moseley was a trailblazer in 'heritage preservation' and lived in this home.",
						},
						{
							id: "q14",
							number: 14,
							type: "matching",
							prompt: "This museum has an account of a wartime event.",
							options: ["A", "B", "C", "D", "E"],
							correctAnswer: "D",
							explanation:
								"Eastgate Museum (D): 'visitors can read the fascinating original diary of pupil Arthur Bell, written during the evacuation of local children in 1941 to protect them from bombing raids'.",
						},
					],
				},
			],
		},
		{
			sectionNumber: 2,
			title: "Section 2",
			passages: [
				{
					id: "s2-p1",
					title: "THE LAW ON PAID HOLIDAY",
					paragraphs: [
						"Who qualifies for paid annual leave?\nAlmost all staff and workers in the UK have a legal right to paid leave, whatever the nature of their contract — full-time, part-time, agency or zero-hours. This right does not extend to those who are self-employed or who undertake unpaid work.",
						"How much leave should I get?\nThe statutory minimum is 5.6 weeks of paid leave per annum. For an employee working a five-day week, this is equivalent to 28 days of leave per year in total. Public holidays may be counted within this allowance, provided the contract specifies. Workers on reduced hours receive the same proportion — three days per week, for example, equates to 16.8 days annually. Employers may grant more generous terms than this, but the minimum cannot under any circumstances be undercut. Any leave not taken within the leave year is forfeited, unless the contract allows for its carry-over to the following twelve months.",
						"I think my leave has been refused unfairly. What can I do?\nWhere leave is being unjustly withheld, several courses of action remain open to you:\nIn the first instance, raise the matter with your line manager. The law is on your side, and a company in breach may be ordered to compensate the worker.\nMembers of a trade union may call upon their representative to act on their behalf.\nShould neither of these prove effective, ACAS (the Advisory, Conciliation and Arbitration Service) operates a free helpline for workplace disputes.\nYou have the statutory right to obtain from your employer a written record of all leave taken and any leave still owed to you. This must be supplied within two weeks of the request, either at the workplace or at another mutually convenient location. Where the employer does not comply with this requirement within the prescribed period, the matter may be taken to an employment tribunal. Any such complaint must be lodged within three months of the date on which the issue first arose.",
					],
				},
				{
					id: "s2-p2",
					title: "Sending Your Office Meeting Invites",
					paragraphs: [
						"Most office workers send several meeting invites each week. A careless one can waste everyone's morning, while a clear invite lets people plan ahead. Without a quick review before clicking send, you risk choosing the wrong people, the wrong time, or no clear purpose. A short check before each invite goes out keeps your colleagues' calendars healthy. The aim is to spot any issues, send the simple invites swiftly, and split bigger meetings into two steps. This guide covers a pre-send check and a two-step approach for larger meetings.",
						"Checking before you send\nBefore you click send, a quick review of your draft invite saves trouble for you and the people on the list. Unnecessary meetings — those that could be handled by a chat message or a simple question — should be skipped altogether. Three checks help you spot the rest before sending:\n• Check who is on the invite list. Does everyone really need to be there? Have you missed anyone whose input matters?\n• Check what the invite says. Is there a clear agenda? Are the items on the agenda in a sensible order?\n• Check the meeting time. Have you used the scheduling tool to spot any clash with key attendees?\nAfter this brief check, you should know which invites warrant extra preparation and which are ready to go straight out.",
						'Sending in two steps\nSome meetings — like a periodic review or a status update — have a clear topic and willing attendees, so their invites can be sent immediately, without further thought. However, for larger meetings requiring a formal agenda or senior-level input, use a two-step approach:\n• Send a meeting placeholder early to secure the date and time for core stakeholders. It is better to include all potential key participants in the invite list from the start and remove them later than to risk a cumbersome schedule conflict once their calendars fill up with other meetings.\n• Agree a firm deadline at the outset for the full details. When ready, simply update the existing calendar invite with the agenda and materials. Do not send extra reminder messages; just ensure the "Send Update" notification is triggered so the new info is reflected in everyone\'s schedule.\nThis secures the time immediately while allowing you the necessary space to finalize logistics, and avoids the chore of chasing updates.',
					],
				},
			],
			groups: [
				{
					id: "g3",
					passageRefId: "s2-p1",
					instruction:
						"Complete the sentences below. Choose NO MORE THAN TWO WORDS AND/OR A NUMBER from the text for each answer. Write your answers in boxes 15–21 on your answer sheet.",
					questions: [
						{
							id: "q15",
							number: 15,
							type: "fill_blank",
							prompt:
								"The right to paid holiday does not apply if you do (15)_______ or work for yourself.",
							correctAnswer: "unpaid work",
							explanation:
								"Passage: 'This right does not extend to those who are self-employed or who undertake unpaid work.'",
						},
						{
							id: "q16",
							number: 16,
							type: "fill_blank",
							prompt:
								"Every eligible worker is owed at least (16)_______ of leave in a twelve-month period.",
							correctAnswer: ["5.6 weeks", "28 days"],
							explanation:
								"Passage: 'The statutory minimum is 5.6 weeks of paid leave per annum. For an employee working a five-day week, this is equivalent to 28 days of leave per year in total.'",
						},
						{
							id: "q17",
							number: 17,
							type: "fill_blank",
							prompt:
								"Holiday days can be saved for the next year, but only if your (17)_______ permits this.",
							correctAnswer: "contract",
							explanation:
								"Passage: 'Any leave not taken within the leave year is forfeited, unless the contract allows for its carry-over to the following twelve months.'",
						},
						{
							id: "q18",
							number: 18,
							type: "fill_blank",
							prompt:
								"Part-time workers receive leave in (18)_______ to their hours.",
							correctAnswer: "proportion",
							explanation:
								"Passage: 'Workers on reduced hours receive the same proportion — three days per week, for example, equates to 16.8 days annually.'",
						},
						{
							id: "q19",
							number: 19,
							type: "fill_blank",
							prompt:
								"Workers who have been denied their full holiday should turn to their (19)_______.",
							correctAnswer: "line manager",
							explanation:
								"Passage: 'In the first instance, raise the matter with your line manager.'",
						},
						{
							id: "q20",
							number: 20,
							type: "fill_blank",
							prompt:
								"By law, you can see your company's (20)_______ of the leave you have used.",
							correctAnswer: ["written record", "record"],
							explanation:
								"Passage: 'You have the statutory right to obtain from your employer a written record of all leave taken and any leave still owed to you.'",
						},
						{
							id: "q21",
							number: 21,
							type: "fill_blank",
							prompt:
								"If the deadline for producing the record is missed, you have (21)_______ to bring a tribunal claim.",
							correctAnswer: ["three months", "3 months"],
							explanation:
								"Passage: 'Any such complaint must be lodged within three months of the date on which the issue first arose.'",
						},
					],
				},
				{
					id: "g4",
					passageRefId: "s2-p2",
					isFlowchart: true,
					instruction:
						"Complete the flow chart below. Choose NO MORE THAN TWO WORDS from the text for each answer. Write your answers in boxes 22–27 on your answer sheet.",
					questions: [
						{
							id: "q22",
							number: 22,
							type: "fill_blank",
							prompt:
								"Unnecessary meetings: Skip meetings that can be handled by a (22)_______ or simple question.",
							correctAnswer: "chat message",
							explanation:
								"Passage: 'Unnecessary meetings — those that could be handled by a chat message or a simple question — should be skipped altogether.'",
						},
						{
							id: "q23",
							number: 23,
							type: "fill_blank",
							prompt:
								"Checking what the invite says: Ensure there is a clear (23)_______ with items in sensible order.",
							correctAnswer: "agenda",
							explanation:
								"Passage: 'Check what the invite says. Is there a clear agenda? Are the items on the agenda in a sensible order?'",
						},
						{
							id: "q24",
							number: 24,
							type: "fill_blank",
							prompt:
								"Checking meeting time: Use scheduling tool to spot any (24)_______ with key attendees.",
							correctAnswer: ["clash", "schedule conflict"],
							explanation:
								"Passage: 'Check the meeting time. Have you used the scheduling tool to spot any clash with key attendees?'",
						},
						{
							id: "q25",
							number: 25,
							type: "fill_blank",
							prompt:
								"Two-step approach: Send a meeting (25)_______ early to secure date and time for core stakeholders.",
							correctAnswer: ["placeholder", "meeting placeholder"],
							explanation:
								"Passage: 'Send a meeting placeholder early to secure the date and time for core stakeholders.'",
						},
						{
							id: "q26",
							number: 26,
							type: "fill_blank",
							prompt:
								"Before updating invite: Agree a firm (26)_______ at the outset for the full details.",
							correctAnswer: ["deadline", "firm deadline"],
							explanation:
								"Passage: 'Agree a firm deadline at the outset for the full details.'",
						},
						{
							id: "q27",
							number: 27,
							type: "fill_blank",
							prompt:
								"When ready: Update calendar invite and ensure the (27)_______ notification is triggered.",
							correctAnswer: ["Send Update", "send update"],
							explanation:
								"Passage: 'just ensure the \"Send Update\" notification is triggered so the new info is reflected in everyone's schedule.'",
						},
					],
				},
			],
		},
		{
			sectionNumber: 3,
			title: "Section 3",
			passages: [
				{
					id: "s3-p1",
					title: "A survivor’s story",
					subtitle:
						"One native bird in New Zealand that has managed to survive the introduction of non-native species",
					paragraphs: [
						"As an island country with a fauna dominated by birds, New Zealand was once home to an owl species which is now extinct, the laughing owl, named for its distinctive cry. This bird was widespread throughout the islands when European settlers arrived in the middle of the 19th century and it remained in good numbers for some years thereafter. Where other native birds suffered from predation by the Polynesian rat, the laughing owl turned the tables and adapted its diet to include the rodent. It was also capable of catching and killing the other New Zealand owl, the morepork, and even larger birds, such as the weka. However, the laughing owl was wiped out around the beginning of the 20th century, its demise caused by specimen collectors, habitat changes, and non-native predators, including cats and stoats. Surprisingly, it is the smaller owl, the morepork, that has managed to survive until this day.",
						"Speckled dark brown, with yellow eyes and long tails, they are around 29 centimetres long from head to tail and 175 grams in weight. Moreporks have fringes on the edge of their feathers, so they can fly almost silently and not alert potential prey. They have acute hearing and their large eyes are very sensitive to light.",
						"Moreporks nest in tree hollows, in clumps of plants, or in cavities among rocks and roots. In the wild, moreporks usually start nesting in October, although two specimens have been recorded nesting in mid-winter, possibly stimulated by an ample food supply. The female lays up to three white eggs, which she incubates for 20 to 30 days. During this time, she rarely hunts, and the male brings food to her.",
						"Once the chicks hatch, she stays mainly in the nest until the young owls are fully feathered. When hatched, chicks are covered in light grey down and have their eyes closed. The eyes do not open until the eighth day after hatching. They can fly at around 35 days.",
						"By day, moreporks sleep in roosts. By night, they hunt a variety of animals – mainly large invertebrates including scarab and huhu beetles, moths, caterpillars and spiders. They also take small birds and mice. They can find suitable food in a pine forest as well as a native forest. A morepork uses its sharp talons to catch or stun its prey, which it then carries away in its bill. Moreporks are clever hunters, and birds such as robins, grey warblers and fantails can end up as their prey. In the day, these small birds sometimes mob drowsy moreporks and chase them away from their roosts; they force the sleepy predators to search for a more peaceful spot.",
						"Moreporks have proved to be ungracious hosts. Scientists trying to establish a population of plovers on Motuora Island in New Zealand’s Hauraki Gulf were mystified as to why only two birds survived out of the 75 placed there. The culprits turned out to be five pairs of moreporks that ate or chased away the new arrivals.",
						"Although moreporks are still considered to be relatively common, it is likely that numbers are in gradual decline due to predation and loss of habitat. As the female is a hole-nester, she is vulnerable to predators such as stoats and possums during the breeding season, and eggs and chicks will also be at risk from rats. The use of pesticides is another possible threat to the owls, though not a direct one. As moreporks are at the top of the food chain, they could be affected by an accumulative poison by consuming prey that has ingested poison.",
						"The New Zealand Department of Conservation is taking steps to ensure the preservation of New Zealand’s only native owl. The department is involved in measuring the population of moreporks and has put transmitters on a number of birds to determine survival and mortality. As well as being New Zealand’s only native owl, the morepork has symbolic and spiritual importance, so in monitoring the birds, it is hoped that the morepork will continue to survive and thrive.",
						"At dusk, the melancholy sound of the morepork can be heard in forests and parks as it calls to other moreporks and claims territory. Its Māori name (ruru) echoes its two-part cry. In the tradition of the Māori people of New Zealand, the morepork, or ruru, was often seen as a careful guardian. A number of sayings referred to the bird’s alertness. As a bird of the night, it was associated with the spirit world. Moreporks were believed to act as messengers to the gods in the heavens, flying along spiritual paths in the sky. They were the mediums used to communicate with the gods. The occasional high, piercing call of the morepork signified bad news, but the lower-pitched and more common “ruru” call heralded good news.",
					],
				},
			],
			groups: [
				{
					id: "g5",
					passageRefId: "s3-p1",
					instruction:
						"Do the following statements agree with the information given in the reading passage? In boxes 28 – 34 on your answer sheet, choose:\nTRUE if the statement agrees with the information\nFALSE if the statement contradicts the information\nNOT GIVEN if there is no information on this",
					questions: [
						{
							id: "q28",
							number: 28,
							type: "true_false_not_given",
							prompt:
								"Early European settlers made detailed studies of the morepork.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "NOT GIVEN",
							explanation:
								"The text states that the laughing owl was widespread when European settlers arrived, but makes no mention of detailed studies of the morepork being made by early settlers.",
						},
						{
							id: "q29",
							number: 29,
							type: "true_false_not_given",
							prompt:
								"The Polynesian rat had a negative effect on the number of laughing owls.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "FALSE",
							explanation:
								"The passage states: 'Where other native birds suffered from predation by the Polynesian rat, the laughing owl turned the tables and adapted its diet to include the rodent.'",
						},
						{
							id: "q30",
							number: 30,
							type: "true_false_not_given",
							prompt: "The laughing owl was larger than the morepork.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "TRUE",
							explanation:
								"The passage mentions the laughing owl was capable of catching and killing the morepork, and refers to 'the smaller owl, the morepork'.",
						},
						{
							id: "q31",
							number: 31,
							type: "true_false_not_given",
							prompt: "Rats pose a risk to young moreporks.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "TRUE",
							explanation:
								"Passage: 'and eggs and chicks will also be at risk from rats.'",
						},
						{
							id: "q32",
							number: 32,
							type: "true_false_not_given",
							prompt:
								"The New Zealand Department of Conservation is hoping to limit the population of moreporks.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "FALSE",
							explanation:
								"Passage: 'The New Zealand Department of Conservation is taking steps to ensure the preservation of New Zealand’s only native owl... it is hoped that the morepork will continue to survive and thrive.'",
						},
						{
							id: "q33",
							number: 33,
							type: "true_false_not_given",
							prompt:
								"Other bird species are frightened away when they hear the morepork’s cry.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "NOT GIVEN",
							explanation:
								"The text mentions other small birds mob drowsy moreporks during the day and chase them away, and discusses human/Māori beliefs about its cries, but does not state that other bird species are frightened away by its cry.",
						},
						{
							id: "q34",
							number: 34,
							type: "true_false_not_given",
							prompt:
								"In Māori tradition, the low call of the morepork had negative associations.",
							options: ["TRUE", "FALSE", "NOT GIVEN"],
							correctAnswer: "FALSE",
							explanation:
								"Passage: 'The occasional high, piercing call of the morepork signified bad news, but the lower-pitched and more common “ruru” call heralded good news.'",
						},
					],
				},
				{
					id: "g6",
					passageRefId: "s3-p1",
					instruction:
						"Complete the notes below. Choose ONE WORD AND/OR A NUMBER from the passage for each answer. Write your answers in boxes 35 – 40 on your answer sheet.",
					questions: [
						{
							id: "q35",
							number: 35,
							type: "fill_blank",
							prompt: "Appearance: approximately (35)_______ in length",
							correctAnswer: ["29 centimetres", "29 cm", "29"],
							explanation:
								"Passage: 'they are around 29 centimetres long from head to tail'.",
						},
						{
							id: "q36",
							number: 36,
							type: "fill_blank",
							prompt:
								"Nesting: nests in trees, plants or spaces in roots and (36)_______",
							correctAnswer: "rocks",
							explanation:
								"Passage: 'Moreporks nest in tree hollows, in clumps of plants, or in cavities among rocks and roots.'",
						},
						{
							id: "q37",
							number: 37,
							type: "fill_blank",
							prompt: "Hunting: transports its prey using its (37)_______",
							correctAnswer: "bill",
							explanation:
								"Passage: 'A morepork uses its sharp talons to catch or stun its prey, which it then carries away in its bill.'",
						},
						{
							id: "q38",
							number: 38,
							type: "fill_blank",
							prompt:
								"Hunting: can be chased away by other birds during the (38)_______",
							correctAnswer: "day",
							explanation:
								"Passage: 'In the day, these small birds sometimes mob drowsy moreporks and chase them away from their roosts'.",
						},
						{
							id: "q39",
							number: 39,
							type: "fill_blank",
							prompt:
								"Hunting: attacked (39)_______ that had been introduced to Motuora Island",
							correctAnswer: "plovers",
							explanation:
								"Passage: 'Scientists trying to establish a population of plovers on Motuora Island... five pairs of moreporks that ate or chased away the new arrivals.'",
						},
						{
							id: "q40",
							number: 40,
							type: "fill_blank",
							prompt:
								"Threats: may take in (40)_______ by eating prey affected by pesticides.",
							correctAnswer: "poison",
							explanation:
								"Passage: 'they could be affected by an accumulative poison by consuming prey that has ingested poison.'",
						},
					],
				},
			],
		},
	],
};
