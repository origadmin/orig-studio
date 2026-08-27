package biz

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/enterprise/media/drm/dto"
)

type Repo interface {
	ListPolicies(ctx context.Context) ([]*dto.DrmPolicyDTO, error)
	GetPolicyByID(ctx context.Context, id string) (*dto.DrmPolicyDTO, error)
	CreatePolicy(ctx context.Context, p *dto.DrmPolicyDTO) (*dto.DrmPolicyDTO, error)
	UpdatePolicy(ctx context.Context, p *dto.DrmPolicyDTO) (*dto.DrmPolicyDTO, error)
	DeletePolicy(ctx context.Context, id string) error

	ListKeysByPolicy(ctx context.Context, policyID string) ([]*dto.DrmKeyDTO, error)
	CreateKey(ctx context.Context, k *dto.DrmKeyDTO) (*dto.DrmKeyDTO, error)
	DeleteKey(ctx context.Context, id string) error

	ListLicenses(ctx context.Context, page, pageSize int) ([]*dto.DrmLicenseDTO, int, error)
}

type UseCase struct {
	repo Repo
	log  *log.Helper
}

func NewUseCase(repo Repo, logger log.Logger) *UseCase {
	return &UseCase{
		repo: repo,
		log:  log.NewHelper(log.With(logger, "module", "enterprise/drm.biz")),
	}
}

func (uc *UseCase) ListPolicies(ctx context.Context) ([]*dto.DrmPolicyDTO, error) {
	return uc.repo.ListPolicies(ctx)
}

func (uc *UseCase) GetPolicyByID(ctx context.Context, id string) (*dto.DrmPolicyDTO, error) {
	return uc.repo.GetPolicyByID(ctx, id)
}

func (uc *UseCase) CreatePolicy(ctx context.Context, p *dto.DrmPolicyDTO) (*dto.DrmPolicyDTO, error) {
	return uc.repo.CreatePolicy(ctx, p)
}

func (uc *UseCase) UpdatePolicy(ctx context.Context, p *dto.DrmPolicyDTO) (*dto.DrmPolicyDTO, error) {
	return uc.repo.UpdatePolicy(ctx, p)
}

func (uc *UseCase) DeletePolicy(ctx context.Context, id string) error {
	return uc.repo.DeletePolicy(ctx, id)
}

func (uc *UseCase) ListKeysByPolicy(ctx context.Context, policyID string) ([]*dto.DrmKeyDTO, error) {
	return uc.repo.ListKeysByPolicy(ctx, policyID)
}

func (uc *UseCase) GenerateKey(ctx context.Context, policyID, contentID string, expiresAt time.Time) (*dto.DrmKeyDTO, error) {
	keyValue, err := generateRandomHex(16)
	if err != nil {
		return nil, err
	}
	keyID, err := generateRandomHex(16)
	if err != nil {
		return nil, err
	}
	iv, err := generateRandomHex(16)
	if err != nil {
		return nil, err
	}
	k := &dto.DrmKeyDTO{
		PolicyID:  policyID,
		ContentID: contentID,
		KeyID:     keyID,
		Iv:        iv,
		ExpiresAt: expiresAt,
	}
	_ = keyValue
	return uc.repo.CreateKey(ctx, k)
}

func (uc *UseCase) DeleteKey(ctx context.Context, id string) error {
	return uc.repo.DeleteKey(ctx, id)
}

func (uc *UseCase) ListLicenses(ctx context.Context, page, pageSize int) ([]*dto.DrmLicenseDTO, int, error) {
	return uc.repo.ListLicenses(ctx, page, pageSize)
}

func generateRandomHex(bytes int) (string, error) {
	b := make([]byte, bytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}