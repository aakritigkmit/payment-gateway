package services

import (
	"context"
	"fmt"

	"github.com/aakritigkmit/payment-gateway/internal/dto"
	"github.com/aakritigkmit/payment-gateway/internal/helpers"
	"github.com/aakritigkmit/payment-gateway/internal/model"
	"github.com/aakritigkmit/payment-gateway/internal/repository"
	"github.com/aakritigkmit/payment-gateway/internal/utils"
)

type OrderService struct {
	repo            *repository.OrderRepo
	transactionRepo *repository.TransactionRepo
}

func NewOrderService(repo *repository.OrderRepo, transactionRepo *repository.TransactionRepo) *OrderService {
	return &OrderService{repo, transactionRepo}
}

func (s *OrderService) PlaceOrder(ctx context.Context, req dto.PlaceOrderRequest) (utils.OrderAPIResponse, error) {

	tokenResp, err := utils.FetchAccessToken(ctx)
	if err != nil {
		return utils.OrderAPIResponse{}, err
	}
	jsonPayload, err := helpers.BuildOrderPayload(req)
	if err != nil {
		return utils.OrderAPIResponse{}, err
	}

	orderResp, err := utils.CreateOrderRequest(ctx, tokenResp.AccessToken, jsonPayload)
	if err != nil {
		return utils.OrderAPIResponse{}, err
	}

	transaction := model.Transaction{
		MerchantOrderReference: req.MerchantOrderReference,
		OrderAmount: model.OrderAmount{
			Value:    req.OrderAmount.Value,
			Currency: req.OrderAmount.Currency,
		},
		PreAuth:               req.PreAuth,
		AllowedPaymentMethods: req.AllowedPaymentMethods,
		Notes:                 req.Notes,
		CallbackURL:           req.CallbackURL,
		FailureCallbackURL:    req.FailureCallbackURL,
		PurchaseDetails: model.PurchaseDetails{
			MerchantMetadata: req.PurchaseDetails.MerchantMetadata,
			Customer: model.Customer{
				EmailID:         req.PurchaseDetails.Customer.EmailID,
				FirstName:       req.PurchaseDetails.Customer.FirstName,
				LastName:        req.PurchaseDetails.Customer.LastName,
				CustomerID:      req.PurchaseDetails.Customer.CustomerID,
				MobileNumber:    req.PurchaseDetails.Customer.MobileNumber,
				BillingAddress:  model.Address(req.PurchaseDetails.Customer.BillingAddress),
				ShippingAddress: model.Address(req.PurchaseDetails.Customer.ShippingAddress),
			},
		},
	}

	if err := s.transactionRepo.SaveTransaction(ctx, transaction); err != nil {
		return utils.OrderAPIResponse{}, err
	}

	order := model.Order{
		UserID:                 req.PurchaseDetails.Customer.CustomerID,
		TransactionReferenceId: orderResp.OrderID,
		Amount:                 req.OrderAmount.Value,
		Currency:               req.OrderAmount.Currency,
		Status:                 "Pending",
	}

	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return utils.OrderAPIResponse{}, err
	}

	return utils.OrderAPIResponse{
		Token:       orderResp.Token,
		OrderID:     orderResp.OrderID,
		RedirectURL: orderResp.RedirectURL,
	}, nil
}

func (s *OrderService) UpdateOrder(referenceID string, payload *dto.UpdateOrderPayload) error {
	if referenceID == "" {
		return fmt.Errorf("transaction reference ID is required")
	}

	if payload.Status != "" && payload.Status != "success" && payload.Status != "failed" && payload.Status != "pending" {
		return fmt.Errorf("invalid status value")
	}

	return s.repo.UpdateOrder(referenceID, payload)
}
