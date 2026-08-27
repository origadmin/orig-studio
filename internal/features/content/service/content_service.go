package service

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	contentv1 "origadmin/application/origstudio/api/gen/v1/content"
	typesv1 "origadmin/application/origstudio/api/gen/v1/types"
	"origadmin/application/origstudio/internal/features/content/biz"
	"origadmin/application/origstudio/internal/features/content/dto"
	userdto "origadmin/application/origstudio/internal/features/user/dto"
	mediadto "origadmin/application/origstudio/internal/features/media/dto"
)

type ContentService struct {
	contentv1.UnimplementedContentServiceServer
	commentUC      *biz.CommentUseCase
	likeFavoriteUC *biz.LikeFavoriteUseCase
	notificationUC *biz.NotificationUseCase
	feedUC         *biz.FeedUseCase
	categoryTagUC  *biz.CategoryTagUseCase
	portalUC       *biz.PortalUseCase
	log            *log.Helper
}

func NewContentService(
	commentUC *biz.CommentUseCase,
	likeFavoriteUC *biz.LikeFavoriteUseCase,
	notificationUC *biz.NotificationUseCase,
	feedUC *biz.FeedUseCase,
	categoryTagUC *biz.CategoryTagUseCase,
	portalUC *biz.PortalUseCase,
	logger log.Logger,
) *ContentService {
	return &ContentService{
		commentUC:      commentUC,
		likeFavoriteUC: likeFavoriteUC,
		notificationUC: notificationUC,
		feedUC:         feedUC,
		categoryTagUC:  categoryTagUC,
		portalUC:       portalUC,
		log:            log.NewHelper(log.With(logger, "module", "content.service")),
	}
}

// ==========================================
// Conversion Functions (Entity to Proto)
// ==========================================

// convertEntityUser converts UserEntityDTO to typesv1.User
func convertEntityUser(u *userdto.UserEntityDTO) *typesv1.User {
	if u == nil {
		return nil
	}
	return &typesv1.User{
		Id:       u.ID,
		Username: u.Username,
		Nickname: u.Name,
		Logo:     u.Logo,
	}
}

// convertEntityMedia converts MediaEntityDTO to typesv1.Media
func convertEntityMedia(m *mediadto.MediaEntityDTO) *typesv1.Media {
	if m == nil {
		return nil
	}
	return &typesv1.Media{
		Id:          m.ID,
		Title:       m.Title,
		Description: "",
		Thumbnail:   m.Thumbnail,
		ViewCount:   m.ViewCount,
		LikeCount:   m.LikeCount,
		Duration:    int32(m.Duration),
		Type:        string(m.Type),
		UserId:      m.UserID,
	}
}

// convertBizComment converts biz.Comment to contentv1.Comment
func convertBizComment(c *biz.Comment) *contentv1.Comment {
	if c == nil {
		return nil
	}

	id, _ := strconv.ParseInt(c.ID, 10, 64)
	mediaId, _ := strconv.ParseInt(c.MediaID, 10, 64)
	userId, _ := strconv.ParseInt(c.UserID, 10, 64)
	var parentId int64
	if c.ParentID != nil {
		parentId, _ = strconv.ParseInt(*c.ParentID, 10, 64)
	}

	replies := make([]*contentv1.Comment, len(c.Replies))
	for i, reply := range c.Replies {
		replies[i] = convertBizComment(reply)
	}

	return &contentv1.Comment{
		Id:        id,
		Uid:       c.ID,
		Text:      c.Text,
		MediaId:   mediaId,
		UserId:    userId,
		ParentId:  parentId,
		CreatedAt: timestamppb.New(c.AddDate),
		UpdatedAt: timestamppb.New(c.UpdateTime),
		User:      convertUser(c.User),
		Replies:   replies,
	}
}

// convertUser converts dto.CommentUserDTO to typesv1.User
func convertUser(u *dto.CommentUserDTO) *typesv1.User {
	if u == nil {
		return nil
	}
	return &typesv1.User{
		Id:       u.ID,
		Username: u.Username,
		Name:     u.Name,
		Avatar:   u.Avatar,
		Slug:     u.Slug,
	}
}

// convertBizNotification converts biz.Notification to contentv1.Notification
func convertBizNotification(n *biz.Notification) *contentv1.Notification {
	if n == nil {
		return nil
	}
	return &contentv1.Notification{
		Id:        int64(n.ID),
		Type:      n.Action,
		Title:     n.Title,
		Body:      n.Body,
		IsRead:    n.IsRead,
		CreatedAt: timestamppb.New(n.CreateTime),
	}
}

