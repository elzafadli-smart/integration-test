package handler

import (
	"encoding/json"
	"strings"

	"be-integration-example/internal/model"
	"be-integration-example/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CustomerHandler struct {
	repo repository.CustomerRepository
}

func NewCustomerHandler(repo repository.CustomerRepository) *CustomerHandler {
	return &CustomerHandler{repo: repo}
}

func (h *CustomerHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/customers")
	group.Post("/", h.Create)
	group.Get("/", h.List)
	group.Get("/:id", h.GetByID)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

func (h *CustomerHandler) Create(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	var req model.CreateCustomerRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid request body"})
	}

	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "validation_error", "message": "name and email are required"})
	}

	customer := &model.Customer{
		ID:        uuid.New(),
		CompanyID: companyID,
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
	}

	if err := h.repo.Create(c.Context(), customer); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": customer})
}

func (h *CustomerHandler) List(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	customers, err := h.repo.FindAll(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": customers})
}

func (h *CustomerHandler) GetByID(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	customer, err := h.repo.FindByID(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}
	if customer == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "customer not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": customer})
}

func (h *CustomerHandler) Update(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	customer, err := h.repo.FindByID(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}
	if customer == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "customer not found"})
	}

	var req model.UpdateCustomerRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid request body"})
	}

	if req.Name != nil {
		customer.Name = *req.Name
	}
	if req.Email != nil {
		customer.Email = *req.Email
	}
	if req.Phone != nil {
		customer.Phone = *req.Phone
	}

	if err := h.repo.Update(c.Context(), customer); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": customer})
}

func (h *CustomerHandler) Delete(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	if err := h.repo.Delete(c.Context(), companyID, id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "customer not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "customer deleted successfully"})
}
