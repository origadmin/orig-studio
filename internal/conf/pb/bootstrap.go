package confpb

import (
	"encoding/json"

	datav1 "github.com/origadmin/runtime/api/gen/go/config/data/v1"
	transportv1 "github.com/origadmin/runtime/api/gen/go/config/transport/v1"
)

type Bootstrap struct {
	Servers  *transportv1.Servers `json:"servers,omitempty"`
	Data     *datav1.Data         `json:"data,omitempty"`
	Clients  *transportv1.Clients `json:"clients,omitempty"`
	Security *Security            `json:"security,omitempty"`
}

func (b *Bootstrap) GetServers() *transportv1.Servers {
	if b == nil {
		return nil
	}
	return b.Servers
}

func (b *Bootstrap) GetData() *datav1.Data {
	if b == nil {
		return nil
	}
	return b.Data
}

func (b *Bootstrap) GetClients() *transportv1.Clients {
	if b == nil {
		return nil
	}
	return b.Clients
}

func (b *Bootstrap) GetSecurity() *Security {
	if b == nil {
		return nil
	}
	return b.Security
}

type Security struct {
	Authn []*AuthnConfig `json:"authn,omitempty"`
}

func (s *Security) GetAuthn() []*AuthnConfig {
	if s == nil {
		return nil
	}
	return s.Authn
}

type AuthnConfig struct {
	Name string     `json:"name,omitempty"`
	Type string     `json:"type,omitempty"`
	JWT  *JWTConfig `json:"jwt,omitempty"`
}

func (a *AuthnConfig) GetName() string {
	if a == nil {
		return ""
	}
	return a.Name
}

func (a *AuthnConfig) GetType() string {
	if a == nil {
		return ""
	}
	return a.Type
}

func (a *AuthnConfig) GetJWT() *JWTConfig {
	if a == nil {
		return nil
	}
	return a.JWT
}

type JWTConfig struct {
	Secret     string `json:"secret,omitempty"`
	AccessTTL  int64  `json:"access_ttl,omitempty"`
	RefreshTTL int64  `json:"refresh_ttl,omitempty"`
	Issuer     string `json:"issuer,omitempty"`
}

func (j *JWTConfig) GetSecret() string {
	if j == nil {
		return ""
	}
	return j.Secret
}

func (j *JWTConfig) GetAccessTTL() int64 {
	if j == nil {
		return 0
	}
	return j.AccessTTL
}

func (j *JWTConfig) GetRefreshTTL() int64 {
	if j == nil {
		return 0
	}
	return j.RefreshTTL
}

func (j *JWTConfig) GetIssuer() string {
	if j == nil {
		return ""
	}
	return j.Issuer
}

func (b *Bootstrap) UnmarshalJSON(data []byte) error {
	type Alias Bootstrap
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	return nil
}
