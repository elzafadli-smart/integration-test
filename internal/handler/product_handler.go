package handler

import (
	"encoding/json"
	"errors"
	"strings"

	"be-integration-example/internal/model"
	"be-integration-example/internal/repository"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductHandler struct {
	repo repository.ProductRepository
}

func NewProductHandler(repo repository.ProductRepository) *ProductHandler {
	return &ProductHandler{repo: repo}
}

func (h *ProductHandler) RegisterRoutes(router fiber.Router) {
	products := router.Group("/products")
	products.Post("/", h.Create)
	products.Get("/", h.List)
	products.Get("/:id", h.GetByID)
	products.Put("/:id", h.Update)
	products.Delete("/:id", h.Delete)
}

func getCompanyID(c fiber.Ctx) (uuid.UUID, error) {
	headerVal := c.Get("X-Company-Id")
	if headerVal == "" {
		return uuid.Nil, errors.New("missing X-Company-Id header")
	}
	return uuid.Parse(headerVal)
}

func (h *ProductHandler) Create(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "unauthorized",
			"message": err.Error(),
		})
	}

	var req model.CreateProductRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "bad_request",
			"message": "invalid request body",
		})
	}

	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.SKU) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "validation_error",
			"message": "name and sku are required",
		})
	}
	if req.Price < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "validation_error",
			"message": "price cannot be negative",
		})
	}

	product := model.Product{
		ID:        uuid.New(),
		CompanyID: companyID,
		Name:      req.Name,
		SKU:       req.SKU,
		Barcode:   req.Barcode,
		Price:     req.Price,
		Stock:     req.Stock,
	}

	if err := h.repo.Create(c.Context(), &product); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"code":    "success",
		"message": "product created successfully",
		"data":    product,
	})
}

func (h *ProductHandler) GetByID(c fiber.Ctx) error {
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
			"message": "invalid product id",
		})
	}

	product, err := h.repo.FindByID(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}
	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "not_found",
			"message": "product not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code": "success",
		"data": product,
	})
}

func (h *ProductHandler) List(c fiber.Ctx) error {
	companyID, err := getCompanyID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "unauthorized",
			"message": err.Error(),
		})
	}

	products, err := h.repo.FindAll(c.Context(), companyID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code": "success",
		"data": products,
	})
}

func (h *ProductHandler) Update(c fiber.Ctx) error {
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
			"message": "invalid product id",
		})
	}

	product, err := h.repo.FindByID(c.Context(), companyID, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}
	if product == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "not_found",
			"message": "product not found",
		})
	}

	var req model.UpdateProductRequest
	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "bad_request",
			"message": "invalid request body",
		})
	}

	if req.Name != nil {
		if strings.TrimSpace(*req.Name) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "validation_error",
				"message": "name cannot be empty",
			})
		}
		product.Name = *req.Name
	}
	if req.Price != nil {
		if *req.Price < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "validation_error",
				"message": "price cannot be negative",
			})
		}
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.Barcode != nil {
		product.Barcode = *req.Barcode
	}

	if err := h.repo.Update(c.Context(), product); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    "success",
		"message": "product updated successfully",
		"data":    product,
	})
}

func (h *ProductHandler) Delete(c fiber.Ctx) error {
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
			"message": "invalid product id",
		})
	}

	if err := h.repo.Delete(c.Context(), companyID, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    "not_found",
				"message": "product not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    "success",
		"message": "product deleted successfully",
	})
}
