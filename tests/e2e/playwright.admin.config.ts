import { defineConfig, devices } from "@playwright/test";
import base from "./playwright.config";

// Explicitly invoked acceptance suite: existing development admin only.
// No role mutations and no real payment/inference providers.
const servers = Array.isArray(base.webServer) ? base.webServer : [];
export default defineConfig({
	...base,
	projects: [
		{ name: "global setup", testMatch: /global\.setup\.ts/ },
		{
			name: "admin acceptance",
			testMatch: /admin-contracts\.spec\.ts/,
			dependencies: ["global setup"],
			use: { ...devices["Desktop Chrome"] },
		},
	],
	webServer: servers.map((server, index) =>
		index === 0
			? {
					...server,
					env: {
						...server.env,
						VNPAY_PAY_URL: "http://127.0.0.1:9/vnpay-test",
						VNPAY_TMN_CODE: "e2e",
						VNPAY_HASH_SECRET: "e2e-vnpay-secret",
						VNPAY_RETURN_URL: `http://localhost:${process.env.E2E_WEB_PORT}/payments/vnpay/return`,
					},
				}
			: server,
	),
});
