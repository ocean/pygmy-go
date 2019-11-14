package library

import (
	"fmt"
	"runtime"
	"sort"

	"github.com/fubarhouse/pygmy/v1/service/dnsmasq"
	"github.com/fubarhouse/pygmy/v1/service/haproxy"
	model "github.com/fubarhouse/pygmy/v1/service/interface"
	"github.com/fubarhouse/pygmy/v1/service/mailhog"
	"github.com/fubarhouse/pygmy/v1/service/resolv"
	"github.com/fubarhouse/pygmy/v1/service/ssh/agent"
	"github.com/fubarhouse/pygmy/v1/service/ssh/key"
	"github.com/spf13/viper"
)

func Setup(c *Config) {

	viper.SetDefault("defaults", true)

	viper.SetDefault("networks", map[string][]string{
		"amazeeio-network": []string{
			"amazeeio-haproxy",
		},
	})

	var ResolvMacOS = resolv.Resolv{
		Data:   "\n# Generated by amazeeio pygmy\nnameserver 127.0.0.1\nport 6053\n",
		File:   "docker.amazee.io",
		Folder: "/etc/resolver",
		Name:   "MacOS Resolver",
	}

	var ResolvGeneric = resolv.Resolv{
		Data:   "\nnameserver 127.0.0.1 # added by amazee.io pygmy",
		File:   "resolv.conf",
		Folder: "/etc",
		Name:   "Linux Resolver",
	}

	if runtime.GOOS == "darwin" {
		viper.SetDefault("resolvers", []resolv.Resolv{
			ResolvMacOS,
		})
	} else if runtime.GOOS == "linux" {
		viper.SetDefault("resolvers", []resolv.Resolv{
			ResolvGeneric,
		})
	} else if runtime.GOOS == "windows" {
		viper.SetDefault("resolvers", []resolv.Resolv{})
	}

	e := viper.Unmarshal(&c)

	if e != nil {
		fmt.Println(e)
	}

	if c.Defaults {
		// If Services have been provided in complete or partially,
		// this will override the defaults allowing any value to
		// be changed by the user in the configuration file ~/.pygmy.yml
		if c.Services == nil {
			c.Services = make(map[string]model.Service, 6)
		}
		c.Services["amazeeio-ssh-agent-show-keys"] = getService(key.NewShower(), c.Services["amazeeio-ssh-agent-show-keys"])
		c.Services["amazeeio-ssh-agent-add-key"] = getService(key.NewAdder(c.Key), c.Services["amazeeio-ssh-agent-add-key"])
		c.Services["amazeeio-dnsmasq"] = getService(dnsmasq.New(), c.Services["amazeeio-dnsmasq"])
		c.Services["amazeeio-haproxy"] = getService(haproxy.New(), c.Services["amazeeio-haproxy"])
		c.Services["mailhog.docker.amazee.io"] = getService(mailhog.New(), c.Services["mailhog.docker.amazee.io"])
		c.Services["amazeeio-ssh-agent"] = getService(agent.New(), c.Services["amazeeio-ssh-agent"])
		c.SortedServices = make([]string, 0, len(c.Services))

		// We need services to be sortable...
		for key := range c.Services {
			c.SortedServices = append(c.SortedServices, key)
		}
	}
	sort.Strings(c.SortedServices)

}