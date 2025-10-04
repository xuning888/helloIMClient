package pkg

// RestResult
// Note: WebAPI 统一的返回类型
type RestResult[T any] struct {
	Code int    `json:"code"`
	Data T      `json:"data"`
	Msg  string `json:"msg"`
}
