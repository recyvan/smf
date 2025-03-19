package command

import "fmt"

// Error 必须实现 error 接口
type Error struct {
	Code    int
	Message string
	Details map[string]interface{}
}

// 实现 error 接口
func (e *Error) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewError 创建新的错误实例（返回指针类型）
func NewError(code int, msg string, details ...map[string]interface{}) *Error {
	err := &Error{
		Code:    code,
		Message: msg,
		Details: make(map[string]interface{}),
	}

	if len(details) > 0 {
		for k, v := range details[0] {
			err.Details[k] = v
		}
	}
	return err
}

// 保证类型断言可用
var _ error = (*Error)(nil) // 添加接口实现检查
