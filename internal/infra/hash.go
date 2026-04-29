/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package infra

import (
	"github.com/origadmin/toolkits/crypto/hash"
	hashtypes "github.com/origadmin/toolkits/crypto/hash/types"
)

// NewHasher creates a new hasher.
func NewHasher() (hash.Crypto, error) {
	return hash.NewCrypto(hashtypes.BCRYPT)
}
