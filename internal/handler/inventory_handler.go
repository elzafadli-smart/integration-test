package handler

import (
	"encoding/json"
	"strings"

	"be-integration-example/internal/model"
	"be-integration-example/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type InventoryHandler struct {
	repo repository.InventoryRepository
}

func NewInventoryHandler(repo repository.InventoryRepository) *InventoryHandler {
	return &InventoryHandler{repo: repo}
}

func (h *InventoryHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/inventory")
	group.Post("/adjust", h.Adjust)
	group.Get("/product/:productId", h.GetByProduct)
	group.Delete("/:id", h.Delete)
}

func (h *InventoryHandler) Adjust(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	var req model.AdjustInventoryRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid request body"})
	}

	if req.ProductID == uuid.Nil || strings.TrimSpace(req.Warehouse) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "validation_error", "message": "productId and warehouse are required"})
	}

	inv := &model.Inventory{
		ID:        uuid.New(),
		CompanyID: companyID,
		ProductID: req.ProductID,
		Warehouse: req.Warehouse,
		Quantity:  req.Quantity,
	}

	if err := h.repo.Adjust(c.Context(), inv); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": inv, "message": "inventory adjusted successfully"})
}

func (h *InventoryHandler) GetByProduct(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	productID, err := uuid.Parse(c.Params("productId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid productId format"})
	}

	items, err := h.repo.FindByProduct(c.Context(), companyID, productID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": items})
}

func (h *InventoryHandler) Delete(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"code": "unauthorized", "message": err.Error()})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "bad_request", "message": "invalid id format"})
	}

	if err := h.repo.Delete(c.Context(), companyID, id); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"code": "not_found", "message": "inventory item not found"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "inventory deleted successfully"})
}
