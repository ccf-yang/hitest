package repo

import (
	"context"

	"hitest/internal/core/global"
	"hitest/internal/shared/model"
)

type TaskRepo struct {
}

type ITaskRepo interface {
	Save(ctx context.Context, task *model.Task) error
	GetFirst(opts ...global.DBOption) (model.Task, error)
	Page(page, size int, opts ...global.DBOption) (int64, []model.Task, error)
	Update(ctx context.Context, task *model.Task) error

	WithByID(id string) global.DBOption
	WithResourceID(id uint) global.DBOption
	WithOperate(taskOperate string) global.DBOption
}
