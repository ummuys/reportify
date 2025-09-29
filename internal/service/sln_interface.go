package service

import "context"

type RepService interface {
	CreateReport(pCtx context.Context, sql string)
}
