import type {
	SepayCheckoutSession,
	SubscriptionPlan,
} from "@engflex/contracts";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { CheckCircle2, Loader2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import {
	Card,
	CardContent,
	CardFooter,
	CardHeader,
	CardTitle,
} from "#/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import { entitlementKeys } from "#/features/payments/entitlements";
import {
	usePlans,
	useSepayStatus,
	useStartSepayCheckout,
} from "#/features/payments/queries";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/payments/pricing")({
	component: PricingPage,
});

function planLabel(plan: SubscriptionPlan): string {
	switch (plan) {
		case "1m":
			return m["payments.pricing.plan.1m"]();
		case "3m":
			return m["payments.pricing.plan.3m"]();
		case "12m":
			return m["payments.pricing.plan.12m"]();
	}
}

function formatVnd(amount: number): string {
	return new Intl.NumberFormat("vi-VN", {
		style: "currency",
		currency: "VND",
	}).format(amount);
}

function SepayQrDialog({
	session,
	onOpenChange,
}: {
	session: SepayCheckoutSession;
	onOpenChange: (open: boolean) => void;
}) {
	const navigate = useNavigate();
	const queryClient = useQueryClient();
	const statusQuery = useSepayStatus(session.reference, true);
	const status = statusQuery.data?.status ?? "pending";

	return (
		<Dialog open onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-md">
				<DialogHeader>
					<DialogTitle>{m["payments.pricing.sepay.title"]()}</DialogTitle>
					<DialogDescription>
						{status === "succeeded"
							? m["payments.pricing.sepay.succeeded"]()
							: m["payments.pricing.sepay.body"]()}
					</DialogDescription>
				</DialogHeader>
				{status === "succeeded" ? (
					<div className="flex flex-col items-center gap-3 py-6">
						<CheckCircle2 className="size-16 text-primary" />
						<Button
							type="button"
							onClick={() => {
								// The cached premium status still says "not premium" —
								// everything gated on it (the sidebar's upsell entry,
								// every locked card) would keep showing the pre-payment
								// state until that query happened to refetch.
								void queryClient.invalidateQueries({
									queryKey: entitlementKeys.me,
								});
								onOpenChange(false);
								void navigate({ to: APP_ROUTES.HOME });
							}}
						>
							{m["payments.pricing.sepay.done"]()}
						</Button>
					</div>
				) : (
					<div className="flex flex-col items-center gap-3">
						<img
							src={session.qrUrl}
							alt={m["payments.pricing.sepay.title"]()}
							className="w-64 rounded-lg border"
						/>
						<div className="flex w-full flex-col gap-1 text-sm">
							<div className="flex justify-between">
								<span className="text-muted-foreground">
									{m["payments.pricing.sepay.bank"]()}
								</span>
								<span className="font-semibold">{session.bankShortName}</span>
							</div>
							<div className="flex justify-between">
								<span className="text-muted-foreground">
									{m["payments.pricing.sepay.account"]()}
								</span>
								<span className="font-semibold">
									{session.bankAccountNumber}
								</span>
							</div>
							<div className="flex justify-between">
								<span className="text-muted-foreground">
									{m["payments.pricing.sepay.holder"]()}
								</span>
								<span className="font-semibold">
									{session.accountHolderName}
								</span>
							</div>
							<div className="flex justify-between">
								<span className="text-muted-foreground">
									{m["payments.pricing.sepay.amount"]()}
								</span>
								<span className="font-semibold">
									{formatVnd(session.amount)}
								</span>
							</div>
							<div className="flex justify-between">
								<span className="text-muted-foreground">
									{m["payments.pricing.sepay.content"]()}
								</span>
								<span className="font-mono font-semibold">
									{session.reference}
								</span>
							</div>
						</div>
						<span className="inline-flex items-center gap-2 text-xs text-muted-foreground">
							<Loader2 className="size-3.5 animate-spin" />
							{m["payments.pricing.sepay.waiting"]()}
						</span>
					</div>
				)}
			</DialogContent>
		</Dialog>
	);
}

function PricingPage() {
	useBreadcrumbs([{ label: m["payments.pricing.crumb"]() }]);
	const { data: plans, isPending, isError } = usePlans();
	const checkout = useStartSepayCheckout();
	const [pendingPlan, setPendingPlan] = useState<SubscriptionPlan | null>(null);
	const [session, setSession] = useState<SepayCheckoutSession | null>(null);

	function handleCheckout(plan: SubscriptionPlan) {
		setPendingPlan(plan);
		checkout.mutate(
			{ plan },
			{
				onSuccess: (newSession) => {
					setPendingPlan(null);
					setSession(newSession);
				},
				onError: (error) => {
					setPendingPlan(null);
					toast.error(
						error instanceof Error
							? error.message
							: m["payments.pricing.checkoutError"](),
					);
				},
			},
		);
	}

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="heart" size={84} />,
				title: m["payments.pricing.title"](),
				description: m["payments.pricing.subtitle"](),
			}}
		>
			{isPending && (
				<p className="text-muted-foreground">
					{m["payments.pricing.loading"]()}
				</p>
			)}
			{isError && (
				<p className="text-destructive">{m["payments.pricing.error"]()}</p>
			)}
			{plans && (
				<div className="grid gap-4 md:grid-cols-3">
					{plans.map((plan) => {
						const perMonth = Math.round(plan.amount / plan.months);
						const isSubmitting =
							checkout.isPending && pendingPlan === plan.plan;
						return (
							<Card key={plan.plan}>
								<CardHeader>
									<CardTitle className="text-xl">
										{planLabel(plan.plan)}
									</CardTitle>
								</CardHeader>
								<CardContent className="flex flex-col gap-1">
									<span className="text-3xl font-bold">
										{formatVnd(plan.amount)}
									</span>
									<span className="text-sm text-muted-foreground">
										{m["payments.pricing.perMonth"]({
											amount: formatVnd(perMonth),
										})}
									</span>
								</CardContent>
								<CardFooter>
									<Button
										type="button"
										className="w-full"
										disabled={checkout.isPending}
										onClick={() => handleCheckout(plan.plan)}
									>
										{isSubmitting ? (
											<Loader2 className="size-4 animate-spin" />
										) : (
											m["payments.pricing.cta"]()
										)}
									</Button>
								</CardFooter>
							</Card>
						);
					})}
				</div>
			)}
			{session ? (
				<SepayQrDialog
					session={session}
					onOpenChange={(open) => {
						if (!open) setSession(null);
					}}
				/>
			) : null}
		</PageLayout>
	);
}
