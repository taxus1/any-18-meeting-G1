package common

// Result 统一返回结构。code=0 成功，非 0 业务失败；前端按 code 判断，不看 HTTP 状态码。
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func OK(data interface{}) Result {
	return Result{Code: 0, Msg: "success", Data: data}
}

func Fail(msg string) Result {
	return Result{Code: 1, Msg: msg}
}

func FailCode(code int, msg string) Result {
	return Result{Code: code, Msg: msg}
}
