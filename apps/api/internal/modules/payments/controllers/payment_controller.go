package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/payments/dtos/requests"
	// Blank: responses appears only in swagger annotations below (ListPlans,
	// Checkout), and swag resolves annotation types through imports.
	_ "engflex-api/internal/modules/payments/dtos/responses"
	"engflex-api/internal/modules/payments/services"
	"engflex-api/internal/server/middleware"
)

// PaymentController exposes the VNPay checkout/IPN/return routes, the
// SePay checkout/webhook/status routes, the plan catalog, and the
// caller's own premium status.
type PaymentController struct {
	plans services.PlanService
	subs  services.SubscriptionService
	vnpay services.VNPayService
	sepay services.SepayService
}

// NewPaymentController wires the controller to its services.
func NewPaymentController(plans services.PlanService, subs services.SubscriptionService, vnpay services.VNPayService, sepay services.SepayService) *PaymentController {
	return &PaymentController{plans: plans, subs: subs, vnpay: vnpay, sepay: sepay}
}

// Me was superseded by the access module's GET /premium/me
// (MyEntitlements). The MyPremiumStatus DTO stays for the web client's
// current premium badge until the frontend migrates (out of scope here).

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

// SepayCheckout godoc
// @Summary      Start a SePay checkout for a premium plan
// @Description  Returns a VietQR image URL to render inline — no redirect.
// @Tags         payments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.Checkout true "plan selection"
// @Success      200 {object} common.ApiResponse{data=responses.SepayCheckoutSession}
// @Router       /payments/sepay/checkout [post]
func (h *PaymentController) SepayCheckout(c *gin.Context) {
	var req requests.Checkout
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	res, err := h.sepay.BuildCheckoutQR(c.Request.Context(), middleware.UserID(c), req.Plan)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "checkout created", res)
}

// SepayWebhook godoc
// @Summary      SePay server-to-server payment notification
// @Tags         payments
// @Accept       json
// @Produce      json
// @Success      200 {object} object{success=bool}
// @Router       /payments/sepay/webhook [post]
func (h *PaymentController) SepayWebhook(c *gin.Context) {
	if !h.sepay.VerifyWebhookAuth(c.GetHeader("Authorization")) {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false})
		return
	}
	var payload services.SepayWebhook
	if err := c.ShouldBindJSON(&payload); err != nil {
		// Still 200+success: a malformed body is SePay's problem to fix,
		// not something retrying will resolve, and a rejection here just
		// triggers endless retries for a payload that will never parse.
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}
	if err := h.sepay.HandleWebhook(c.Request.Context(), payload); err != nil {
		common.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// SepayStatus godoc
// @Summary      Poll a SePay checkout's payment status
// @Tags         payments
// @Security     BearerAuth
// @Produce      json
// @Param        reference path string true "checkout reference"
// @Success      200 {object} common.ApiResponse{data=responses.SepayStatus}
// @Router       /payments/sepay/status/{reference} [get]
func (h *PaymentController) SepayStatus(c *gin.Context) {
	res, err := h.sepay.GetStatus(c.Request.Context(), c.Param("reference"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "status", res)
}
