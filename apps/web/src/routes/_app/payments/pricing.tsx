import type { SubscriptionPlan } from "@engflex/contracts";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
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
import { usePlans, useStartCheckout } from "#/features/payments/queries";
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

function PricingPage() {
	useBreadcrumbs([{ label: m["payments.pricing.crumb"]() }]);
	const { data: plans, isPending, isError } = usePlans();
	const checkout = useStartCheckout();
	const [pendingPlan, setPendingPlan] = useState<SubscriptionPlan | null>(null);

	function handleCheckout(plan: SubscriptionPlan) {
		setPendingPlan(plan);
		checkout.mutate(
			{ plan },
			{
				onSuccess: (session) => {
					window.location.href = session.paymentUrl;
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
										{isSubmitting
											? m["payments.pricing.redirecting"]()
											: m["payments.pricing.cta"]()}
									</Button>
								</CardFooter>
							</Card>
						);
					})}
				</div>
			)}
		</PageLayout>
	);
}
