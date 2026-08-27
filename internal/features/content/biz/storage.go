/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package biz

import (
	"fmt"
	"time"

	origadmin_conf "origadmin/application/origstudio/internal/conf"
)

// AssetStorage is the stable contract for business asset path generation
// (avatars, ads, banners, covers, etc. — NOT content_media records).
// Callers depend on this interface, never on concrete implementations.
type AssetStorage interface {
	// Avatar returns the relative key for a user avatar.
	Avatar(userID string, t time.Time, hash string) string
	// Ad returns the relative key for an ad image.
	Ad(t time.Time, hash string) string
	// Banner returns the relative key for a portal banner.
	Banner(t time.Time, hash string) string
	// MediaCover returns the relative key for a custom media cover.
	MediaCover(uuid string) string
	// Category returns the relative key for a category thumbnail.
	Category(categoryID, hash string) string
	// Channel returns the relative key for a channel avatar or banner.
	Channel(channelID, subType, hash string) string
	// Article returns the relative key for an article inline image or thumbnail.
	Article(t time.Time, hash string) string
	// Misc returns the relative key for other business images.
	Misc(t time.Time, hash string) string
}

// GalleryStorage is the stable contract for gallery (multi-image) content paths.
type GalleryStorage interface {
	// ImagePath returns the relative key for a gallery image by index.
	ImagePath(uuid string, idx int, isThumb bool) string
	// CoverPath returns the relative key for a gallery cover image.
	CoverPath(uuid string) string
}

// assetStorageV1 is the default implementation for business assets.
type assetStorageV1 struct {
	sp *origadmin_conf.StoragePaths
}

// galleryStorageV1 is the default implementation for gallery content.
type galleryStorageV1 struct {
	sp *origadmin_conf.StoragePaths
}

// NewAssetStorage creates an AssetStorage implementation.
// strategy selects the implementation version ("v1" default).
func NewAssetStorage(sp *origadmin_conf.StoragePaths, strategy string) AssetStorage {
	switch strategy {
	// case "v2": return &assetStorageV2{sp: sp}  // future
	default:
		return &assetStorageV1{sp: sp}
	}
}

// NewGalleryStorage creates a GalleryStorage implementation.
// strategy selects the implementation version ("v1" default, "v2" future hash sharding).
func NewGalleryStorage(sp *origadmin_conf.StoragePaths, strategy string) GalleryStorage {
	switch strategy {
	// case "v2": return &galleryStorageV2{sp: sp}  // future
	default:
		return &galleryStorageV1{sp: sp}
	}
}

func (a *assetStorageV1) Avatar(userID string, t time.Time, hash string) string {
	return a.sp.Relative("assets/avatars", userID,
		fmt.Sprintf("%04d%02d", t.Year(), t.Month()), hash+".jpg")
}

func (a *assetStorageV1) Ad(t time.Time, hash string) string {
	return a.sp.Relative("assets/ads",
		fmt.Sprintf("%04d%02d", t.Year(), t.Month()), hash+".jpg")
}

func (a *assetStorageV1) Banner(t time.Time, hash string) string {
	return a.sp.Relative("assets/banners",
		fmt.Sprintf("%04d%02d", t.Year(), t.Month()), hash+".jpg")
}

func (a *assetStorageV1) MediaCover(uuid string) string {
	return a.sp.Relative("assets/covers", uuid+".jpg")
}

func (a *assetStorageV1) Category(categoryID, hash string) string {
	return a.sp.Relative("assets/categories", categoryID, hash+".jpg")
}

func (a *assetStorageV1) Channel(channelID, subType, hash string) string {
	return a.sp.Relative("assets/channels", channelID, subType, hash+".jpg")
}

func (a *assetStorageV1) Article(t time.Time, hash string) string {
	return a.sp.Relative("assets/articles",
		fmt.Sprintf("%04d%02d", t.Year(), t.Month()), hash+".jpg")
}

func (a *assetStorageV1) Misc(t time.Time, hash string) string {
	return a.sp.Relative("assets/misc",
		fmt.Sprintf("%04d%02d", t.Year(), t.Month()), hash+".jpg")
}

func (g *galleryStorageV1) ImagePath(uuid string, idx int, isThumb bool) string {
	suffix := ""
	if isThumb {
		suffix = "_thumb"
	}
	return g.sp.Relative("gallery", uuid, fmt.Sprintf("%02d%s.jpg", idx, suffix))
}

func (g *galleryStorageV1) CoverPath(uuid string) string {
	return g.sp.Relative("gallery", uuid, "cover.jpg")
}

// RegisterAssetPaths registers all business asset storage directories.
// Must be called at startup after NewStoragePaths.
func RegisterAssetPaths(sp *origadmin_conf.StoragePaths) error {
	subs := []string{"avatars", "ads", "banners", "covers", "channels", "categories", "articles", "misc"}
	for _, sub := range subs {
		name := "assets/" + sub
		if err := sp.Register(origadmin_conf.PathSpec{
			Name:        name,
			DefaultDir:  name,
			Description: "business asset: " + sub,
		}); err != nil {
			return err
		}
	}
	return nil
}

// RegisterGalleryPaths registers the gallery storage directory.
// Must be called at startup after NewStoragePaths.
func RegisterGalleryPaths(sp *origadmin_conf.StoragePaths) error {
	return sp.Register(origadmin_conf.PathSpec{
		Name:        "gallery",
		DefaultDir:  "gallery",
		Description: "gallery content (multi-image)",
	})
}
