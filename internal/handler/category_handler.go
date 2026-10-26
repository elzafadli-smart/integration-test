package handler

import (
	"encoding/json"
	"strings"

	"be-integration-example/internal/model"
	"be-integration-example/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CategoryHandler struct {
	repo repository.CategoryRepository
}

func NewCategoryHandler(repo repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{repo: repo}
}

func (h *CategoryHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/categories")
	group.Post("/", h.Create)
	group.Get("/", h.List)
	group.Get("/:id", h.GetByID)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

func (h *CategoryHandler) Create(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	var req model.CreateCategoryRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid request body"})
	}

	if strings.TrimSpace(req.Name) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "validation_error", "message": "name is required"})
	}

	category := &model.Category{
		ID:          uuid.New(),
		CompanyID:   companyID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.repo.Create(c.Context(), category); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": category})
}

func (h *CategoryHandler) List(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	categories, err := h.repo.FindAll(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": categories})
}

func (h *CategoryHandler) GetByID(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	category, err := h.repo.FindByID(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}
	if category == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "category not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": category})
}

func (h *CategoryHandler) Update(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	category, err := h.repo.FindByID(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}
	if category == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "category not found"})
	}

	var req model.UpdateCategoryRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid request body"})
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = *req.Description
	}

	if err := h.repo.Update(c.Context(), category); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": category})
}

func (h *CategoryHandler) Delete(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	if err := h.repo.Delete(c.Context(), companyID, id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "category not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "category deleted successfully"})
}
