package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/pavel97go/subscriptions/internal/domain"
	"github.com/pavel97go/subscriptions/internal/repo"
	"github.com/pavel97go/subscriptions/internal/util"
)

type Handler struct{ r *repo.Repo }

func NewHandler(r *repo.Repo) *Handler { return &Handler{r: r} }

func (h *Handler) Create(c *fiber.Ctx) error {
	var in domain.SubscriptionDTO
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	sm, err := util.ParseMonth(in.StartDate)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid start_date, expected MM-YYYY")
	}
	var em *time.Time
	if in.EndDate != nil && *in.EndDate != "" {
		t, err := util.ParseMonth(*in.EndDate)
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid end_date, expected MM-YYYY")
		}
		em = &t
	}
	s := domain.Subscription{
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID,
		StartMonth:  sm,
		EndMonth:    em,
	}
	id, err := h.r.Create(context.Background(), s)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid id")
	}
	s, err := h.r.Get(context.Background(), id)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, err.Error())
	}
	return c.JSON(toResp(s))
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit, offset := 50, 0
	if v := c.QueryInt("limit"); v > 0 && v <= 200 {
		limit = v
	}
	if v := c.QueryInt("offset"); v >= 0 {
		offset = v
	}

	var uid *uuid.UUID
	if s := c.Query("user_id"); s != "" {
		u, err := uuid.Parse(s)
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid user_id")
		}
		uid = &u
	}
	var svc *string
	if s := c.Query("service_name"); s != "" {
		svc = &s
	}

	items, err := h.r.ListFiltered(
		context.Background(),
		repo.ListFilter{UserID: uid, ServiceName: svc},
		limit, offset,
	)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	out := make([]domain.SubscriptionResponse, 0, len(items))
	for _, s := range items {
		out = append(out, toResp(s))
	}
	return c.JSON(out)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid id")
	}
	var in domain.SubscriptionDTO
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	sm, err := util.ParseMonth(in.StartDate)
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid start_date, expected MM-YYYY")
	}
	var em *time.Time
	if in.EndDate != nil && *in.EndDate != "" {
		t, err := util.ParseMonth(*in.EndDate)
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid end_date, expected MM-YYYY")
		}
		em = &t
	}
	s := domain.Subscription{
		ServiceName: in.ServiceName,
		Price:       in.Price,
		UserID:      in.UserID, // не проверяем существование
		StartMonth:  sm,
		EndMonth:    em,
	}
	if err := h.r.Update(context.Background(), id, s); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid id")
	}
	if err := h.r.Delete(context.Background(), id); err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	from, err := util.ParseMonth(c.Query("from"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "from (MM-YYYY) required")
	}
	to, err := util.ParseMonth(c.Query("to"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "to (MM-YYYY) required")
	}
	if to.Before(from) {
		return fiber.NewError(http.StatusBadRequest, "`to` must be >= `from`")
	}

	var uid *uuid.UUID
	if s := c.Query("user_id"); s != "" {
		u, err := uuid.Parse(s)
		if err != nil {
			return fiber.NewError(http.StatusBadRequest, "invalid user_id")
		}
		uid = &u
	}
	var svc *string
	if s := c.Query("service_name"); s != "" {
		svc = &s
	}

	total, err := h.r.Summary(
		c.Context(),
		repo.SummaryFilter{UserID: uid, ServiceName: svc, From: from, To: to},
	)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"total": total})
}

func toResp(s domain.Subscription) domain.SubscriptionResponse {
	out := domain.SubscriptionResponse{
		ID:          s.ID,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID,
		StartDate:   util.MonthStr(s.StartMonth),
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
	if s.EndMonth != nil {
		e := util.MonthStr(*s.EndMonth)
		out.EndDate = &e
	}
	return out
}
