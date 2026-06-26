package handlers

import "github.com/gofiber/fiber/v2"

const (
	defaultLimit = 20
	maxLimit     = 100
)

// pageParams membaca query ?page=&limit=, menormalkan nilainya, dan menghitung offset.
// page minimal 1; limit dibatasi 1..maxLimit dengan default defaultLimit.
func pageParams(c *fiber.Ctx) (page, limit, offset int) {
	page = c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit = c.QueryInt("limit", defaultLimit)
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	offset = (page - 1) * limit
	return page, limit, offset
}

// paginated membungkus data + metadata pagination dengan bentuk yang konsisten
// di seluruh endpoint daftar (warga, user, transaksi).
func paginated(c *fiber.Ctx, data any, page, limit int, total int64) error {
	return c.JSON(fiber.Map{
		"data": data,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
