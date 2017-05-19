package vault

import (
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Config struct {
	ServerTransitKey  string
	UserTransitKey    string
	TransitBackend    string
	DefaultSecretPath string
	BulletinPath      string
	SlackWebhook      string
	SlackChannel      string
}

var config Config
var configLock = new(sync.RWMutex)
var LastUpdated time.Time

func GetConfig() Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func loadConfigFromVault(path string) error {
	resp, err := vaultClient.Logical().Read(path)
	if err != nil {
		return err
	} else if resp == nil {
		return errors.New("Failed to read config secret from vault")
	}

	// marshall into temp config to ensure it is valid
	temp := Config{}
	if b, err := json.Marshal(resp.Data); err == nil {
		if err := json.Unmarshal(b, &temp); err != nil {
			return err
		}
	} else {
		return err
	}

	// improperly formed slack webhooks are not allowed
	if !strings.HasPrefix(temp.SlackWebhook, "https://hooks.slack.com/services") {
		temp.SlackWebhook = ""
		temp.SlackChannel = ""
	}

	// don't waste a lock if nothing has changed
	if reflect.DeepEqual(temp, config) {
		return nil
	}

	// RWLock.Lock() will block read lock requests until it is done
	configLock.Lock()
	defer configLock.Unlock()

	config = temp
	LastUpdated = time.Now()
	log.Println("Goldfish config reloaded")

	return nil
}

func loadDevModeConfig() {
	config = Config {
		ServerTransitKey  : "goldfish",
		UserTransitKey    : "usertransit",
		TransitBackend    : "transit",
		DefaultSecretPath : "secret/",
		BulletinPath      : "secret/bulletins/",
	}
}
