import { createHmac, randomUUID } from "node:crypto";
import { createClerkClient } from "@clerk/backend";
import { clerk, clerkSetup } from "@clerk/testing/playwright";
import { expect, test } from "@playwright/test";
import { clerkUserId, closeDb, db } from "./helpers";

let adminId = "";
let createdProfile = false;
const adminEmail = () => {
	const email = process.env.E2E_ADMIN_CLERK_USER_EMAIL;
	if (!email)
		throw new Error(
			"E2E_ADMIN_CLERK_USER_EMAIL must name an authorized existing development admin",
		);
	return email;
};

test.beforeAll(async () => {
	if (!process.env.CLERK_SECRET_KEY?.startsWith("sk_test_"))
		throw new Error("Development Clerk required");
	if (!new URL(process.env.DATABASE_URL ?? "").pathname.endsWith("_e2e"))
		throw new Error("Dedicated _e2e database required");
	await clerkSetup();
	const client = createClerkClient({ secretKey: process.env.CLERK_SECRET_KEY });
	const result = await client.users.getUserList({
		emailAddress: [adminEmail()],
	});
	const admin = result.data[0];
	if (!admin || admin.publicMetadata.role !== "admin")
		throw new Error(
			"Identity is not an existing admin; this suite never changes roles",
		);
	adminId = admin.id;
	const profile = await db().query(
		"INSERT INTO user_profiles (clerk_user_id, onboarding_completed_at) VALUES ($1, now()) ON CONFLICT (clerk_user_id) DO NOTHING RETURNING clerk_user_id",
		[adminId],
	);
	createdProfile = profile.rowCount === 1;
});

test.beforeEach(async ({ page, context }) => {
	test.setTimeout(90_000);
	await context.clearCookies();
	await context.addCookies([
		{ name: "PARAGLIDE_LOCALE", value: "en", domain: "localhost", path: "/" },
	]);
	await page.goto("/");
	await clerk.signIn({ page, emailAddress: adminEmail() });
	page.setDefaultTimeout(15_000);
});

test.afterEach(async ({ page }, info) => {
	if (info.status !== info.expectedStatus) {
		await info.attach("admin-visible-state", {
			body: await page.locator("body").innerText({ timeout: 2000 }),
			contentType: "text/plain",
		});
	}
	await clerk.signOut({ page });
});
test.afterAll(async () => {
	if (createdProfile)
		await db().query("DELETE FROM user_profiles WHERE clerk_user_id=$1", [
			adminId,
		]);
	await closeDb();
});

