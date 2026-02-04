package config

type server struct {
	Host  string `env:"HOST" default:"0.0.0.0"`
	Port  int    `env:"PORT" default:"3000"`
	Debug bool   `env:"DEBUG" default:"false"`
}

type database struct {
	Driver string `env:"DB_DRIVER" default:"sqlite"`
	DSN    string `env:"DSN" default:"metachan.db"`
}

type sync struct {
	AniSync bool `env:"ANISYNC" default:"false"`
}

type api struct {
	TMDBKey       string `env:"TMDB_API_KEY" default:""`
	TMDBReadToken string `env:"TMDB_READ_ACCESS_TOKEN" default:""`
	TVDBKey       string `env:"TVDB_API_KEY" default:""`
}
