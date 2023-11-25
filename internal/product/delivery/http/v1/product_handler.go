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
	//beego.Router("/api/v1/link-voucher/:id", pHandler, "get:GetLinkVoucher")
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
// @Param    max_count query int false "max_count"
// @Router /v1/scrape-product-tokopedia-phone-category [post]
func (h *ProductHandler) ScrapeProducts() {
	maxCount, err := strconv.Atoi(h.Ctx.Input.Query("max_count"))
	if err != nil {
		h.ResponseError(h.Ctx, http.StatusBadRequest, response.QueryParamInvalidCode, response.ErrorCodeText(response.QueryParamInvalidCode, h.Locale.Lang), err)
		return
	}


	err = h.ProductUsecase.ScrapeProducts(h.Ctx, maxCount)
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
//
//// GetLinkVoucher
//// @Title GetLinkVoucher
//// @Tags Product
//// @Summary GetLinkVoucher
//// @Produce json
//// @Param Accept-Language header string false "lang"
//// @Success 200 {object} swagger.BaseResponse{data=[]domain.ProductVoucherBookResponse,errors=[]object}
//// @Failure 408 {object} swagger.RequestTimeoutResponse{errors=[]object,data=object}
//// @Failure 500 {object} swagger.InternalServerErrorResponse{errors=[]object,data=object}
//// @Param    id path int true "id customer"
//// @router /v1/link-voucher/{id} [get]
//func (h *ProductHandler) GetLinkVoucher() {
//	pathParam, err := strconv.Atoi(h.Ctx.Input.Param(":id"))
//
//	if err != nil || pathParam < 1 {
//		h.ResponseError(h.Ctx, http.StatusBadRequest, response.PathParamInvalidCode, response.ErrorCodeText(response.PathParamInvalidCode, h.Locale.Lang), err)
//		return
//	}
//
//	result, err := h.ProductUsecase.GetVoucherByProductId(h.Ctx, pathParam)
//	if err != nil {
//		if errors.Is(err, response.ErrVoucherNotAvailable) {
//			h.ResponseError(h.Ctx, http.StatusBadRequest, response.VoucherNotAvailable, response.ErrorCodeText(response.VoucherNotAvailable, h.Locale.Lang), err)
//			return
//		}
//		if errors.Is(err, response.ErrTransactionCompletePurchase30Days) {
//			h.ResponseError(h.Ctx, http.StatusBadRequest, response.TransactionCompletePurchase30Days, response.ErrorCodeText(response.TransactionCompletePurchase30Days, h.Locale.Lang), err)
//			return
//		}
//		if errors.Is(err, response.ErrTransactionMinimum) {
//			h.ResponseError(h.Ctx, http.StatusBadRequest, response.TransactionMinimum, response.ErrorCodeText(response.TransactionMinimum, h.Locale.Lang), err)
//			return
//		}
//		if errors.Is(err, response.ErrProductAlreadyBookVoucher) {
//			h.ResponseError(h.Ctx, http.StatusBadRequest, response.ProductAlreadyBookVoucher, response.ErrorCodeText(response.ProductAlreadyBookVoucher, h.Locale.Lang), err)
//			return
//		}
//		if errors.Is(err, response.ErrProductAlreadyGetVoucher) {
//			h.ResponseError(h.Ctx, http.StatusBadRequest, response.ProductAlreadyGetVoucher, response.ErrorCodeText(response.ProductAlreadyGetVoucher, h.Locale.Lang), err)
//			return
//		}
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			h.ResponseError(h.Ctx, http.StatusBadRequest, response.DataNotFoundCodeError, response.ErrorCodeText(response.DataNotFoundCodeError, h.Locale.Lang), err)
//			return
//		}
//		if errors.Is(err, context.DeadlineExceeded) {
//			h.ResponseError(h.Ctx, http.StatusRequestTimeout, response.RequestTimeoutCodeError, response.ErrorCodeText(response.RequestTimeoutCodeError, h.Locale.Lang), err)
//			return
//		}
//		h.ResponseError(h.Ctx, http.StatusInternalServerError, response.ServerErrorCode, response.ErrorCodeText(response.ServerErrorCode, h.Locale.Lang), err)
//		return
//	}
//	h.Ok(h.Ctx, h.Tr("message.success"), result)
//	return
//}
