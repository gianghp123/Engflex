/** Mirrors `enums.UserRole` on the API side (internal/common/enums/user_role.go) — kept here by hand since that enum is excluded from the generated contracts (server-only claim). */
export const ADMIN_ROLE = "admin";

declare global {
	interface CustomJwtSessionClaims {
		role?: string;
	}
}

export function isAdminClaims(
	claims: CustomJwtSessionClaims | null | undefined,
): boolean {
	return claims?.role === ADMIN_ROLE;
}