// convertMediaInfo converts biz.MediaInfo to typesv1.Media
func convertMediaInfo(m *biz.MediaInfo) *typesv1.Media {
	if m == nil {
		return nil
	}
	return &typesv1.Media{
		Id:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		Thumbnail:   m.Thumbnail,
		ViewCount:   m.ViewCount,
		Duration:    int32(m.Duration),
		Type:        m.Type,
	}
}

// ==========================================
// Comment Interfaces
// ==========================================

func (s *ContentService) CreateComment(ctx context.Context, req *contentv1.CreateCommentRequest) (*contentv1.CreateCommentResponse, error) {
	comment := &biz.Comment{
		Text:    req.GetText(),
		MediaID: strconv.FormatInt(req.GetMediaId(), 10),
	}
	if req.GetParentId() > 0 {
		parentId := strconv.FormatInt(req.GetParentId(), 10)
		comment.ParentID = &parentId
	}

	created, err := s.commentUC.CreateComment(ctx, comment)
	if err != nil {
		s.log.Errorf("failed to create comment: %v", err)
		return nil, err
	}

	return &contentv1.CreateCommentResponse{
		Comment: convertBizComment(created),
	}, nil
}

func (s *ContentService) GetComment(ctx context.Context, req *contentv1.GetCommentRequest) (*contentv1.GetCommentResponse, error) {
	comment, err := s.commentUC.GetComment(ctx, strconv.FormatInt(req.GetId(), 10))
	if err != nil {
		s.log.Errorf("failed to get comment: %v", err)
		return nil, err
	}

	return &contentv1.GetCommentResponse{
		Comment: convertBizComment(comment),
	}, nil
}

func (s *ContentService) UpdateComment(ctx context.Context, req *contentv1.UpdateCommentRequest) (*contentv1.UpdateCommentResponse, error) {
	comment, err := s.commentUC.UpdateComment(ctx, strconv.FormatInt(req.GetId(), 10), "", false, req.GetText())
	if err != nil {
		s.log.Errorf("failed to update comment: %v", err)
		return nil, err
	}

	return &contentv1.UpdateCommentResponse{
		Comment: convertBizComment(comment),
	}, nil
}

func (s *ContentService) DeleteComment(ctx context.Context, req *contentv1.DeleteCommentRequest) (*contentv1.DeleteCommentResponse, error) {
	err := s.commentUC.DeleteComment(ctx, strconv.FormatInt(req.GetId(), 10), "", false)
	if err != nil {
		s.log.Errorf("failed to delete comment: %v", err)
		return nil, err
	}

	return &contentv1.DeleteCommentResponse{}, nil
}

