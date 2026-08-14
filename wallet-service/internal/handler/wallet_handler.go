package handler

import (
	"net/http"
	"wallet-service/internal/model"
	"wallet-service/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	srv *service.WalletService
}

func NewWalletHandler(srv *service.WalletService) *WalletHandler {
	return &WalletHandler{srv: srv}
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
