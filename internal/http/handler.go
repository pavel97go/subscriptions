package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/pavel97go/subscriptions/internal/domain"
	"github.com/pavel97go/subscriptions/internal/logger"
	"github.com/pavel97go/subscriptions/internal/repo"
	"github.com/pavel97go/subscriptions/internal/util"
)

type Handler struct{ r *repo.Repo }

func NewHandler(r *repo.Repo) *Handler { return &Handler{r: r} }

func reqCtx(c *fiber.Ctx) context.Context {
	if uc := c.UserContext(); uc != nil {
		return uc
	}
	return context.Background()
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var in domain.SubscriptionDTO
	if err := c.BodyParser(&in); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	in.ServiceName = strings.TrimSpace(in.ServiceName)
	if in.ServiceName == "" {
		return fiber.NewError(http.StatusBadRequest, "service_name is required")
	}
	if in.Price < 0 {
		return fiber.NewError(http.StatusBadRequest, "price must be >= 0")
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
		if t.Before(sm) {
			return fiber.NewError(http.StatusBadRequest, "end_date must be >= start_date")
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
	logger.Log.Infof("http create: user_id=%s service=%s", s.UserID, s.ServiceName)
	id, err := h.r.Create(reqCtx(c), s)
	if err != nil {
		logger.Log.Errorf("http create error: %v", err)
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.Status(http.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid id")
	}
	s, err := h.r.Get(reqCtx(c), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fiber.NewError(http.StatusNotFound, "not found")
		}
		logger.Log.Errorf("http get error: %v", err)
		return fiber.NewError(http.StatusInternalServerError, "internal error")
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
	if s := strings.TrimSpace(c.Query("service_name")); s != "" {
		svc = &s
	}

	logger.Log.Infof("http list: limit=%d offset=%d user_id=%v service=%v", limit, offset, uid, svc)
	items, err := h.r.ListFiltered(
		reqCtx(c),
		repo.ListFilter{UserID: uid, ServiceName: svc},
		limit, offset,
	)
	if err != nil {
		logger.Log.Errorf("http list error: %v", err)
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
	in.ServiceName = strings.TrimSpace(in.ServiceName)
	if in.ServiceName == "" {
		return fiber.NewError(http.StatusBadRequest, "service_name is required")
	}
	if in.Price < 0 {
		return fiber.NewError(http.StatusBadRequest, "price must be >= 0")
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
		if t.Before(sm) {
			return fiber.NewError(http.StatusBadRequest, "end_date must be >= start_date")
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
	logger.Log.Infof("http update: id=%s user_id=%s service=%s", id, s.UserID, s.ServiceName)
	if err := h.r.Update(reqCtx(c), id, s); err != nil {
		logger.Log.Errorf("http update error: %v", err)
		return fiber.NewError(http.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(http.StatusNoContent)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(http.StatusBadRequest, "invalid id")
	}
	logger.Log.Infof("http delete: id=%s", id)
	if err := h.r.Delete(reqCtx(c), id); err != nil {
		logger.Log.Errorf("http delete error: %v", err)
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
	if s := strings.TrimSpace(c.Query("service_name")); s != "" {
		svc = &s
	}

	logger.Log.Infof("http summary: from=%s to=%s user_id=%v service=%v", util.MonthStr(from), util.MonthStr(to), uid, svc)
	total, err := h.r.Summary(
		reqCtx(c),
		repo.SummaryFilter{UserID: uid, ServiceName: svc, From: from, To: to},
	)
	if err != nil {
		logger.Log.Errorf("http summary error: %v", err)
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
