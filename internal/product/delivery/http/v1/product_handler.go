package v1

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/radyatamaa/scrap-brick-app/internal"
	"github.com/radyatamaa/scrap-brick-app/internal/domain"
	"github.com/radyatamaa/scrap-brick-app/pkg/response"
	"github.com/radyatamaa/scrap-brick-app/pkg/zaplogger"
	"gorm.io/gorm"
)

type ProductHandler struct {
	ZapLogger zaplogger.Logger
	internal.BaseController
	response.ApiResponse
	ProductUsecase domain.ProductUseCase
}

func NewProductHandler(customerUsecase domain.ProductUseCase, zapLogger zaplogger.Logger) {
	pHandler := &ProductHandler{
		ZapLogger:       zapLogger,
		ProductUsecase: customerUsecase,
	}
	beego.Router("/api/v1/scrape-product-tokopedia-phone-category", pHandler, "post:ScrapeProducts")
	beego.Router("/api/v1/product", pHandler, "get:GetProducts")
}

func (h *ProductHandler) Prepare() {
	// check user access when needed
	h.SetLangVersion()
}

// ScrapeProducts
// @Title ScrapeProducts
// @Tags Product
// @Summary ScrapeProducts
// @Produce json
// @Param Accept-Language header string false "lang"
// @Success 200 {object} swagger.BaseResponse{errors=[]object}
// @Failure 400 {object} swagger.BadRequestErrorValidationResponse{errors=[]swagger.ValidationErrors,data=object}
// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
// @Router /v1/scrape-product-tokopedia-phone-category [post]
func (h *ProductHandler) ScrapeProducts() {
	err := h.ProductUsecase.ScrapeProducts(h.Ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.ResponseError(h.Ctx, http.StatusRequestTimeout, response.RequestTimeoutCodeError, response.ErrorCodeText(response.RequestTimeoutCodeError, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.DataNotFoundCodeError, response.ErrorCodeText(response.DataNotFoundCodeError, h.Locale.Lang), err)
			return
		}
		h.ResponseError(h.Ctx, http.StatusInternalServerError, response.ServerErrorCode, response.ErrorCodeText(response.ServerErrorCode, h.Locale.Lang), err)
		return
	}
	h.Ok(h.Ctx, h.Tr("message.success"), nil)
	return
}

// GetProducts
// @Title GetProducts
// @Tags Product
// @Summary GetProducts
// @Produce json
// @Param Accept-Language header string false "lang"
// @Success 200 {object} swagger.BaseResponse{data=[]domain.Product,errors=[]object}
// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
// @Param    limit query int true "limit"
// @router /v1/product [get]
func (h *ProductHandler) GetProducts() {
	limit, err := strconv.Atoi(h.Ctx.Input.Query("limit"))
	if err != nil {
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.QueryParamInvalidCode, response.ErrorCodeText(response.QueryParamInvalidCode, h.Locale.Lang), err)
		return
	}

	result, err := h.ProductUsecase.GetProducts(h.Ctx, limit)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.ResponseError(h.Ctx, http.StatusBadRequest, response.DataNotFoundCodeError, response.ErrorCodeText(response.DataNotFoundCodeError, h.Locale.Lang), err)
			return
		}
		if errors.Is(err, context.DeadlineExceeded) {
			h.ResponseError(h.Ctx, http.StatusRequestTimeout, response.RequestTimeoutCodeError, response.ErrorCodeText(response.RequestTimeoutCodeError, h.Locale.Lang), err)
			return
		}
		h.ResponseError(h.Ctx, http.StatusInternalServerError, response.ServerErrorCode, response.ErrorCodeText(response.ServerErrorCode, h.Locale.Lang), err)
		return
	}
	h.Ok(h.Ctx, h.Tr("message.success"), result)
	return
}
