package main

import (
	"net/http"
	"strings"

	"github.com/zy99978455-otw/flash-monitor/internal/data"
	"github.com/zy99978455-otw/flash-monitor/internal/validator"
)

func (app *application) listTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// DTO
	// 1. 定义一个输入结构体来承接查询参数
	var input struct {
		FromAddress string
		ToAddress   string
		data.Filters
	}

	// 2. 初始化验证器并从 URL 拿参数
	v := validator.New()
	qs := r.URL.Query()

	input.FromAddress = strings.ToLower(app.readString(qs, "from_address", ""))
	input.ToAddress = strings.ToLower(app.readString(qs, "to_address", ""))

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// 默认按块号倒序排
	input.Filters.Sort = app.readString(qs, "sort", "-block_number")
	input.Filters.SortSafelist = []string{"block_number", "amount", "-block_number", "-amount"}

	// 3. 执行校验
	if input.FromAddress != "" {
		v.Check(validator.IsEthAddress(input.FromAddress), "from_address", "必须是合法的以太坊16进制地址格式")
	}
	if input.ToAddress != "" {
		v.Check(validator.IsEthAddress(input.ToAddress), "to_address", "必须是合法的以太坊16进制地址格式")
	}

	// 校验基础的分页与排序规则
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// 4. 调用升级后的 GetAll
	events, metadata, err := app.models.TransferEvents.GetAll(input.FromAddress, input.ToAddress, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// 5. 返回带元数据的 JSON
	err = app.writeJSON(w, http.StatusOK, envelope{"data": events, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
