package util

import (
	"context"
	"strings"

	"github.com/lindorof/gilix"
)

func ContextRet(ctx context.Context) gilix.RET {
	if ctx == nil {
		return gilix.RET_SUCCESS
	}

	err := ctx.Err()
	if err == nil {
		return gilix.RET_SUCCESS
	}
	if strings.Contains(strings.ToLower(err.Error()), "deadline") {
		return gilix.RET_TIMEOUT
	}
	if strings.Contains(strings.ToLower(err.Error()), "cancel") {
		return gilix.RET_CANCELLED
	}

	return gilix.RET_CANCELLED
}
