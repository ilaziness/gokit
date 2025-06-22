// Copyright (c) 2024 ilaziness. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
//
// Author: ilaziness  https://github.com/ilaziness

// Package errcode 错误码相关的功能

package errcode

import "fmt"

type Code struct {
	messageData []any
	Code        int
	Message     string
	Data        any
}

// NewCode 新建一个错误码对象
// 需要提供一个错误码和对应错误消息
func NewCode(code int, msg string) *Code {
	return &Code{
		Code:    code,
		Message: msg,
		Data:    struct{}{},
	}
}

func (ec *Code) Error() string {
	if len(ec.messageData) > 0 {
		return fmt.Sprintf(ec.Message, ec.messageData...)
	}
	return ec.Message
}

// SetData 设置响应数据
func (ec *Code) SetData(data any) *Code {
	ec.Data = data
	return ec
}

// SetMessage 设置消息内容
func (ec *Code) SetMessage(msg string) *Code {
	ec.Message = msg
	return ec
}

// SetMessageData 设置消息格式化动态数据
func (ec *Code) SetMessageData(data ...any) *Code {
	ec.messageData = data
	return ec
}

var (
	ReqErr = NewCode(400, "request error")
)
