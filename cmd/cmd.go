package main

import (
	"github.com/spf13/viper"
	"log"
	httpup "roob.re/prosody-httpupload"
	"strings"
)

func main() {
	v := viper.New()
	v.SetDefault("listen-address", ":8889")
	v.SetDefault("storage-path", "data")
	v.SetDefault("secret", "")

	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.SetEnvPrefix("HTTPUP")
	v.AutomaticEnv()

	serverConfig := httpup.Config{}
	err := v.Unmarshal(&serverConfig)
	if err != nil {
		log.Fatalf("error reading config: %v", err)
	}

	server, err := httpup.New(serverConfig)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Run()
	if err != nil {
		log.Fatal(err)
	}
}
