package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"wallet-service/internal/model"
	"wallet-service/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	srv           *service.WalletService
	webhookSecret string
}

func NewWalletHandler(srv *service.WalletService) *WalletHandler {
	secret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("RAZORPAY_WEBHOOK_SECRET must be set")
	}
	return &WalletHandler{srv: srv, webhookSecret: secret}
}

// GetBalance returns the float64 balance for a given userId (student ID) or email
func (h *WalletHandler) GetBalance(c *gin.Context) {
	userId := c.Query("userId")
	email := c.Query("email")
	rfid := c.Query("rfid")

	var student *model.StudentWalletAccount
	var err error

	if userId != "" {
		student, err = h.srv.GetStudentByIdOrEmail(userId)
	} else if email != "" {
		student, err = h.srv.GetStudentByIdOrEmail(email)
	} else if rfid != "" {
		student, err = h.srv.GetStudentByRfid(rfid)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one of 'userId', 'email', or 'rfid' is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet account not found", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"studentId": student.StudentID,
		"email":     student.Email,
		"balance":   student.Balance,
	})
}

// GetStudent queries a student profile by search string (query) or rfid
func (h *WalletHandler) GetStudent(c *gin.Context) {
	rfid := c.Query("rfid")
	query := c.Query("query")

	var student *model.StudentWalletAccount
	var err error

	if rfid != "" {
		student, err = h.srv.GetStudentByRfid(rfid)
	} else if query != "" {
		student, err = h.srv.GetStudentByIdOrEmail(query)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Either 'rfid' or 'query' is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student wallet profile not found"})
		return
	}

	c.JSON(http.StatusOK, student)
}

// NfcRecharge handles card swipe top-ups
func (h *WalletHandler) NfcRecharge(c *gin.Context) {
	var req model.NfcRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student, record, err := h.srv.ProcessNfcRecharge(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "NFC Wallet Recharge successful!",
		"record":         record,
		"updatedStudent": student,
	})
}

// ManualRecharge handles cashier desk manual top-ups
func (h *WalletHandler) ManualRecharge(c *gin.Context) {
	var req model.ManualRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student, record, err := h.srv.ProcessManualRecharge(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "Manual Wallet Recharge successful!",
		"record":         record,
		"updatedStudent": student,
	})
}

// DeductBalance handles digital wallet deductions during checkout
func (h *WalletHandler) DeductBalance(c *gin.Context) {
	var req model.DeductBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.srv.DeductBalance(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Payment successful!",
	})
}

// GetHistory returns list of recharge transactions
func (h *WalletHandler) GetHistory(c *gin.Context) {
	studentId := c.Query("studentId")
	history, err := h.srv.GetRechargeHistory(studentId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recharge history"})
		return
	}
	c.JSON(http.StatusOK, history)
}

// GetSummary returns dashboard metrics cards summary stats
func (h *WalletHandler) GetSummary(c *gin.Context) {
	summary, err := h.srv.GetRechargeSummaryToday()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load wallet dashboard stats"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// RazorpayWebhook handles incoming payment capture notifications from Razorpay
func (h *WalletHandler) RazorpayWebhook(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// 1. Read signature header
	sigHeader := c.GetHeader("X-Razorpay-Signature")

	// 2. Validate signature using stored webhookSecret
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write(bodyBytes)
	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedSig), []byte(sigHeader)) {
		log.Printf("[Error] Webhook signature verification failed. Expected: %s, Received: %s", expectedSig, sigHeader)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// 3. Parse JSON payload
	var payload model.RazorpayWebhookPayload
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	// 4. Process only payment.captured event
	if payload.Event == "payment.captured" {
		payment := payload.Payload.Payment.Entity
		studentID := payment.Notes.StudentID
		if studentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Student ID missing from payment notes"})
			return
		}

		// Amount is in paise, convert to INR
		amountINR := payment.Amount / 100.0

		// Update wallet balance via service layer
		student, record, err := h.srv.ProcessOnlineRecharge(studentID, amountINR, payment.ID, payment.Method)
		if err != nil {
			log.Printf("[Error] Failed to process online recharge: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Printf("[Success] Online recharge processed for Student: %s, Amount: ₹%.2f, PaymentID: %s", studentID, amountINR, payment.ID)
		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"message":        "Payment captured and wallet credited successfully",
			"paymentId":      payment.ID,
			"studentId":      studentID,
			"amount":         amountINR,
			"record":         record,
			"updatedStudent": student,
		})
		return
	}

	// Handle other events gracefully (ignore them but return 200 OK)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook received but not processed (ignored event: " + payload.Event + ")",
	})
}

// OnlineRecharge handles direct Razorpay/online payment captures from the authenticated client
func (h *WalletHandler) OnlineRecharge(c *gin.Context) {
	var req model.OnlineRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	student, record, err := h.srv.ProcessOnlineRecharge(req.Email, req.Amount, req.PaymentID, req.PaymentMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "Wallet recharged successfully via Razorpay!",
		"record":         record,
		"updatedStudent": student,
	})
}
