package usecase

import (
	"fmt"

	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

type SettingsUseCase struct {
	settingsRepo *repository.SettingsRepo
}

func NewSettingsUseCase(settingsRepo *repository.SettingsRepo) *SettingsUseCase {
	return &SettingsUseCase{settingsRepo: settingsRepo}
}

func (uc *SettingsUseCase) Get(key string) (string, error) {
	return uc.settingsRepo.Get(key)
}

func (uc *SettingsUseCase) GetAll() (map[string]interface{}, error) {
	settings, err := uc.settingsRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("خطا در دریافت تنظیمات")
	}

	boolFields := map[string]bool{"recording_enabled": true, "maintenance_mode": true, "allow_student_video": true}
	result := make(map[string]interface{})
	for k, v := range settings {
		if boolFields[k] {
			result[k] = v == "true"
		} else {
			result[k] = v
		}
	}
	return result, nil
}

func (uc *SettingsUseCase) Update(settings map[string]string) error {
	return uc.settingsRepo.SetAll(settings)
}
