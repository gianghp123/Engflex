import { chromium } from "@playwright/test";

const base = "http://localhost:3000";
const lesson = "33333333-0000-4000-8000-000000000011";
const url = `${base}/lessons/${lesson}/parts/reading`;

const browser = await chromium.launch();
const context = await browser.newContext({
	storageState: "../playwright/.clerk/user.json",
	viewport: { width: 1440, height: 1000 },
});
const page = await context.newPage();
page.on("response", async (res) => {
	if (res.status() >= 400) {
		console.log("HTTP", res.status(), res.request().method(), res.url());
		try {
			const text = await res.text();
			console.log("  body:", text.slice(0, 300));
		} catch {}
	}
});
page.on("console", (msg) => {
	if (msg.type() === "error") console.log("CONSOLE:", msg.text().slice(0, 500));
});
await page
	.goto(url, { waitUntil: "networkidle", timeout: 30000 })
	.catch((e) => console.log("goto failed:", e.message));
await page.waitForTimeout(2500);
await page.screenshot({ path: "/tmp/opencode/shot-reading.png" });
console.log("final url:", page.url());
const body = await page
	.locator("body")
	.innerText()
	.catch(() => "");
console.log("body text head:", body.slice(0, 400));
await browser.close();
