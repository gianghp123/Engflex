import { randomUUID } from "node:crypto";
import { expect, test } from "@playwright/test";
import { apiToken, clerkUserId, closeDb, db } from "./helpers";

test.afterAll(closeDb);

for (const [locale, title] of [
	["en", "Page not found"],
	["vi", "Không tìm thấy trang"],
] as const) {
	test(`SSR and browser render the ${locale} locale on signed-out errors`, async ({
		request,
		page,
		context,
	}) => {
		const response = await request.get("/modular-missing-route", {
			headers: { Cookie: `PARAGLIDE_LOCALE=${locale}` },
		});
		expect(response.status()).toBe(404);
		const html = await response.text();
		expect(html).toContain(`<html lang="${locale}"`);
		expect(html).toContain(title);
		await context.clearCookies();
		await context.addCookies([
			{
				name: "PARAGLIDE_LOCALE",
				value: locale,
				domain: "localhost",
				path: "/",
			},
		]);
		await page.goto("/modular-missing-route");
		await expect(
			page.getByRole("heading", { name: title, exact: true }),
		).toBeVisible();
		expect(await page.locator("html").getAttribute("lang")).toBe(locale);
	});
}

test("direct child/action routes enforce publication and current entitlement", async ({
	request,
}) => {
	const userId = await clerkUserId();
	const lesson = randomUUID();
	const video = randomUUID();
	const activity = randomUUID();
	const subscription = randomUUID();
	const previous = await db().query(
		"UPDATE subscriptions SET status = 'cancelled' WHERE user_id = $1 AND status = 'active' RETURNING id",
		[userId],
	);
	try {
		for (const [table, id] of [
			["lessons", lesson],
			["video_exercises", video],
		] as const) {
			await db().query(
				`INSERT INTO ${table} SELECT (jsonb_populate_record(NULL::${table}, to_jsonb(source) || $1::jsonb)).* FROM ${table} source LIMIT 1`,
				[
					JSON.stringify({
						id,
						slug: `policy-${id}`,
						title: "Policy test content",
						status: "published",
						is_premium: true,
					}),
				],
			);
		}
		await db().query(
			`INSERT INTO lesson_activities (id, lesson_id, part_number, type, title) VALUES ($1, $2, 1, 'reading', 'Protected activity')`,
			[activity, lesson],
		);
		await db().query(
			`INSERT INTO subscriptions (id, user_id, plan, status, amount, expires_at) VALUES ($1, $2, '1m', 'active', 1, now() + interval '1 day')`,
			[subscription, userId],
		);
		const headers = { Authorization: `Bearer ${await apiToken()}` };
		const attemptId = randomUUID();
		const routes = [
			{ method: "GET", path: `/lessons/${lesson}` },
			{ method: "GET", path: `/lessons/${lesson}/attempts/open` },
			{ method: "POST", path: `/lessons/${lesson}/attempts`, data: {} },
			{ method: "GET", path: `/lesson-activities/${activity}` },
			{ method: "GET", path: `/lesson-activities?lessonId=${lesson}` },
			{ method: "GET", path: `/video-exercises/${video}` },
			{ method: "GET", path: `/video-transcripts/video-exercise/${video}` },
			{
				method: "GET",
				path: `/video-exercises/${video}/attempts/open?type=dictation`,
			},
			{
				method: "POST",
				path: `/video-exercises/${video}/attempts`,
				data: { type: "dictation" },
			},
			{
				method: "PUT",
				path: `/video-exercises/${video}/attempts/${attemptId}/progress`,
				data: { revision: 1 },
			},
			{
				method: "POST",
				path: `/video-exercises/${video}/attempts/${attemptId}/complete`,
				data: {},
			},
		];
		const base = `http://localhost:${process.env.E2E_API_PORT}/api/v1`;
		for (const state of ["expired", "revoked", "draft"] as const) {
			await db().query(
				"UPDATE subscriptions SET status=$2, expires_at=now() + ($3 * interval '1 day') WHERE id=$1",
				[
					subscription,
					state === "revoked" ? "cancelled" : "active",
					state === "expired" ? -1 : 1,
				],
			);
			for (const [table, id] of [
				["lessons", lesson],
				["video_exercises", video],
			] as const) {
				await db().query(`UPDATE ${table} SET status=$2 WHERE id=$1`, [
					id,
					state === "draft" ? "draft" : "published",
				]);
			}
			for (const route of routes) {
				const response = await request.fetch(base + route.path, {
					method: route.method,
					data: route.data,
					headers,
				});
				expect(
					response.status(),
					`${state}: ${route.method} ${route.path}`,
				).toBe(state === "draft" ? 404 : 403);
			}
		}
		for (const [table, id] of [
			["lessons", lesson],
			["video_exercises", video],
		] as const) {
			await db().query(`UPDATE ${table} SET status='published' WHERE id=$1`, [
				id,
			]);
			const path = table === "lessons" ? "lessons" : "video-exercises";
			expect(
				(await request.get(`${base}/${path}/${id}`, { headers })).status(),
			).toBe(200);
		}
		const attempts = await db().query(
			"SELECT id FROM attempts WHERE lesson_id=$1 OR video_exercise_id=$2",
			[lesson, video],
		);
		expect(attempts.rowCount).toBe(0);
	} finally {
		await db().query("DELETE FROM subscriptions WHERE id=$1", [subscription]);
		await db().query("DELETE FROM lessons WHERE id=$1", [lesson]);
		await db().query("DELETE FROM video_exercises WHERE id=$1", [video]);
		await db().query(
			"UPDATE subscriptions SET status='active' WHERE id=ANY($1::uuid[])",
			[previous.rows.map((row) => row.id)],
		);
	}
});
