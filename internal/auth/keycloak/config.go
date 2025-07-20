package keycloak

type Config struct {
	BaseURL           string `mapstructure:"base_url"`
	Realm             string `mapstructure:"realm"`
	ClientID          string `mapstructure:"client_id"`
	ClientSecret      string `mapstructure:"client_secret"`
	AdminClientID     string `mapstructure:"admin_client_id"`
	AdminClientSecret string `mapstructure:"admin_client_secret"`
}
