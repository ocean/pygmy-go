package library

import (
	"fmt"
	"runtime"
	"sort"
	"strings"

	"github.com/fubarhouse/pygmy-go/service/dnsmasq"
	"github.com/fubarhouse/pygmy-go/service/haproxy"
	model "github.com/fubarhouse/pygmy-go/service/interface"
	"github.com/fubarhouse/pygmy-go/service/mailhog"
	"github.com/fubarhouse/pygmy-go/service/network"
	"github.com/fubarhouse/pygmy-go/service/resolv"
	"github.com/fubarhouse/pygmy-go/service/ssh/agent"
	"github.com/fubarhouse/pygmy-go/service/ssh/key"
	"github.com/spf13/viper"
)

func Setup(c *Config) {

	viper.SetDefault("defaults", true)

	var ResolvMacOS = resolv.Resolv{
		Data:   "# Generated by amazeeio pygmy\nnameserver 127.0.0.1\nport 6053\n",
		File:   "docker.amazee.io",
		Folder: "/etc/resolver",
		Name:   "MacOS Resolver",
	}

	var ResolvGeneric = resolv.Resolv{
		Data:   "nameserver 127.0.0.1 # added by amazee.io pygmy",
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
		c.Services["amazeeio-ssh-agent-add-key"] = getService(key.NewAdder(), c.Services["amazeeio-ssh-agent-add-key"])
		c.Services["amazeeio-dnsmasq"] = getService(dnsmasq.New(), c.Services["amazeeio-dnsmasq"])
		c.Services["amazeeio-haproxy"] = getService(haproxy.New(), c.Services["amazeeio-haproxy"])
		c.Services["mailhog.docker.amazee.io"] = getService(mailhog.New(), c.Services["mailhog.docker.amazee.io"])
		c.Services["amazeeio-ssh-agent"] = getService(agent.New(), c.Services["amazeeio-ssh-agent"])

		// We need Port 80 to be configured by default.
		// If a port on amazeeio-haproxy isn't explicitly declared,
		// then we should set this value. This is far more creative
		// than needed, so feel free to revisit if you can compile it.
		if c.Services["amazeeio-haproxy"].HostConfig.PortBindings == nil {
			c.Services["amazeeio-haproxy"] = getService(haproxy.NewDefaultPorts(), c.Services["amazeeio-haproxy"])
		}

		// It's sensible to use the same logic for port 1025.
		// If a user needs to configure it, the default value should not be set also.
		if c.Services["mailhog.docker.amazee.io"].HostConfig.PortBindings == nil {
			c.Services["mailhog.docker.amazee.io"] = getService(mailhog.NewDefaultPorts(), c.Services["mailhog.docker.amazee.io"])
		}

		// If networks are not provided, we should provide defaults.
		// Defaults will be provided if nothing is found in configuration is
		// completely absent.
		viper.SetDefault("networks", map[string]model.Network{
			"amazeeio-network": network.New(),
		})

	}

	// It is because of interdependent containers we introduce a weighting system.
	// Containers are sorted alphabetically, and then sorted based on their weight.
	// Note that for now, the sorting is handled as a string. So weight identifiers
	// Should have an identical length.
	{
		length := 0
		c.SortedServices = make([]string, 0, len(c.Services))
		for key, value := range c.Services {
			c.SortedServices = append(c.SortedServices, fmt.Sprintf("%v|%v", value.Weight, key))
			// If the length is changed mid-flight, we should at least warn against unexpected container ordering.
			if len(string(value.Weight)) > length && length != 0 {
				fmt.Printf("Warning: please check the Weight attribute of the %v container configuration, ordering may not work correctly.\n", value.Name)
			}
			// Increment the length check as needed.
			if len(string(value.Weight)) > length {
				length = len(string(value.Weight))
			}
		}
		sort.Strings(c.SortedServices)

		for n, v := range c.SortedServices {
			c.SortedServices[n] = strings.Split(v, "|")[1]
		}
	}

}
