package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"origadmin/application/origstudio/internal/pkg/idutil"
)

type PortalBanner struct {
	ent.Schema
}

func (PortalBanner) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("portal_banners"),
	}
}

func (PortalBanner) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Unique().MaxLen(36).DefaultFunc(idutil.GenUUIDv7),
		field.String("title").MaxLen(255).NotEmpty(),
		field.JSON("title_i18n", map[string]string{}).Optional(),
		field.String("subtitle").MaxLen(255).Optional(),
		field.JSON("subtitle_i18n", map[string]string{}).Optional(),
		field.String("badge_text").MaxLen(64).Optional(),
		field.String("image_url").MaxLen(512).Optional(),
		field.String("image_mobile_url").MaxLen(512).Optional(),
		field.String("video_url").MaxLen(512).Optional(),
		field.String("bg_color_start").MaxLen(32).Optional(),
		field.String("bg_color_end").MaxLen(32).Optional(),
		field.Float("bg_overlay_opacity").Default(0),
		field.String("primary_btn_text").MaxLen(64).Optional(),
		field.String("primary_btn_url").MaxLen(512).Optional(),
		field.String("secondary_btn_text").MaxLen(64).Optional(),
		field.String("secondary_btn_url").MaxLen(512).Optional(),
		field.Int("sequence").Default(0),
		field.Bool("is_active").Default(true),
		field.Time("start_at").Optional(),
		field.Time("end_at").Optional(),
		field.Int("auto_slide_interval").Default(5000),
		// Banner 类型化：custom=手工配置轮播；hot_videos=按播放量聚合；
		// new_videos=按时间聚合；ad=广告位（T5 接入）。
		field.String("type").MaxLen(32).Default("custom"),
		// 视频类横幅聚合条数（custom/ad 忽略）。
		field.Int("count").Default(5),
		// 视频类横幅可选分类过滤（category 服务的一级频道 ID）。
		field.String("category_id").MaxLen(36).Optional(),
		// 显示模式：wide=宽屏（21:9 电影感，默认）；narrow=窄屏（16:9 标准）。
		field.String("display_mode").MaxLen(16).Default("wide"),
	}
}

func (PortalBanner) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("sequence"),
		index.Fields("is_active"),
		index.Fields("type"),
	}
}
