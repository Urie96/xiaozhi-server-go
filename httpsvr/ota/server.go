package ota

type DefaultOTAService struct {
	UpdateURL string
}

// NewDefaultOTAService 构造函数
func NewDefaultOTAService(updateURL string) *DefaultOTAService {
	return &DefaultOTAService{UpdateURL: updateURL}
}
