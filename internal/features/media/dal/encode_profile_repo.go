package dal

import (
	"context"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/encodeprofile"
	"origadmin/application/origstudio/internal/features/media/dto"
)

type encodeProfileRepo struct {
	db *entity.Client
}

// NewEncodeProfileRepo creates a new EncodeProfile repository.
func NewEncodeProfileRepo(db *entity.Client) dto.EncodeProfileRepo {
	return &encodeProfileRepo{db: db}
}

func (r *encodeProfileRepo) ListActive(ctx context.Context) ([]*dto.EncodeProfile, error) {
	items, err := r.db.EncodeProfile.Query().
		Where(encodeprofile.IsActiveEQ(true)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.EncodeProfile, len(items))
	for i, item := range items {
		result[i] = convertEncodeProfileToDTO(item)
	}
	return result, nil
}

func (r *encodeProfileRepo) ListAll(ctx context.Context) ([]*dto.EncodeProfile, error) {
	items, err := r.db.EncodeProfile.Query().All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.EncodeProfile, len(items))
	for i, item := range items {
		result[i] = convertEncodeProfileToDTO(item)
	}
	return result, nil
}

func (r *encodeProfileRepo) Get(ctx context.Context, id int) (*dto.EncodeProfile, error) {
	item, err := r.db.EncodeProfile.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertEncodeProfileToDTO(item), nil
}

func (r *encodeProfileRepo) GetByName(ctx context.Context, name string) (*dto.EncodeProfile, error) {
	item, err := r.db.EncodeProfile.Query().Where(encodeprofile.NameEQ(name)).First(ctx)
	if err != nil {
		return nil, err
	}
	return convertEncodeProfileToDTO(item), nil
}

func (r *encodeProfileRepo) Create(
	ctx context.Context,
	profile *dto.EncodeProfile,
) (*dto.EncodeProfile, error) {
	builder := r.db.EncodeProfile.Create().
		SetName(profile.Name).
		SetDescription(profile.Description).
		SetExtension(profile.Extension).
		SetResolution(profile.Resolution).
		SetVideoCodec(profile.VideoCodec).
		SetVideoBitrate(profile.VideoBitrate).
		SetAudioCodec(profile.AudioCodec).
		SetAudioBitrate(profile.AudioBitrate).
		SetBentoParameters(profile.BentoParameters).
		SetIsActive(true)
	if profile.IsActive {
		builder = builder.SetIsActive(true)
	} else {
		builder = builder.SetIsActive(false)
	}
	item, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertEncodeProfileToDTO(item), nil
}

func (r *encodeProfileRepo) Update(
	ctx context.Context,
	profile *dto.EncodeProfile,
) (*dto.EncodeProfile, error) {
	item, err := r.db.EncodeProfile.UpdateOneID(profile.Id).
		SetName(profile.Name).
		SetDescription(profile.Description).
		SetExtension(profile.Extension).
		SetResolution(profile.Resolution).
		SetVideoCodec(profile.VideoCodec).
		SetVideoBitrate(profile.VideoBitrate).
		SetAudioCodec(profile.AudioCodec).
		SetAudioBitrate(profile.AudioBitrate).
		SetBentoParameters(profile.BentoParameters).
		SetIsActive(profile.IsActive).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return convertEncodeProfileToDTO(item), nil
}

func (r *encodeProfileRepo) Delete(ctx context.Context, id int) error {
	return r.db.EncodeProfile.DeleteOneID(id).Exec(ctx)
}

func convertEncodeProfileToDTO(m *entity.EncodeProfile) *dto.EncodeProfile {
	return &dto.EncodeProfile{
		Id:              m.ID,
		Name:            m.Name,
		Description:     m.Description,
		Extension:       m.Extension,
		Resolution:      m.Resolution,
		VideoCodec:      m.VideoCodec,
		VideoBitrate:    m.VideoBitrate,
		AudioCodec:      m.AudioCodec,
		AudioBitrate:    m.AudioBitrate,
		BentoParameters: m.BentoParameters,
		IsActive:        m.IsActive,
	}
}
