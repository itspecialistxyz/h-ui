package service

import (
	"errors"
	"fmt"
	"h-ui/dao"
	"h-ui/model/bo"
	"h-ui/model/constant"
	"h-ui/model/entity"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func UpdateConfig(key string, value string) error {
	if key == constant.Hysteria2Enable {
		if value == "1" {
			hysteria2Config, err := GetHysteria2Config()
			if err != nil {
				return err
			}
			if hysteria2Config.Listen == nil || *hysteria2Config.Listen == "" {
				logrus.Errorf("hysteria2 config is empty")
				return errors.New("hysteria2 config is empty")
			}
			// 启动Hysteria2
			if err = StartHysteria2(); err != nil {
				return err
			}
		} else {
			if err := StopHysteria2(); err != nil {
				return err
			}
		}
	}
	return dao.UpdateConfig([]string{key}, map[string]interface{}{"value": value})
}

func GetConfig(key string) (entity.Config, error) {
	return dao.GetConfig("key = ?", key)
}

func ListConfig(keys []string) ([]entity.Config, error) {
	return dao.ListConfig("key in ?", keys)
}

func ListConfigNotIn(keys []string) ([]entity.Config, error) {
	return dao.ListConfig("key not in ?", keys)
}

func GetHysteria2Config() (bo.Hysteria2ServerConfig, error) {
	var serverConfig bo.Hysteria2ServerConfig
	config, err := dao.GetConfig("key = ?", constant.Hysteria2Config)
	if err != nil {
		return serverConfig, err
	}
	if config.Value == nil || *config.Value == "" {
		// Return empty config if not set, to avoid unmarshal errors on empty string
		return bo.Hysteria2ServerConfig{}, nil
	}
	if err = yaml.Unmarshal([]byte(*config.Value), &serverConfig); err != nil {
		return serverConfig, err
	}
	return serverConfig, nil
}

func UpdateHysteria2Config(hysteria2ServerConfig bo.Hysteria2ServerConfig) error {
	// 默认值
	config, err := dao.ListConfig("key in ?", []string{constant.HUIWebPort, constant.JwtSecret})
	if err != nil {
		return err
	}

	var hUIWebPort string
	var jwtSecret string
	for _, item := range config {
		if *item.Key == constant.HUIWebPort {
			hUIWebPort = *item.Value
		} else if *item.Key == constant.JwtSecret {
			jwtSecret = *item.Value
		}
	}

	if hUIWebPort == "" || jwtSecret == "" {
		logrus.Errorf("hUIWebPort or jwtSecret is nil")
		return errors.New(constant.SysError)
	}

	authHttpUrl, err := GetAuthHttpUrl()
	if err != nil {
		return err
	}

	crtPath, keyPath, err := getHUIKeyAndCrtPath()
	if err != nil {
		logrus.Errorf("Failed to get HUI cert paths for Hysteria2 auth config: %v", err)
		return errors.New("failed to get HUI cert paths")
	}

	authHttpInsecure := true // Default to insecure if not using TLS
	if crtPath != "" && keyPath != "" {
		authHttpInsecure = false // H-UI is using TLS, so callback can be secure
	}

	authType := "http"
	var auth bo.ServerConfigAuth
	auth.Type = &authType
	var httpAuth bo.ServerConfigAuthHTTP // Renamed to avoid conflict with http package
	httpAuth.URL = &authHttpUrl
	httpAuth.Insecure = &authHttpInsecure
	auth.HTTP = &httpAuth
	hysteria2ServerConfig.Auth = &auth
	hysteria2ServerConfig.TrafficStats.Secret = &jwtSecret

	yamlConfig, err := yaml.Marshal(&hysteria2ServerConfig)
	if err != nil {
		return err
	}
	return dao.UpdateConfig([]string{constant.Hysteria2Config}, map[string]interface{}{"value": string(yamlConfig)})
}

func SetHysteria2Config(hysteria2ServerConfig bo.Hysteria2ServerConfig) error {
	config, err := yaml.Marshal(&hysteria2ServerConfig)
	if err != nil {
		return err
	}
	return dao.UpdateConfig([]string{constant.Hysteria2Config}, map[string]interface{}{"value": string(config)})
}

func UpsertConfig(configs []entity.Config) error {
	return dao.UpsertConfig(configs)
}

func GetHysteria2ApiPort() (int64, error) {
	hysteria2Config, err := GetHysteria2Config()
	if err != nil {
		return 0, err
	}
	if hysteria2Config.TrafficStats == nil || hysteria2Config.TrafficStats.Listen == nil {
		errMsg := "hysteria2 Traffic Stats API (HTTP) Listen is nil"
		logrus.Errorf(errMsg)
		return 0, errors.New(errMsg)
	}
	apiPort, err := strconv.ParseInt(strings.Split(*hysteria2Config.TrafficStats.Listen, ":")[1], 10, 64)
	if err != nil {
		errMsg := fmt.Sprintf("apiPort: %s is invalid", *hysteria2Config.TrafficStats.Listen)
		logrus.Errorf(errMsg)
		return 0, errors.New(errMsg)
	}
	return apiPort, nil
}

func GetPortAndCert() (int64, string, string, error) {
	configs, err := dao.ListConfig("key in ?", []string{constant.HUIWebPort, constant.HUICrtPath, constant.HUIKeyPath})
	if err != nil {
		return 0, "", "", err
	}
	port := ""
	crtPath := ""
	keyPath := ""
	for _, config := range configs {
		if config.Value == nil || config.Key == nil {
			continue
		}
		value := *config.Value
		if *config.Key == constant.HUIWebPort {
			port = value
		} else if *config.Key == constant.HUICrtPath {
			crtPath = value
		} else if *config.Key == constant.HUIKeyPath {
			keyPath = value
		}
	}

	if port == "" { // Ensure port has a default or error if critical
		logrus.Warnf("HUIWebPort is not set, using default or potentially failing if required.")
		// Consider returning an error or using a default if appropriate
		// For now, let it attempt parse and fail if empty, as per original logic
	}

	portInt, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		logrus.Errorf("port: '%s' is invalid: %v", port, err)
		return 0, "", "", fmt.Errorf("port: '%s' is invalid: %w", port, err)
	}

	return portInt, crtPath, keyPath, nil
}

func getHUIKeyAndCrtPath() (string, string, error) {
	configs, err := dao.ListConfig("key IN (?, ?)", constant.HUICrtPath, constant.HUIKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to query cert paths: %w", err)
	}
	var crtPath, keyPath string
	for _, c := range configs {
		if c.Key != nil && c.Value != nil {
			if *c.Key == constant.HUICrtPath {
				crtPath = *c.Value
			} else if *c.Key == constant.HUIKeyPath {
				keyPath = *c.Value
			}
		}
	}
	return crtPath, keyPath, nil
}

func GetAuthHttpUrl() (string, error) {
	port, crtPath, keyPath, err := GetPortAndCert()
	if err != nil {
		return "", err
	}
	protocol := "http"
	if crtPath != "" && keyPath != "" {
		protocol = "https"
	}
	return fmt.Sprintf("%s://127.0.0.1:%d/hui/hysteria2/auth", protocol, port), nil
}
