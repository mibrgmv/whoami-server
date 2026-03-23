package keycloak

type Config struct {
	BaseURL      string `mapstructure:"base_url"`
	IssuerURL    string `mapstructure:"issuer_url"`
	Realm        string `mapstructure:"realm"`
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
}
