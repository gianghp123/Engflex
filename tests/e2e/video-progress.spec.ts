import { randomUUID } from "node:crypto";
import { expect, test } from "./authenticated-test";
import { apiToken, closeDb, db } from "./helpers";

test.afterAll(closeDb);

test("video conflict requires refetch and premium denial preserves the draft", async ({
	page,
	context,
	request,
}) => {
	await context.addCookies([
		{ name: "PARAGLIDE_LOCALE", value: "en", domain: "localhost", path: "/" },
	]);
	const videoId = randomUUID();
	const transcriptId = randomUUID();
	const api = `http://localhost:${process.env.E2E_API_PORT}/api/v1/video-exercises/${videoId}`;
	const headers = { Authorization: `Bearer ${await apiToken()}` };
	try {
		await db().query(
			`INSERT INTO video_exercises (id, title, video_url, status) VALUES ($1, 'Revision regression', 'http://127.0.0.1:9/media', 'published')`,
			[videoId],
		);
		await db().query(
			`INSERT INTO video_transcripts (id, video_exercise_id, sequence, content, start_timestamp, end_timestamp) VALUES ($1, $2, 0, 'Hello world', 0, 2)`,
			[transcriptId, videoId],
		);
		const started = await request.post(`${api}/attempts`, {
			headers,
			data: { type: "dictation" },
		});
		expect(started.status()).toBe(201);
		const { data: attempt } = await started.json();
		await page.goto(`/videos/${videoId}/dictation`);
		await expect(
			page.getByRole("radio", { name: "Normal", exact: true }),
		).toBeChecked();
		const otherWriter = await request.put(
			`${api}/attempts/${attempt.id}/progress`,
			{ headers, data: { revision: attempt.revision, level: "hard" } },
		);
		expect(otherWriter.status()).toBe(200);
		await page.getByRole("radio", { name: "Easy", exact: true }).click();
		await expect(page.getByRole("alert")).toContainText(
			"Progress changed in another session",
		);
		await page
			.getByRole("button", { name: "Load saved progress (replace this draft)" })
			.click();
		await expect(
			page.getByRole("radio", { name: "Hard", exact: true }),
		).toBeChecked();
		await page.getByRole("radio", { name: "Normal", exact: true }).click();
		await expect
			.poll(
				async () =>
					(
						await db().query("SELECT revision FROM attempts WHERE id=$1", [
							attempt.id,
						])
					).rows[0].revision,
			)
			.toBe(3);
		await page.getByRole("textbox").fill("my draft");
		await db().query("UPDATE video_exercises SET is_premium=true WHERE id=$1", [
			videoId,
		]);
		await page.getByRole("radio", { name: "Easy", exact: true }).click();
		await expect(page.getByRole("alert")).toContainText(
			"Access is no longer available",
		);
		await expect(page.getByRole("textbox")).toHaveValue("my draft");
		expect(
			(
				await db().query("SELECT revision FROM attempts WHERE id=$1", [
					attempt.id,
				])
			).rows[0].revision,
		).toBe(3);
		await db().query(
			"UPDATE video_exercises SET is_premium=false WHERE id=$1",
			[videoId],
		);
		await page.getByRole("button", { name: "Retry with this draft" }).click();
		await expect
			.poll(
				async () =>
					(
						await db().query("SELECT revision FROM attempts WHERE id=$1", [
							attempt.id,
						])
					).rows[0].revision,
			)
			.toBe(4);
		await expect(page.getByRole("textbox")).toHaveValue("my draft");

		// Commit the completion on the real backend, then lose only its RPC
		// response. Retrying must replay completion rather than save a closed run.
		let droppedCompletion = false;
		await page.route("**/_serverFn/**", async (route) => {
			const payload = route.request().postData() ?? "";
			if (
				!droppedCompletion &&
				payload.includes(videoId) &&
				payload.includes("/complete")
			) {
				droppedCompletion = true;
				await route.fetch();
				await route.abort("failed");
			} else {
				await route.continue();
			}
		});
		await page.getByRole("textbox").fill("Hello");
		await page.getByRole("textbox").press("Space");
		await page.getByRole("textbox").fill("world");
		await page.getByRole("textbox").press("Space");
		await page.getByRole("button", { name: "Next", exact: true }).click();
		await expect(page.getByRole("alert")).toContainText(
			"Progress could not be saved",
		);
		expect(droppedCompletion).toBe(true);
		const committed = (
			await db().query("SELECT status, revision FROM attempts WHERE id=$1", [
				attempt.id,
			])
		).rows[0];
		expect(committed.status).toBe("completed");
		await page.getByRole("button", { name: "Retry with this draft" }).click();
		await expect(page.getByRole("dialog")).toBeVisible();
		expect(
			(
				await db().query("SELECT revision FROM attempts WHERE id=$1", [
					attempt.id,
				])
			).rows[0].revision,
		).toBe(committed.revision);
	} finally {
		await db().query("DELETE FROM attempts WHERE video_exercise_id=$1", [
			videoId,
		]);
		await db().query("DELETE FROM video_exercises WHERE id=$1", [videoId]);
	}
});
