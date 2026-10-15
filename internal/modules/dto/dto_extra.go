package dto

// 移植自 legacy common_req.go / setting.go（只留各模块用到的）

type OperationWithName struct {
	Name string `json:"name" validate:"required"`
}

type UpdateByNameAndFile struct {
	Name string `json:"name" validate:"required"`
	File string `json:"file"`
}

type SettingUpdate struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
}
