// Code generated by protoc-gen-ecode v0.1, DO NOT EDIT.
// source: api.proto

package api

import (
	"github.com/itering/kratos/pkg/ecode"
)

// to suppressed 'imported but not used warning'
var _ ecode.Codes

// UserErrCode ecode
var (
	UserNotExist         = ecode.New(-404)
	UserUpdateNameFailed = ecode.New(10000)
)
