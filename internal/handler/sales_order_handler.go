package handler

import (
	"encoding/json"
	"strings"

	"be-integration-example/internal/model"
	"be-integration-example/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type SalesOrderHandler struct {
	repo repository.SalesOrderRepository
}

func NewSalesOrderHandler(repo repository.SalesOrderRepository) *SalesOrderHandler {
	return &SalesOrderHandler{repo: repo}
}

func (h *SalesOrderHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/sales-orders")
	group.Post("/", h.Create)
	group.Get("/", h.List)
	group.Get("/:id", h.GetByID)
	group.Patch("/:id/status", h.UpdateStatus)
	group.Delete("/:id", h.Delete)
}

func (h *SalesOrderHandler) Create(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	var req model.CreateSalesOrderRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid request body"})
	}

	if req.CustomerID == uuid.Nil || strings.TrimSpace(req.OrderNumber) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "validation_error", "message": "customerId and orderNumber are required"})
	}

	order := &model.SalesOrder{
		ID:          uuid.New(),
		CompanyID:   companyID,
		CustomerID:  req.CustomerID,
		OrderNumber: req.OrderNumber,
		TotalAmount: req.TotalAmount,
		Status:      "DRAFT",
	}

	if err := h.repo.Create(c.Context(), order); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": order})
}

func (h *SalesOrderHandler) List(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	orders, err := h.repo.FindAll(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": orders})
}

func (h *SalesOrderHandler) GetByID(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	order, err := h.repo.FindByID(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}
	if order == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "sales order not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": order})
}

func (h *SalesOrderHandler) UpdateStatus(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	var req model.UpdateSalesOrderStatusRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid request body"})
	}

	if strings.TrimSpace(req.Status) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "validation_error", "message": "status is required"})
	}

	if err := h.repo.UpdateStatus(c.Context(), companyID, id, req.Status); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "sales order not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "status updated successfully"})
}

func (h *SalesOrderHandler) Delete(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	if err := h.repo.Delete(c.Context(), companyID, id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "sales order not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "sales order deleted successfully"})
}
