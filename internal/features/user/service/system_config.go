package service

import (
	"context"
	"fmt"
	"runtime"

	mediav1 "origadmin/application/origstudio/api/gen/v1/media"
	types "origadmin/application/origstudio/api/gen/v1/types"

	"github.com/origadmin/runtime/log"
)

type SystemConfigService struct {
	mediav1.UnimplementedSystemConfigServiceServer
	log *log.Helper
}

func NewSystemConfigService(logger log.Logger) *SystemConfigService {
	return &SystemConfigService{
		log: log.NewHelper(log.With(logger, "module", "user.service.config")),
	}
}

func (s *SystemConfigService) GetSettingsByCategory(
	ctx context.Context,
	req *mediav1.GetSettingsByCategoryRequest,
) (*mediav1.GetSettingsByCategoryResponse, error) {
	category := req.GetCategory()
	settings := map[string]string{}

	switch category {
	case "site":
		settings["title"] = "OrigAdmin"
		settings["description"] = "Content Management System"
		settings["logo"] = ""
		settings["favicon"] = ""
		settings["keywords"] = "cms,admin"
	case "security":
		settings["passwordMinLength"] = "6"
		settings["requireSpecialChar"] = "false"
		settings["sessionTimeout"] = "3600"
		settings["twoFactorAuthEnabled"] = "false"
	case "email":
		settings["enabled"] = "false"
		settings["smtpHost"] = ""
		settings["smtpPort"] = "587"
		settings["smtpUser"] = ""
		settings["fromAddress"] = "noreply@example.com"
		settings["fromName"] = "OrigAdmin"
	case "storage":
		settings["provider"] = "local"
		settings["maxSize"] = "10485760"
		settings["allowedTypes"] = "image/jpeg,image/png,image/gif,video/mp4,application/pdf"
	default:
	}

	return &mediav1.GetSettingsByCategoryResponse{
		Category: category,
		Settings: settings,
	}, nil
}

func (s *SystemConfigService) GetSettingByKey(
	ctx context.Context,
	req *mediav1.GetSettingByKeyRequest,
) (*mediav1.GetSettingByKeyResponse, error) {
	return &mediav1.GetSettingByKeyResponse{
		Key:   req.GetKey(),
		Value: "",
	}, nil
}

func (s *SystemConfigService) UpdateSettingByKey(
	ctx context.Context,
	req *mediav1.UpdateSettingByKeyRequest,
) (*mediav1.UpdateSettingByKeyResponse, error) {
	return &mediav1.UpdateSettingByKeyResponse{
		Key:   req.GetKey(),
		Value: req.GetValue(),
	}, nil
}

func (s *SystemConfigService) DeleteSettingByKey(
	ctx context.Context,
	req *mediav1.DeleteSettingByKeyRequest,
) (*mediav1.DeleteSettingByKeyResponse, error) {
	return &mediav1.DeleteSettingByKeyResponse{}, nil
}

func (s *SystemConfigService) GetSystemSetting(
	ctx context.Context,
	req *mediav1.GetSystemSettingRequest,
) (*mediav1.GetSystemSettingResponse, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	key := req.GetKey()
	var value string

	switch key {
	case "version":
		value = "1.0.0"
	case "goVersion":
		value = runtime.Version()
	case "database":
		value = "SQLite"
	case "os":
		value = runtime.GOOS + "/" + runtime.GOARCH
	case "totalMemory":
		value = fmt.Sprintf("%d", m.Sys)
	case "usedMemory":
		value = fmt.Sprintf("%d", m.Alloc)
	case "numCPU":
		value = fmt.Sprintf("%d", runtime.NumCPU())
	case "numGoroutine":
		value = fmt.Sprintf("%d", runtime.NumGoroutine())
	default:
		value = ""
	}

	return &mediav1.GetSystemSettingResponse{
		Key:   key,
		Value: value,
	}, nil
}

func (s *SystemConfigService) ResetSystemSetting(
	ctx context.Context,
	req *mediav1.ResetSystemSettingRequest,
) (*mediav1.ResetSystemSettingResponse, error) {
	return &mediav1.ResetSystemSettingResponse{Key: req.GetKey()}, nil
}

func (s *SystemConfigService) GetEmailStatus(
	ctx context.Context,
	req *mediav1.GetEmailStatusRequest,
) (*mediav1.GetEmailStatusResponse, error) {
	return &mediav1.GetEmailStatusResponse{Configured: false}, nil
}

func (s *SystemConfigService) TestEmail(
	ctx context.Context,
	req *mediav1.TestEmailRequest,
) (*mediav1.TestEmailResponse, error) {
	return &mediav1.TestEmailResponse{Success: false, Message: "email not configured"}, nil
}

func (s *SystemConfigService) GetChannelLimits(
	ctx context.Context,
	req *mediav1.GetChannelLimitsRequest,
) (*mediav1.GetChannelLimitsResponse, error) {
	// Return sensible defaults; ChannelService gRPC-gateway route should override this.
	// -1 max_channels = unlimited
	return &mediav1.GetChannelLimitsResponse{
		Limits: &types.ChannelLimits{
			MaxChannels:  -1,
			CurrentCount: 0,
			CanCreate:    true,
		},
	}, nil
}
