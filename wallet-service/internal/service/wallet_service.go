package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"wallet-service/internal/model"
	"wallet-service/internal/repository"
)

type WalletService struct {
	repo *repository.WalletRepository
}

func NewWalletService(repo *repository.WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) GetStudentByRfid(rfid string) (*model.StudentWalletAccount, error) {
	return s.repo.GetStudentByRfid(rfid)
}

func (s *WalletService) GetStudentByIdOrEmail(query string) (*model.StudentWalletAccount, error) {
	return s.repo.GetStudentByIdOrEmail(query)
}

func (s *WalletService) GetRechargeHistory(studentID string) ([]model.RechargeRecord, error) {
	return s.repo.GetRechargeHistory(studentID)
}

func (s *WalletService) GetRechargeSummaryToday() (map[string]interface{}, error) {
	count, amount, err := s.repo.GetRechargeSummaryToday()
	if err != nil {
		return nil, err
	}

	nfcCount, err := s.repo.GetNfcRechargesCount()
	if err != nil {
		return nil, err
	}

	manualCount, err := s.repo.GetManualRechargesCount()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"todaysTotalCount":       count,
		"todaysTotalAmount":      amount,
		"nfcRechargesCount":      nfcCount,
		"manualRechargesCount":   manualCount,
	}, nil
}

func (s *WalletService) ProcessNfcRecharge(req model.NfcRechargeRequest) (*model.StudentWalletAccount, *model.RechargeRecord, error) {
	student, err := s.repo.GetStudentByRfid(req.RFIDCardID)
	if err != nil {
		return nil, nil, fmt.Errorf("NFC Card \"%s\" not recognized in system", req.RFIDCardID)
	}

	prevBalance := student.Balance
	newBalance := prevBalance + req.Amount

	err = s.repo.UpdateStudentBalance(student.StudentID, newBalance)
	if err != nil {
		return nil, nil, err
	}

	// Refetch student state
	student.Balance = newBalance

	rechargeID := fmt.Sprintf("RCH-%d", 1000+rand.Intn(9000))
	rechargedBy := req.RechargedBy
	if rechargedBy == "" {
		rechargedBy = "Manager"
	}

	record := model.RechargeRecord{
		ID:              rechargeID,
		StudentID:       student.StudentID,
		StudentName:     student.Name,
		StudentEmail:    student.Email,
		RFIDCardID:      student.RFIDCardID,
		Amount:          req.Amount,
		Method:          "nfc",
		PaymentType:     req.PaymentType,
		PreviousBalance: prevBalance,
		NewBalance:      newBalance,
		Timestamp:       time.Now(),
		RechargedBy:     rechargedBy,
		Status:          "success",
	}

	err = s.repo.CreateRechargeRecord(&record)
	if err != nil {
		return nil, nil, err
	}

	return student, &record, nil
}

func (s *WalletService) ProcessManualRecharge(req model.ManualRechargeRequest) (*model.StudentWalletAccount, *model.RechargeRecord, error) {
	student, err := s.repo.GetStudentByIdOrEmail(req.StudentID)
	
	var prevBalance float64
	var isNew bool

	if err != nil {
		// Create new student record if not found
		isNew = true
		prevBalance = 0.0

		// Generate random RFIDCardId
		rand.Seed(time.Now().UnixNano())
		rfidCardId := fmt.Sprintf("RFID-%03d-%s", 100+rand.Intn(900), strings.ToUpper(req.StudentID[len(req.StudentID)-2:]))

		student = &model.StudentWalletAccount{
			StudentID:  strings.ToUpper(req.StudentID),
			Name:       req.Name,
			Email:      strings.ToLower(req.Email),
			RFIDCardID: rfidCardId,
			Department: "General Campus",
			Avatar:     fmt.Sprintf("https://api.dicebear.com/9.x/avataaars/svg?seed=%s", urlEncode(req.Name)),
			Balance:    req.Amount,
			Phone:      "+91 98000 00000",
		}

		err = s.repo.CreateStudentWallet(student)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to register student wallet: %v", err)
		}
	} else {
		// Update existing student
		prevBalance = student.Balance
		newBalance := prevBalance + req.Amount

		err = s.repo.UpdateStudentBalance(student.StudentID, newBalance)
		if err != nil {
			return nil, nil, err
		}
		student.Balance = newBalance
	}

	rechargeID := fmt.Sprintf("RCH-%d", 1000+rand.Intn(9000))
	rechargedBy := req.RechargedBy
	if rechargedBy == "" {
		rechargedBy = "Manager"
	}

	record := model.RechargeRecord{
		ID:              rechargeID,
		StudentID:       student.StudentID,
		StudentName:     student.Name,
		StudentEmail:    student.Email,
		RFIDCardID:      student.RFIDCardID,
		Amount:          req.Amount,
		Method:          "manual",
		PaymentType:     req.PaymentType,
		PreviousBalance: prevBalance,
		NewBalance:      student.Balance,
		Timestamp:       time.Now(),
		RechargedBy:     rechargedBy,
		Status:          "success",
	}

	err = s.repo.CreateRechargeRecord(&record)
	if err != nil {
		return nil, nil, err
	}

	_ = isNew // for tracking purposes
	return student, &record, nil
}

func (s *WalletService) DeductBalance(req model.DeductBalanceRequest) error {
	student, err := s.repo.GetStudentByEmail(req.Email)
	if err != nil {
		return errors.New("student wallet account not found")
	}

	if student.Balance < req.Amount {
		return fmt.Errorf("Insufficient RFID wallet balance! Current: ₹%.2f", student.Balance)
	}

	newBalance := student.Balance - req.Amount
	return s.repo.UpdateStudentBalance(student.StudentID, newBalance)
}

func urlEncode(str string) string {
	return strings.ReplaceAll(str, " ", "+")
}
