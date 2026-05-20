package tagcolor

import "hash/fnv"

func ColorFromName(name string) string {
	if name == "" {
		return Palette[0]
	}
	h := fnv.New32a()
	h.Write([]byte(name))
	idx := h.Sum32() % uint32(len(Palette))
	return Palette[idx]
}

func IsValidHex(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, c := range s[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
