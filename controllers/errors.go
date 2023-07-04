package controllers

import (
	minErros "errors"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsAlreadyExistsError 检查错误是否由于资源已存在引起的
func IsAlreadyExistsError(err error) bool {
	return errors.IsAlreadyExists(err)
}

// IsNotFoundError 检查错误是否由于资源未找到引起的
func IsNotFoundError(err error) bool {
	return errors.IsNotFound(err)
}

// IsConflictError 检查错误是否由于资源冲突引起的
func IsConflictError(err error) bool {
	return errors.IsConflict(err)
}

// IsInvalidError 检查错误是否由于无效请求引起的
func IsInvalidError(err error) bool {
	return errors.IsInvalid(err)
}

// IsMethodNotSupportedError 检查错误是否由于请求方法不支持引起的
func IsMethodNotSupportedError(err error) bool {
	return errors.IsMethodNotSupported(err)
}

// IsBadRequestError 检查错误是否由于请求不符合 API 要求引起的
func IsBadRequestError(err error) bool {
	return errors.IsBadRequest(err)
}

// IsServiceUnavailableError 检查错误是否由于 API 服务器不可用引起的
func IsServiceUnavailableError(err error) bool {
	return errors.IsServiceUnavailable(err)
}

// IsUnauthorizedError 检查错误是否由于未授权的请求引起的
func IsUnauthorizedError(err error) bool {
	return errors.IsUnauthorized(err)
}

// IsForbiddenError 检查错误是否由于被禁止的请求引起的
func IsForbiddenError(err error) bool {
	return errors.IsForbidden(err)
}

// IsTimeoutError 检查错误是否由于请求超时引起的
func IsTimeoutError(err error) bool {
	return errors.IsTimeout(err)
}

// IsInternalError 检查错误是否由于 API 服务器内部错误引起的
func IsInternalError(err error) bool {
	return errors.IsInternalError(err)
}

// IsNodePortAllocationError 检查错误是否由于 NodePort 分配失败引起的
func IsNodePortAllocationError(err error) bool {
	if errors.IsInvalid(err) {
		var status *errors.StatusError
		if minErros.As(err, &status) {
			for _, cause := range status.ErrStatus.Details.Causes {
				if cause.Type == metav1.CauseTypeFieldValueInvalid && cause.Field == "spec.ports[0].nodePort" {
					// 如果错误是由 NodePort 分配失败引起的，则返回 true
					return true
				}
			}
		}
	}

	// 否则，返回 false
	return false
}