test("dictionary nullable metadata and deck picker CRUD preserve shared definitions", async ({
	page,
}) => {
	const term = `codex-${randomUUID()}`;
	const deckId = randomUUID();
	const deckName = `Deck ${term}`;
	try {
		await page.goto("/admin/vocabulary");
		await expect(
			page.getByRole("cell", { name: "idempotent", exact: true }),
		).toBeVisible({ timeout: 20_000 });
		await page
			.getByRole("tabpanel")
			.getByRole("button", { name: "Create", exact: true })
			.click();
		let dialog = page.getByRole("dialog");
		await dialog.getByLabel("Term", { exact: true }).fill(term);
		await dialog
			.getByLabel("Definition", { exact: true })
			.fill("A shared definition");
		await dialog.getByLabel("IPA", { exact: true }).fill("/test/");
		await dialog.getByRole("combobox").nth(1).click();
		await page.getByRole("option", { name: /Backend/i }).click();
		await dialog.getByRole("combobox").nth(2).click();
		await page.getByRole("option", { name: "noun", exact: true }).click();
		await dialog.getByRole("button", { name: "Save", exact: true }).click();
		await expect(dialog).not.toBeVisible();
		let row = page.getByRole("row").filter({ hasText: term });
		await expect(row).toBeVisible();
		const item = (
			await db().query(
				"SELECT id, domain, part_of_speech, details FROM vocabulary_items WHERE term=$1",
				[term],
			)
		).rows[0];
		expect(item.domain).toBe("backend_db");
		expect(item.part_of_speech).toBe("noun");
		expect(item.details.ipa).toBe("/test/");
		await row.getByRole("button", { name: "Edit", exact: true }).click();
		dialog = page.getByRole("dialog");
		for (const index of [1, 2]) {
			await dialog.getByRole("combobox").nth(index).click();
			await page.getByRole("option", { name: "—", exact: true }).click();
		}
		await dialog.getByLabel("IPA", { exact: true }).fill("/edited/");
		await dialog.getByRole("button", { name: "Save", exact: true }).click();
		await expect(dialog).not.toBeVisible();
		const updated = (
			await db().query(
				"SELECT domain, part_of_speech, details FROM vocabulary_items WHERE id=$1",
				[item.id],
			)
		).rows[0];
		expect(updated.domain).toBeNull();
		expect(updated.part_of_speech).toBeNull();
		expect(updated.details.ipa).toBe("/edited/");
		await db().query(
			"INSERT INTO vocabulary_decks (id,user_id,is_default,name) VALUES ($1,'',true,$2)",
			[deckId, deckName],
		);
		await page.getByRole("tab", { name: "Decks", exact: true }).click();
		row = page.getByRole("row").filter({ hasText: deckName });
		await row.getByRole("button", { name: "Cards", exact: true }).click();
		dialog = page.getByRole("dialog");
		await dialog.getByLabel("Search dictionary").fill(term);
		await dialog.getByRole("combobox").click();
		await page.getByRole("option", { name: term, exact: true }).click();
		await expect(dialog.getByLabel("Term", { exact: true })).toHaveCount(0);
		await dialog.getByLabel("Order", { exact: true }).fill("3");
		await dialog
			.getByLabel("Example", { exact: true })
			.fill("A card-specific example.");
		await dialog.getByRole("button", { name: "Create", exact: true }).click();
		const card = dialog.getByRole("row").filter({ hasText: term });
		await expect(card).toContainText("A shared definition");
		await card.getByRole("button", { name: "Edit", exact: true }).click();
		await dialog.getByLabel("Note", { exact: true }).fill("My card note");
		await dialog.getByRole("button", { name: "Save", exact: true }).click();
		await expect
			.poll(
				async () =>
					(
						await db().query(
							"SELECT note FROM vocabulary_deck_items WHERE deck_id=$1",
							[deckId],
						)
					).rows[0]?.note,
			)
			.toBe("My card note");
		expect(
			(
				await db().query(
					"SELECT definition FROM vocabulary_items WHERE id=$1",
					[item.id],
				)
			).rows[0].definition,
		).toBe("A shared definition");
		await card.getByRole("button", { name: "Delete", exact: true }).click();
		await page
			.getByRole("alertdialog")
			.getByRole("button", { name: "Delete", exact: true })
			.click();
		await expect(card).toHaveCount(0);
	} finally {
		await db().query("DELETE FROM vocabulary_decks WHERE id=$1", [deckId]);
		await db().query("DELETE FROM vocabulary_items WHERE term=$1", [term]);
	}
});