func (s *ContentService) ListComments(ctx context.Context, req *contentv1.ListCommentsRequest) (*contentv1.ListCommentsResponse, error) {
	comments, total, err := s.commentUC.ListMediaComments(ctx, strconv.FormatInt(req.GetMediaId(), 10), int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		s.log.Errorf("failed to list comments: %v", err)
		return nil, err
	}

	commentList := make([]*contentv1.Comment, len(comments))
	for i, c := range comments {
		commentList[i] = convertBizComment(c)
	}

	return &contentv1.ListCommentsResponse{
		Total:    int32(total),
		Comments: commentList,
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// ==========================================
// Like & Favorite Interfaces
// ==========================================

func (s *ContentService) ToggleLike(ctx context.Context, req *contentv1.ToggleLikeRequest) (*contentv1.ToggleLikeResponse, error) {
	if s.likeFavoriteUC == nil {
		return &contentv1.ToggleLikeResponse{
			Liked:       false,
			TotalLikes:  0,
		}, nil
	}
	// TODO: Implement with proper LikeFavoriteUseCase
	return &contentv1.ToggleLikeResponse{
		Liked:       true,
		TotalLikes:  100,
	}, nil
}

func (s *ContentService) ListLikes(ctx context.Context, req *contentv1.ListLikesRequest) (*contentv1.ListLikesResponse, error) {
	return &contentv1.ListLikesResponse{
		Total: 0,
		Users: []*typesv1.User{},
	}, nil
}

func (s *ContentService) ToggleFavorite(ctx context.Context, req *contentv1.ToggleFavoriteRequest) (*contentv1.ToggleFavoriteResponse, error) {
	if s.likeFavoriteUC == nil {
		return &contentv1.ToggleFavoriteResponse{
			Favorited:      false,
			TotalFavorites: 0,
		}, nil
	}
	// TODO: Implement with proper LikeFavoriteUseCase
	return &contentv1.ToggleFavoriteResponse{
		Favorited:      true,
		TotalFavorites: 50,
	}, nil
}

func (s *ContentService) ListFavorites(ctx context.Context, req *contentv1.ListFavoritesRequest) (*contentv1.ListFavoritesResponse, error) {
	return &contentv1.ListFavoritesResponse{
		Total:    0,
		Medias:   []*typesv1.Media{},
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// ==========================================
// Notification Interfaces
// ==========================================

func (s *ContentService) ListNotifications(ctx context.Context, req *contentv1.ListNotificationsRequest) (*contentv1.ListNotificationsResponse, error) {
	notifications, total, err := s.notificationUC.ListUserNotifications(ctx, "", int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		s.log.Errorf("failed to list notifications: %v", err)
		return nil, err
	}

	notificationList := make([]*contentv1.Notification, len(notifications))
	for i, n := range notifications {
		notificationList[i] = convertBizNotification(n)
	}

	unreadCount, _ := s.notificationUC.GetUnreadCount(ctx, "")

	return &contentv1.ListNotificationsResponse{
		Total:        int32(total),
		Notifications: notificationList,
		UnreadCount:  int32(unreadCount),
	}, nil
}

func (s *ContentService) MarkNotificationRead(ctx context.Context, req *contentv1.MarkNotificationReadRequest) (*contentv1.MarkNotificationReadResponse, error) {
	err := s.notificationUC.MarkAsRead(ctx, int(req.GetId()), "")
	if err != nil {
		s.log.Errorf("failed to mark notification as read: %v", err)
		return nil, err
	}
	return &contentv1.MarkNotificationReadResponse{}, nil
}

func (s *ContentService) MarkAllNotificationsRead(ctx context.Context, req *contentv1.MarkAllNotificationsReadRequest) (*contentv1.MarkAllNotificationsReadResponse, error) {
	err := s.notificationUC.MarkAllAsRead(ctx, "")
	if err != nil {
		s.log.Errorf("failed to mark all notifications as read: %v", err)
		return nil, err
	}
	return &contentv1.MarkAllNotificationsReadResponse{}, nil
}

// ==========================================
// Feed Interfaces
// ==========================================

func (s *ContentService) GetFeed(ctx context.Context, req *contentv1.GetFeedRequest) (*contentv1.GetFeedResponse, error) {
	medias, total, err := s.feedUC.GetHomeFeed(ctx, int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		s.log.Errorf("failed to get feed: %v", err)
		return nil, err
	}

	mediaList := make([]*typesv1.Media, len(medias))
	for i, m := range medias {
		mediaList[i] = convertMediaInfo(m)
	}

	return &contentv1.GetFeedResponse{
		Total:    int32(total),
		Items:    mediaList,
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// ==========================================
// Portal Interfaces
// ==========================================

func (s *ContentService) GetPortalHome(ctx context.Context, req *contentv1.GetPortalHomeRequest) (*contentv1.GetPortalHomeResponse, error) {
	featured, _, err := s.feedUC.ListFeatured(ctx, 1, int(req.GetLimit()))
	if err != nil {
		s.log.Errorf("failed to get featured: %v", err)
		featured = []*biz.MediaInfo{}
	}

	recommended, _, err := s.feedUC.GetHomeFeed(ctx, 1, int(req.GetLimit()))
	if err != nil {
		s.log.Errorf("failed to get recommended: %v", err)
		recommended = []*biz.MediaInfo{}
	}

	featuredList := make([]*typesv1.Media, len(featured))
	for i, m := range featured {
		featuredList[i] = convertMediaInfo(m)
	}

	recommendedList := make([]*typesv1.Media, len(recommended))
	for i, m := range recommended {
		recommendedList[i] = convertMediaInfo(m)
	}

	return &contentv1.GetPortalHomeResponse{
		Featured:        featuredList,
		Recommended:     recommendedList,
		PopularChannels: []*typesv1.Channel{},
	}, nil
}

func (s *ContentService) GetPortalTrending(ctx context.Context, req *contentv1.GetPortalTrendingRequest) (*contentv1.GetPortalTrendingResponse, error) {
	trending, _, err := s.feedUC.GetTrendingFeed(ctx, 1, int(req.GetLimit()))
	if err != nil {
		s.log.Errorf("failed to get trending: %v", err)
		trending = []*biz.MediaInfo{}
	}

	trendingList := make([]*typesv1.Media, len(trending))
	for i, m := range trending {
		trendingList[i] = convertMediaInfo(m)
	}

	return &contentv1.GetPortalTrendingResponse{
		Items: trendingList,
	}, nil
}

func (s *ContentService) GetPortalSubscriptionFeed(ctx context.Context, req *contentv1.GetPortalSubscriptionFeedRequest) (*contentv1.GetPortalSubscriptionFeedResponse, error) {
	// TODO: Implement subscription feed
	return &contentv1.GetPortalSubscriptionFeedResponse{
		Total:    0,
		Items:    []*typesv1.Media{},
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// ==========================================
// Category & Tag Interfaces
// ==========================================

func (s *ContentService) ListCategories(ctx context.Context, req *contentv1.ListCategoriesRequest) (*contentv1.ListCategoriesResponse, error) {
	return &contentv1.ListCategoriesResponse{
		Total:      0,
		Categories: []*contentv1.Category{},
	}, nil
}

func (s *ContentService) ListTags(ctx context.Context, req *contentv1.ListTagsRequest) (*contentv1.ListTagsResponse, error) {
	return &contentv1.ListTagsResponse{
		Total: 0,
		Tags:  []*contentv1.Tag{},
	}, nil
}

func (s *ContentService) ListAllNotifications(ctx context.Context, req *contentv1.ListAllNotificationsRequest) (*contentv1.ListAllNotificationsResponse, error) {
	notifications, total, err := s.notificationUC.ListUserNotifications(ctx, "", int(req.GetPage()), int(req.GetPageSize()))
	if err != nil {
		s.log.Errorf("failed to list all notifications: %v", err)
		return nil, err
	}

	notificationList := make([]*contentv1.Notification, len(notifications))
	for i, n := range notifications {
		notificationList[i] = convertBizNotification(n)
	}

	unreadCount, _ := s.notificationUC.GetUnreadCount(ctx, "")

	return &contentv1.ListAllNotificationsResponse{
		Total:         int32(total),
		Notifications: notificationList,
		UnreadCount:   int32(unreadCount),
	}, nil
}

func (s *ContentService) GetUnreadNotificationCount(ctx context.Context, req *contentv1.GetUnreadNotificationCountRequest) (*contentv1.GetUnreadNotificationCountResponse, error) {
	count, err := s.notificationUC.GetUnreadCount(ctx, "")
	if err != nil {
		s.log.Errorf("failed to get unread notification count: %v", err)
		return nil, err
	}

	return &contentv1.GetUnreadNotificationCountResponse{
		Count: int32(count),
	}, nil
}

func (s *ContentService) MarkNotificationReadPost(ctx context.Context, req *contentv1.MarkNotificationReadPostRequest) (*contentv1.MarkNotificationReadPostResponse, error) {
	id, err := strconv.ParseInt(req.GetId(), 10, 64)
	if err != nil {
		return nil, err
	}

	err = s.notificationUC.MarkAsRead(ctx, int(id), "")
	if err != nil {
		s.log.Errorf("failed to mark notification as read: %v", err)
		return nil, err
	}

	return &contentv1.MarkNotificationReadPostResponse{
		Success: true,
	}, nil
}

func (s *ContentService) MarkAllNotificationsReadPost(ctx context.Context, req *contentv1.MarkAllNotificationsReadPostRequest) (*contentv1.MarkAllNotificationsReadPostResponse, error) {
	err := s.notificationUC.MarkAllAsRead(ctx, "")
	if err != nil {
		s.log.Errorf("failed to mark all notifications as read: %v", err)
		return nil, err
	}

	return &contentv1.MarkAllNotificationsReadPostResponse{
		Success: true,
	}, nil
}

func (s *ContentService) DeleteNotification(ctx context.Context, req *contentv1.DeleteNotificationRequest) (*contentv1.DeleteNotificationResponse, error) {
	id, err := strconv.ParseInt(req.GetId(), 10, 64)
	if err != nil {
		return nil, err
	}

	err = s.notificationUC.DeleteNotification(ctx, int(id), "")
	if err != nil {
		s.log.Errorf("failed to delete notification: %v", err)
		return nil, err
	}

	return &contentv1.DeleteNotificationResponse{
		Success: true,
	}, nil
}

func (s *ContentService) GetExploreTrending(ctx context.Context, req *contentv1.GetExploreTrendingRequest) (*contentv1.GetExploreTrendingResponse, error) {
	limit := int(req.GetLimit())
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	medias, _, err := s.feedUC.GetTrendingFeed(ctx, 1, limit)
	if err != nil {
		s.log.Errorf("failed to get explore trending: %v", err)
		return nil, err
	}

	items := make([]*typesv1.Media, len(medias))
	for i, m := range medias {
		items[i] = convertMediaInfo(m)
	}

	return &contentv1.GetExploreTrendingResponse{
		Items: items,
	}, nil
}
