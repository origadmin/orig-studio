package validation

import "regexp"

var (
	ShortTokenRegex = regexp.MustCompile(`^[a-zA-Z0-9]{6,12}$`)
	UsernameRegex   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{2,38}$`)
	UUIDRegex       = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

func IsValidShortToken(s string) bool {
	return ShortTokenRegex.MatchString(s)
}

func IsValidUsername(s string) bool {
	return UsernameRegex.MatchString(s)
}

func IsValidUUID(s string) bool {
	return UUIDRegex.MatchString(s)
}
