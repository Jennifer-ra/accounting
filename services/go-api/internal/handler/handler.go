package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Jennifer-ra/accounting/services/go-api/internal/middleware"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/response"
	"github.com/Jennifer-ra/accounting/services/go-api/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	auth *service.AuthService
	app  *service.AccountingService
}

func New(authSvc *service.AuthService, appSvc *service.AccountingService) *Handler {
	return &Handler{auth: authSvc, app: appSvc}
}

func (h *Handler) Register(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) { response.OK(c, gin.H{"status": "ok"}) })
	api := router.Group("/api/v1")
	api.POST("/auth/wechat-login", h.wechatLogin)

	private := api.Group("")
	private.Use(middleware.Auth(h.auth))
	private.GET("/me", h.me)
	private.GET("/bootstrap", h.bootstrap)
	private.GET("/categories", h.categories)
	private.POST("/categories", h.addCategory)
	private.GET("/accounts", h.accounts)
	private.POST("/accounts", h.addAccount)
	private.GET("/transactions", h.transactions)
	private.POST("/transactions", h.createTransaction)
	private.PUT("/transactions/:id", h.updateTransaction)
	private.DELETE("/transactions/:id", h.deleteTransaction)
	private.GET("/summary", h.summary)
	private.PUT("/budgets/:month", h.setBudget)
}

func (h *Handler) wechatLogin(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", "invalid login payload")
		return
	}
	result, err := h.auth.LoginWithWeChat(input)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "wechat_login_failed", err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) me(c *gin.Context) {
	book, err := h.app.DefaultBook(middleware.UserID(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	response.OK(c, gin.H{"book": book})
}

func (h *Handler) bootstrap(c *gin.Context) {
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))
	result, err := h.app.Bootstrap(middleware.UserID(c), month)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	response.OK(c, result)
}

func (h *Handler) categories(c *gin.Context) {
	items, err := h.app.Categories(middleware.UserID(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	response.OK(c, items)
}

func (h *Handler) addCategory(c *gin.Context) {
	var input service.NameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", "invalid category payload")
		return
	}
	item, err := h.app.AddCategory(middleware.UserID(c), input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	response.Created(c, item)
}

func (h *Handler) accounts(c *gin.Context) {
	items, err := h.app.Accounts(middleware.UserID(c))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	response.OK(c, items)
}

func (h *Handler) addAccount(c *gin.Context) {
	var input service.NameInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", "invalid account payload")
		return
	}
	item, err := h.app.AddAccount(middleware.UserID(c), input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	response.Created(c, item)
}

func (h *Handler) transactions(c *gin.Context) {
	items, err := h.app.Transactions(middleware.UserID(c), c.Query("month"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	response.OK(c, gin.H{"items": items})
}

func (h *Handler) createTransaction(c *gin.Context) {
	var input service.TransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", "invalid transaction payload")
		return
	}
	item, err := h.app.CreateTransaction(middleware.UserID(c), input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	response.Created(c, item)
}

func (h *Handler) updateTransaction(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", "invalid transaction id")
		return
	}
	var input service.TransactionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", "invalid transaction payload")
		return
	}
	item, err := h.app.UpdateTransaction(middleware.UserID(c), id, input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	response.OK(c, item)
}

func (h *Handler) deleteTransaction(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", "invalid transaction id")
		return
	}
	if err := h.app.DeleteTransaction(middleware.UserID(c), id); err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) summary(c *gin.Context) {
	month := c.DefaultQuery("month", time.Now().Format("2006-01"))
	item, err := h.app.Summary(middleware.UserID(c), month)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	response.OK(c, item)
}

func (h *Handler) setBudget(c *gin.Context) {
	var input service.BudgetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", "invalid budget payload")
		return
	}
	item, err := h.app.SetBudget(middleware.UserID(c), c.Param("month"), input)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	response.OK(c, item)
}

func parseID(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	return uint(id), err
}
