export const ReviewStatus = {
  UNKNOWN: 'unknown',
  PENDING_REVIEW: 'pending_review',
  REVIEWED: 'reviewed',
  REJECTED: 'rejected',
} as const;
export type ReviewStatus = typeof ReviewStatus[keyof typeof ReviewStatus];

export const MediaState = {
  UNKNOWN: 'unknown',
  DRAFT: 'draft',
  ACTIVE: 'active',
  DELETED: 'deleted',
} as const;
export type MediaState = typeof MediaState[keyof typeof MediaState];

export const CommentStatus = {
  UNKNOWN: 'unknown',
  PENDING: 'pending',
  APPROVED: 'approved',
  REJECTED: 'rejected',
  REPORTED: 'reported',
} as const;
export type CommentStatus = typeof CommentStatus[keyof typeof CommentStatus];

export const EncodingTaskStatus = {
  UNKNOWN: 'unknown',
  PENDING: 'pending',
  PROCESSING: 'processing',
  SUCCESS: 'success',
  FAILED: 'failed',
  SKIPPED: 'skipped',
  PARTIAL: 'partial',
} as const;
export type EncodingTaskStatus = typeof EncodingTaskStatus[keyof typeof EncodingTaskStatus];

export const MediaEncodingStatus = {
  UNKNOWN: 'unknown',
  PENDING: 'pending',
  PROCESSING: 'processing',
  SUCCESS: 'success',
  FAILED: 'failed',
  PARTIAL: 'partial',
} as const;
export type MediaEncodingStatus = typeof MediaEncodingStatus[keyof typeof MediaEncodingStatus];

export const UploadStatus = {
  UNKNOWN: 'unknown',
  PENDING: 'pending',
  UPLOADING: 'uploading',
  COMPLETED: 'completed',
  ABORTED: 'aborted',
} as const;
export type UploadStatus = typeof UploadStatus[keyof typeof UploadStatus];

export const SpriteStatus = {
  UNKNOWN: 'unknown',
  PENDING: 'pending',
  PROCESSING: 'processing',
  COMPLETED: 'completed',
  FAILED: 'failed',
} as const;
export type SpriteStatus = typeof SpriteStatus[keyof typeof SpriteStatus];

export const UserStatus = {
  UNKNOWN: 'unknown',
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const;
export type UserStatus = typeof UserStatus[keyof typeof UserStatus];

export const TagStatus = {
  UNKNOWN: 'unknown',
  ACTIVE: 'active',
  INACTIVE: 'inactive',
} as const;
export type TagStatus = typeof TagStatus[keyof typeof TagStatus];

export const ChannelStatus = {
  UNKNOWN: 'unknown',
  ACTIVE: 'active',
  PENDING: 'pending',
  INACTIVE: 'inactive',
} as const;
export type ChannelStatus = typeof ChannelStatus[keyof typeof ChannelStatus];
