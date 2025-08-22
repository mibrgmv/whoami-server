package keycloak

type Config struct {
	BaseURL            string `mapstructure:"base_url"`
	Realm              string `mapstructure:"realm"`
	PublicClientID     string `mapstructure:"public_client_id"`
	PublicClientSecret string `mapstructure:"public_client_secret"`
	AdminClientID      string `mapstructure:"admin_client_id"`
	AdminClientSecret  string `mapstructure:"admin_client_secret"`
}
