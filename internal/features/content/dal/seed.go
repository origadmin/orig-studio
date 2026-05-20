package dal

import (
	"context"
	"log/slog"

	"origadmin/application/origstudio/internal/dal/entity"
)

func SeedCategories(ctx context.Context, client *entity.Client) error {
	count, err := client.Category.Query().Count(ctx)
	if err != nil {
		return err
	}

	if count > 0 {
		slog.Info("Categories already seeded, skipping")
		return nil
	}

	categories := []struct {
		Name        string
		Slug        string
		Icon        string
		Color       string
		Sequence    int
		Description string
		NameI18n    map[string]string
		IsGlobal    bool
	}{
		{
			Name:       "音乐",
			Slug:       "music",
			Icon:       "music",
			Color:      "#E74C3C",
			Sequence:   1,
			Description: "音乐视频 - Music videos",
			NameI18n:  map[string]string{"en": "Music", "ja": "音楽"},
			IsGlobal:  true,
		},
		{
			Name:       "游戏",
			Slug:       "gaming",
			Icon:       "gamepad-2",
			Color:      "#8E44AD",
			Sequence:   2,
			Description: "游戏内容 - Gaming content",
			NameI18n:  map[string]string{"en": "Gaming", "ja": "ゲーム"},
			IsGlobal:  true,
		},
		{
			Name:       "教育",
			Slug:       "education",
			Icon:       "graduation-cap",
			Color:      "#2980B9",
			Sequence:   3,
			Description: "教育内容 - Educational content",
			NameI18n:  map[string]string{"en": "Education", "ja": "教育"},
			IsGlobal:  true,
		},
		{
			Name:       "科技",
			Slug:       "technology",
			Icon:       "cpu",
			Color:      "#16A085",
			Sequence:   4,
			Description: "科技内容 - Technology & science",
			NameI18n:  map[string]string{"en": "Technology", "ja": "テクノロジー"},
			IsGlobal:  true,
		},
		{
			Name:       "生活",
			Slug:       "lifestyle",
			Icon:       "heart",
			Color:      "#F39C12",
			Sequence:   5,
			Description: "生活方式 - Lifestyle & vlog",
			NameI18n:  map[string]string{"en": "Lifestyle", "ja": "ライフスタイル"},
			IsGlobal:  true,
		},
		{
			Name:       "体育",
			Slug:       "sports",
			Icon:       "trophy",
			Color:      "#27AE60",
			Sequence:   6,
			Description: "体育内容 - Sports content",
			NameI18n:  map[string]string{"en": "Sports", "ja": "スポーツ"},
			IsGlobal:  true,
		},
		{
			Name:       "娱乐",
			Slug:       "entertainment",
			Icon:       "sparkles",
			Color:      "#E67E22",
			Sequence:   7,
			Description: "娱乐内容 - Entertainment",
			NameI18n:  map[string]string{"en": "Entertainment", "ja": "エンターテイメント"},
			IsGlobal:  true,
		},
		{
			Name:       "纪录片",
			Slug:       "documentary",
			Icon:       "film",
			Color:      "#2C3E50",
			Sequence:   8,
			Description: "纪录片 - Documentary films",
			NameI18n:  map[string]string{"en": "Documentary", "ja": "ドキュメンタリー"},
			IsGlobal:  true,
		},
		{
			Name:       "其他",
			Slug:       "other",
			Icon:       "ellipsis",
			Color:      "#95A5A6",
			Sequence:   99,
			Description: "其他分类 - Other categories",
			NameI18n:  map[string]string{"en": "Other", "ja": "その他"},
			IsGlobal:  true,
		},
	}

	for _, c := range categories {
		create := client.Category.Create().
			SetName(c.Name).
			SetSlug(c.Slug).
			SetIcon(c.Icon).
			SetColor(c.Color).
			SetSequence(c.Sequence).
			SetIsGlobal(c.IsGlobal)

		if c.Description != "" {
			create.SetDescription(c.Description)
		}
		if c.NameI18n != nil {
			create.SetNameI18n(c.NameI18n)
		}

		if _, err := create.Save(ctx); err != nil {
			slog.Error("failed to seed category", "name", c.Name, "err", err)
			return err
		}
	}

	slog.Info("Successfully seeded categories", "count", len(categories))
	return nil
}

func SeedTags(ctx context.Context, client *entity.Client) error {
	count, err := client.Tag.Query().Count(ctx)
	if err != nil {
		return err
	}

	if count > 0 {
		slog.Info("Tags already seeded, skipping")
		return nil
	}

	tags := []struct {
		Title       string
		Slug        string
		Color       string
		Description string
		TitleI18n   map[string]string
	}{
		{
			Title:       "热门",
			Slug:        "trending",
			Color:       "#E74C3C",
			Description: "热门内容 - Trending content",
			TitleI18n:   map[string]string{"en": "Trending", "ja": "トレンド"},
		},
		{
			Title:       "推荐",
			Slug:        "recommended",
			Color:       "#3498DB",
			Description: "编辑推荐 - Editor's picks",
			TitleI18n:   map[string]string{"en": "Recommended", "ja": "おすすめ"},
		},
		{
			Title:       "原创",
			Slug:        "original",
			Color:       "#2ECC71",
			Description: "原创内容 - Original content",
			TitleI18n:   map[string]string{"en": "Original", "ja": "オリジナル"},
		},
		{
			Title:       "转载",
			Slug:        "repost",
			Color:       "#95A5A6",
			Description: "转载内容 - Reposted content",
			TitleI18n:   map[string]string{"en": "Repost", "ja": "転載"},
		},
		{
			Title:       "4K",
			Slug:        "4k",
			Color:       "#F39C12",
			Description: "4K 分辨率 - 4K resolution",
			TitleI18n:   map[string]string{"en": "4K", "ja": "4K"},
		},
		{
			Title:       "高清",
			Slug:        "hd",
			Color:       "#E67E22",
			Description: "高清画质 - HD quality",
			TitleI18n:   map[string]string{"en": "HD", "ja": "HD"},
		},
		{
			Title:       "中文字幕",
			Slug:        "zh-subtitle",
			Color:       "#1ABC9C",
			Description: "包含中文字幕 - With Chinese subtitles",
			TitleI18n:   map[string]string{"en": "Chinese Subtitles", "ja": "中国語字幕"},
		},
		{
			Title:       "英文字幕",
			Slug:        "en-subtitle",
			Color:       "#9B59B6",
			Description: "包含英文字幕 - With English subtitles",
			TitleI18n:   map[string]string{"en": "English Subtitles", "ja": "英語字幕"},
		},
		{
			Title:       "直播",
			Slug:        "live",
			Color:       "#E74C3C",
			Description: "直播内容 - Live content",
			TitleI18n:   map[string]string{"en": "Live", "ja": "ライブ"},
		},
		{
			Title:       "短视频",
			Slug:        "short",
			Color:       "#F1C40F",
			Description: "短视频 - Short videos",
			TitleI18n:   map[string]string{"en": "Short", "ja": "ショート"},
		},
	}

	for _, t := range tags {
		create := client.Tag.Create().
			SetTitle(t.Title).
			SetSlug(t.Slug)

		if t.Color != "" {
			create.SetColor(t.Color)
		}
		if t.Description != "" {
			create.SetDescription(t.Description)
		}
		if t.TitleI18n != nil {
			create.SetTitleI18n(t.TitleI18n)
		}

		if _, err := create.Save(ctx); err != nil {
			slog.Error("failed to seed tag", "title", t.Title, "err", err)
			return err
		}
	}

	slog.Info("Successfully seeded tags", "count", len(tags))
	return nil
}