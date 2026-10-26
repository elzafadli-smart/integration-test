package handler

import (
	"encoding/json"
	"strings"

	"be-integration-example/internal/model"
	"be-integration-example/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type SupplierHandler struct {
	repo repository.SupplierRepository
}

func NewSupplierHandler(repo repository.SupplierRepository) *SupplierHandler {
	return &SupplierHandler{repo: repo}
}

func (h *SupplierHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/suppliers")
	group.Post("/", h.Create)
	group.Get("/", h.List)
	group.Get("/:id", h.GetByID)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

func (h *SupplierHandler) Create(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	var req model.CreateSupplierRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid request body"})
	}

	if strings.TrimSpace(req.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "validation_error", "message": "name is required"})
	}

	supplier := &model.Supplier{
		ID:            uuid.New(),
		CompanyID:     companyID,
		Name:          req.Name,
		ContactPerson: req.ContactPerson,
		Phone:         req.Phone,
	}

	if err := h.repo.Create(c.Context(), supplier); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": supplier})
}

func (h *SupplierHandler) List(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	suppliers, err := h.repo.FindAll(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": suppliers})
}

func (h *SupplierHandler) GetByID(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	supplier, err := h.repo.FindByID(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}
	if supplier == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "supplier not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": supplier})
}

func (h *SupplierHandler) Update(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	supplier, err := h.repo.FindByID(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}
	if supplier == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "supplier not found"})
	}

	var req model.UpdateSupplierRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid request body"})
	}

	if req.Name != nil {
		supplier.Name = *req.Name
	}
	if req.ContactPerson != nil {
		supplier.ContactPerson = *req.ContactPerson
	}
	if req.Phone != nil {
		supplier.Phone = *req.Phone
	}

	if err := h.repo.Update(c.Context(), supplier); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": supplier})
}

func (h *SupplierHandler) Delete(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	if err := h.repo.Delete(c.Context(), companyID, id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "supplier not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "supplier deleted successfully"})
}
