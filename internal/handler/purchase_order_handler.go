package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"be-integration-example/internal/model"
	"be-integration-example/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PurchaseOrderHandler struct {
	poRepo      repository.PurchaseOrderRepository
	productRepo repository.ProductRepository
}

func NewPurchaseOrderHandler(poRepo repository.PurchaseOrderRepository, productRepo repository.ProductRepository) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		poRepo:      poRepo,
		productRepo: productRepo,
	}
}

func (h *PurchaseOrderHandler) RegisterRoutes(router fiber.Router) {
	po := router.Group("/purchase-orders")
	po.Post("/", h.Create)
	po.Get("/", h.List)
	po.Get("/:id", h.GetByID)
	po.Patch("/:id/status", h.UpdateStatus)
	po.Delete("/:id", h.Delete)
}

func (h *PurchaseOrderHandler) Create(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "unauthorized",
			"message": err.Error(),
		})
	}

	var req model.CreatePurchaseOrderRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "bad_request",
			"message": "invalid request body",
		})
	}

	if strings.TrimSpace(req.VendorName) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "validation_error",
			"message": "vendorName is required",
		})
	}

	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "validation_error",
			"message": "items cannot be empty",
		})
	}

	poID := uuid.New()
	orderNumber := fmt.Sprintf("PO/%s/%s", time.Now().Format("200601"), strings.ToUpper(uuid.NewString()[:6]))

	var totalAmount float64
	poItems := make([]model.PurchaseOrderItem, 0, len(req.Items))

	for _, it := range req.Items {
		if it.Quantity <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "validation_error",
				"message": "quantity must be greater than zero",
			})
		}
		if it.UnitPrice < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "validation_error",
				"message": "unitPrice cannot be negative",
			})
		}

		productName := "Custom Item"
		if it.ProductID != uuid.Nil {
			p, err := h.productRepo.FindByID(c.Context(), companyID, it.ProductID)
			if err == nil && p != nil {
				productName = p.Name
			}
		}

		subtotal := float64(it.Quantity) * it.UnitPrice
		totalAmount += subtotal

		poItems = append(poItems, model.PurchaseOrderItem{
			ID:              uuid.New(),
			CompanyID:       companyID,
			PurchaseOrderID: poID,
			ProductID:       it.ProductID,
			ProductName:     productName,
			Quantity:        it.Quantity,
			UnitPrice:       it.UnitPrice,
			Subtotal:        subtotal,
		})
	}

	po := model.PurchaseOrder{
		ID:          poID,
		CompanyID:   companyID,
		OrderNumber: orderNumber,
		VendorName:  req.VendorName,
		Status:      "DRAFT",
		TotalAmount: totalAmount,
		Notes:       req.Notes,
		Items:       poItems,
	}

	if err := h.poRepo.Create(c.Context(), &po); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"code":    "success",
		"message": "purchase order created successfully",
		"data":    po,
	})
}

func (h *PurchaseOrderHandler) GetByID(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "unauthorized",
			"message": err.Error(),
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "bad_request",
			"message": "invalid purchase order id",
		})
	}

	po, err := h.poRepo.FindByID(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}
	if po == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "not_found",
			"message": "purchase order not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code": "success",
		"data": po,
	})
}

func (h *PurchaseOrderHandler) List(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "unauthorized",
			"message": err.Error(),
		})
	}

	pos, err := h.poRepo.FindAll(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code": "success",
		"data": pos,
	})
}

func (h *PurchaseOrderHandler) UpdateStatus(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "unauthorized",
			"message": err.Error(),
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "bad_request",
			"message": "invalid purchase order id",
		})
	}

	var req model.UpdatePOStatusRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "bad_request",
			"message": "invalid request body",
		})
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if status != "CONFIRMED" && status != "CANCELLED" && status != "DRAFT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "validation_error",
			"message": "status must be DRAFT, CONFIRMED, or CANCELLED",
		})
	}

	if err := h.poRepo.UpdateStatus(c.Context(), companyID, id, status); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "not_found",
				"message": "purchase order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    "success",
		"message": "purchase order status updated successfully",
	})
}

func (h *PurchaseOrderHandler) Delete(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "unauthorized",
			"message": err.Error(),
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "bad_request",
			"message": "invalid purchase order id",
		})
	}

	if err := h.poRepo.Delete(c.Context(), companyID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "not_found",
				"message": "purchase order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    "success",
		"message": "purchase order deleted successfully",
	})
}