test("admin attempts labels, checkout IPN and grant/revoke UI stay consistent", async ({
	page,
	browser,
	request,
}) => {
	const learnerId = await clerkUserId();
	const attemptId = randomUUID();
	const examAttemptId = randomUUID();
	const videoAttemptId = randomUUID();
	const categoryId = randomUUID();
	const examId = randomUUID();
	const prefix = `acceptance-${randomUUID()}`;
	const lesson = (await db().query("SELECT id,title FROM lessons LIMIT 1"))
		.rows[0];
	const video = (
		await db().query("SELECT id,title FROM video_exercises LIMIT 1")
	).rows[0];
	const previous = (
		await db().query("SELECT id,status FROM subscriptions WHERE user_id=$1", [
			learnerId,
		])
	).rows;
	await db().query(
		"UPDATE subscriptions SET status='cancelled' WHERE user_id=$1 AND status='active'",
		[learnerId],
	);
	const learner = await browser.newContext();
	try {
		await db().query(
			"INSERT INTO exam_categories (id,slug,title,type) VALUES ($1,$2,'Acceptance exam category','toeic')",
			[categoryId, prefix],
		);
		await db().query(
			"INSERT INTO exam_tests (id,category_id,source_key,slug,title,content) VALUES ($1,$2,$3,$3,'Acceptance nested exam label','{}')",
			[examId, categoryId, prefix],
		);
		await db().query(
			"INSERT INTO attempts (id,user_id,exam_test_id,type,status,mode,selected_parts,started_at,submitted_at) VALUES ($1,$2,$3,'exam','completed','practice',ARRAY[1]::smallint[],now()-interval '1 minute',now())",
			[examAttemptId, learnerId, examId],
		);
		await db().query(
			"INSERT INTO attempts (id,user_id,video_exercise_id,type,status) VALUES ($1,$2,$3,'dictation','completed')",
			[videoAttemptId, learnerId, video.id],
		);
		await db().query(
			"INSERT INTO attempts (id,user_id,lesson_id,type,status,score,duration_sec,started_at,submitted_at) VALUES ($1,$2,$3,'lesson','completed',80,60,now()-interval '1 minute',now())",
			[attemptId, learnerId, lesson.id],
		);
		await page.goto("/admin/users");
		await page
			.getByPlaceholder("Search by name, email or username")
			.fill(process.env.E2E_CLERK_USER_EMAIL ?? "");
		const user = page
			.getByRole("row")
			.filter({ hasText: process.env.E2E_CLERK_USER_EMAIL ?? "" });
		await user
			.getByRole("button", { name: "Learning history", exact: true })
			.click();
		await expect(page.getByRole("dialog")).toContainText(lesson.title);
		await expect(page.getByRole("dialog")).toContainText(video.title);
		await expect(page.getByRole("dialog")).toContainText(
			"Acceptance nested exam label",
		);
		await page
			.getByRole("dialog")
			.getByRole("button", { name: "Close", exact: true })
			.click();
		await user
			.getByRole("button", { name: "Grant premium", exact: true })
			.click();
		await page.getByRole("dialog").getByLabel("Note (optional)").fill(prefix);
		await page
			.getByRole("dialog")
			.getByRole("button", { name: "Save", exact: true })
			.click();
		await expect(user).toContainText("Premium");
		await user
			.getByRole("button", { name: "Grant premium", exact: true })
			.click();
		await page.getByRole("dialog").getByRole("combobox").click();
		await page
			.getByRole("option", { name: "None (remove premium)", exact: true })
			.click();
		await page
			.getByRole("dialog")
			.getByRole("button", { name: "Save", exact: true })
			.click();
		await expect(user).toContainText("Free");
		await learner.addCookies([
			{ name: "PARAGLIDE_LOCALE", value: "en", domain: "localhost", path: "/" },
		]);
		const learnerPage = await learner.newPage();
		await learnerPage.goto(`http://localhost:${process.env.E2E_WEB_PORT}/`);
		await clerk.signIn({
			page: learnerPage,
			emailAddress: process.env.E2E_CLERK_USER_EMAIL ?? "",
		});
		let checkoutURL = "";
		await learnerPage.route(
			"http://127.0.0.1:9/vnpay-test**",
			async (route) => {
				checkoutURL = route.request().url();
				await route.fulfill({
					contentType: "text/html",
					body: "<h1>VNPay test double</h1>",
				});
			},
		);
		await learnerPage.goto(
			`http://localhost:${process.env.E2E_WEB_PORT}/payments/pricing`,
		);
		await learnerPage
			.getByRole("button", { name: "Pay with VNPay", exact: true })
			.first()
			.click();
		await expect(
			learnerPage.getByRole("heading", { name: "VNPay test double" }),
		).toBeVisible();
		const ref = new URL(checkoutURL).searchParams.get("vnp_TxnRef");
		const txn = (
			await db().query(
				"SELECT amount,subscription_id FROM payment_transactions WHERE provider_txn_ref=$1",
				[ref],
			)
		).rows[0];
		const params = new URLSearchParams({
			vnp_Amount: String(txn.amount * 100),
			vnp_ResponseCode: "00",
			vnp_TransactionNo: prefix,
			vnp_TxnRef: ref ?? "",
		});
		params.sort();
		params.set(
			"vnp_SecureHash",
			createHmac("sha512", "e2e-vnpay-secret")
				.update(params.toString())
				.digest("hex"),
		);
		const ipn = await request.get(
			`http://localhost:${process.env.E2E_API_PORT}/api/v1/payments/vnpay/ipn?${params}`,
		);
		expect((await ipn.json()).RspCode).toBe("00");
		await learnerPage.goto(
			`http://localhost:${process.env.E2E_WEB_PORT}/payments/vnpay/return?vnp_ResponseCode=00`,
		);
		await expect(
			learnerPage.getByRole("heading", { name: "Payment successful" }),
		).toBeVisible();
		expect(
			(
				await db().query("SELECT status FROM subscriptions WHERE id=$1", [
					txn.subscription_id,
				])
			).rows[0].status,
		).toBe("active");
		await page.reload();
		await expect(user).toContainText("Premium");
	} finally {
		await learner.close();
		await db().query("DELETE FROM attempts WHERE id=ANY($1::uuid[])", [
			[attemptId, examAttemptId, videoAttemptId],
		]);
		await db().query("DELETE FROM exam_tests WHERE id=$1", [examId]);
		await db().query("DELETE FROM exam_categories WHERE id=$1", [categoryId]);
		await db().query(
			"DELETE FROM payment_transactions WHERE subscription_id IN (SELECT id FROM subscriptions WHERE user_id=$1 AND NOT (id=ANY($2::uuid[])))",
			[learnerId, previous.map((row) => row.id)],
		);
		await db().query(
			"DELETE FROM subscriptions WHERE user_id=$1 AND NOT (id=ANY($2::uuid[]))",
			[learnerId, previous.map((row) => row.id)],
		);
		await db().query(
			"UPDATE subscriptions SET status='active' WHERE id=ANY($1::uuid[])",
			[previous.filter((row) => row.status === "active").map((row) => row.id)],
		);
	}
});
