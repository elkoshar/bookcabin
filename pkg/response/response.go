package response

import (
	"net/http"
	"time"

	"github.com/go-chi/render"
)

type Response struct {
	Code               int            `json:"-"`
	Data               interface{}    `json:"data,omitempty"`
	Error              Error          `json:"error"`
	Message            string         `json:"message"`
	ServerTime         int64          `json:"serverTime"`
	IsStringCode       bool           `json:"-"`
	CodeRender         interface{}    `json:"code"`
	OverrideStatusText map[int]string `json:"-"`
}

func NewResponse() Response {
	return Response{}
}

type Error struct {
	Status bool   `json:"status" example:"false"`
	Msg    string `json:"msg" example:" "`
	Code   int    `json:"code" example:"0"`
}

func NewError(err error, code int) *Error {
	return &Error{
		Status: true,
		Msg:    err.Error(),
		Code:   code,
	}
}

func (e Error) Error() string {
	return e.Msg
}

func (res *Response) Render(w http.ResponseWriter, r *http.Request, statusCode ...int) {

	if res.Code == 0 {
		res.Code = http.StatusOK
	}

	res.ServerTime = time.Now().Unix()
	render.Status(r, res.Code)

	if len(statusCode) > 0 {
		render.Status(r, statusCode[0])
	} else {
		render.Status(r, res.Code)
	}

	res.RenderStatusCode()
	render.JSON(w, r, res)
}

func (res *Response) RenderStatusCode() {
	res.CodeRender = res.Code
	if !res.IsStringCode {
		return
	}

	statusText := res.OverrideStatusText[res.Code]
	if statusText == "" {
		statusText = http.StatusText(res.Code)
	}
	res.CodeRender = statusText
}
