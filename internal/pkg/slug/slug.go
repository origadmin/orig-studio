package slug

import (
	"math/rand"
	"origadmin/application/origstudio/internal/pkg/hashtag"
	"regexp"
	"strings"
)

const (
	MaxSlugLength    = 100
	DefaultSuffixLen = 3
	FallbackSlug     = "item"
)

var (
	validSlugRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*[a-z0-9]$`)
	reservedSlugs  = map[string]bool{
		"admin": true, "support": true, "help": true, "api": true,
		"www": true, "mail": true, "ftp": true, "root": true,
		"system": true, "moderator": true, "mod": true, "staff": true,
		"user": true, "users": true, "channel": true, "channels": true,
		"media": true, "video": true, "playlist": true, "category": true,
		"tag": true, "tags": true, "article": true, "articles": true,
		"search": true, "explore": true, "settings": true, "auth": true,
		"login": true, "register": true, "signup": true, "signin": true,
		"signout": true, "me": true, "notifications": true, "comments": true,
	}
)

type ExistsFunc func(slug string) (bool, error)

func Generate(name string) string {
	return generateWithSuffix(name, DefaultSuffixLen)
}

func GenerateWithoutSuffix(name string) string {
	base := hashtag.GenerateTagSlug(name)
	if base == hashtag.FallbackSlug {
		base = FallbackSlug
	}
	if len(base) > MaxSlugLength {
		base = base[:MaxSlugLength]
	}
	return base
}

func GenerateUserSlug(nickname, username string) string {
	if nickname != "" {
		slug := hashtag.GenerateTagSlug(nickname)
		if slug != "" && len(slug) >= 3 && !IsReserved(slug) {
			return slug
		}
	}
	if username != "" {
		return hashtag.GenerateTagSlug(username)
	}
	return FallbackSlug
}

func EnsureUnique(baseSlug string, existsFn ExistsFunc) (string, error) {
	slug := baseSlug
	for attempt := 0; attempt < 10; attempt++ {
		taken, err := existsFn(slug)
		if err != nil {
			return "", err
		}
		if !taken {
			return slug, nil
		}
		slug = baseSlug + "-" + randomSuffix(DefaultSuffixLen)
	}
	return baseSlug + "-" + randomSuffix(DefaultSuffixLen), nil
}

func IsValid(slug string) bool {
	if len(slug) < 3 || len(slug) > 64 {
		return false
	}
	return validSlugRegex.MatchString(slug)
}

func IsReserved(slug string) bool {
	return reservedSlugs[strings.ToLower(slug)]
}

func generateWithSuffix(name string, suffixLen int) string {
	base := hashtag.GenerateTagSlug(name)
	if base == hashtag.FallbackSlug {
		base = FallbackSlug
	}
	if len(base) > MaxSlugLength-suffixLen-1 {
		base = base[:MaxSlugLength-suffixLen-1]
	}
	return base + "-" + randomSuffix(suffixLen)
}

func randomSuffix(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
