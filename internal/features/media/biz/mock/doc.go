/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package mock provides gomock-generated mocks for media biz and dto interfaces.
// Regenerate mocks with: go generate ./internal/features/media/biz/mock/
package mock

//go:generate mockgen -destination=storage_mock.go -package=mock origadmin/application/origstudio/internal/features/media/biz Storage
//go:generate mockgen -destination=transcode_worker_mock.go -package=mock origadmin/application/origstudio/internal/features/media/biz TranscodeWorker
//go:generate mockgen -destination=media_repo_mock.go -package=mock origadmin/application/origstudio/internal/features/media/dto MediaRepo
//go:generate mockgen -destination=upload_repo_mock.go -package=mock origadmin/application/origstudio/internal/features/media/dto UploadRepo
//go:generate mockgen -destination=encoding_task_repo_mock.go -package=mock origadmin/application/origstudio/internal/features/media/dto EncodingTaskRepo
//go:generate mockgen -destination=encode_profile_repo_mock.go -package=mock origadmin/application/origstudio/internal/features/media/dto EncodeProfileRepo
