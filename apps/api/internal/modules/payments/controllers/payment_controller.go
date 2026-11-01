package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/payments/dtos/requests"
	"engflex-api/internal/modules/payments/dtos/responses"
	"engflex-api/internal/modules/payments/services"
	"engflex-api/internal/server/middleware"
)

// PaymentController exposes the VNPay checkout, IPN, return, plan catalog,
// and the caller's own premium status.
type PaymentController struct {
	plans services.PlanService
	subs  services.SubscriptionService
	vnpay services.VNPayService
}

// NewPaymentController wires the controller to its services.
func NewPaymentController(plans services.PlanService, subs services.SubscriptionService, vnpay services.VNPayService) *PaymentController {
	return &PaymentController{plans: plans, subs: subs, vnpay: vnpay}
}

// Me godoc
// @Summary      Get my current premium status
// @Tags         payments
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=responses.MyPremiumStatus}
// @Router       /premium/me [get]
func (h *PaymentController) Me(c *gin.Context) {
	sub, err := h.subs.GetActive(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	status := responses.MyPremiumStatus{IsPremium: sub != nil}
	if sub != nil {
		status.ExpiresAt = sub.ExpiresAt
	}
	common.OK(c, "premium status", status)
}

// ListPlans godoc
// @Summary      List premium plans
// @Tags         payments
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=[]responses.Plan}
// @Router       /payments/vnpay/plans [get]
func (h *PaymentController) ListPlans(c *gin.Context) {
	items, err := h.plans.ListActive(c.Request.Context())
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "plans listed", items)
}

// Checkout godoc
// @Summary      Start a VNPay checkout for a premium plan
// @Tags         payments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.Checkout true "plan selection"
// @Success      200 {object} common.ApiResponse{data=responses.CheckoutSession}
// @Router       /payments/vnpay/checkout [post]
func (h *PaymentController) Checkout(c *gin.Context) {
	var req requests.Checkout
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	res, err := h.vnpay.BuildCheckoutURL(c.Request.Context(), middleware.UserID(c), req.Plan, c.ClientIP())
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "checkout created", res)
}

// vnpayAckBody is VNPay's own IPN ack shape — not the house
// common.ApiResponse envelope, because VNPay's server parses this reply,
// not a browser.
type vnpayAckBody struct {
	RspCode string `json:"RspCode"`
	Message string `json:"Message"`
}

// IPN godoc
// @Summary      VNPay server-to-server payment notification
// @Tags         payments
// @Produce      json
// @Success      200 {object} vnpayAckBody
// @Router       /payments/vnpay/ipn [get]
func (h *PaymentController) IPN(c *gin.Context) {
	code, err := h.vnpay.VerifyAndHandleIPN(c.Request.Context(), c.Request.URL.Query())
	if err != nil {
		c.JSON(http.StatusOK, vnpayAckBody{RspCode: "99", Message: services.VNPayAckMessage("99")})
		return
	}
	c.JSON(http.StatusOK, vnpayAckBody{RspCode: code, Message: services.VNPayAckMessage(code)})
}

// Return godoc
// @Summary      VNPay browser return — UX status only, never trusted alone
// @Tags         payments
// @Produce      json
// @Success      200 {object} common.ApiResponse
// @Router       /payments/vnpay/return [get]
func (h *PaymentController) Return(c *gin.Context) {
	common.OK(c, "payment processed", gin.H{
		"responseCode": c.Query("vnp_ResponseCode"),
		"txnRef":       c.Query("vnp_TxnRef"),
	})
}
