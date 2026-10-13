import { chromium } from "@playwright/test";

const base = "http://localhost:3000";
const lesson = "33333333-0000-4000-8000-000000000011";
const shots = [
	["reading", `${base}/lessons/${lesson}/parts/reading`],
	["listening", `${base}/lessons/${lesson}/parts/listening`],
	["writing", `${base}/lessons/${lesson}/parts/writing`],
	["speaking", `${base}/lessons/${lesson}/parts/speaking`],
	["detail", `${base}/lessons/${lesson}`],
	["hub", `${base}/lessons`],
];

const browser = await chromium.launch();
const context = await browser.newContext({
	storageState: "../playwright/.clerk/user.json",
	viewport: { width: 1440, height: 1000 },
});
const page = await context.newPage();
page.on("console", (msg) => {
	if (msg.type() === "error") console.log("CONSOLE ERROR:", msg.text());
});
for (const [name, url] of shots) {
	await page
		.goto(url, { waitUntil: "networkidle", timeout: 30000 })
		.catch((e) => console.log("goto failed", name, e.message));
	await page.waitForTimeout(1200);
	await page.screenshot({
		path: `/tmp/opencode/shot-${name}.png`,
		fullPage: false,
	});
	console.log("shot", name, "->", page.url());
}
await browser.close();
