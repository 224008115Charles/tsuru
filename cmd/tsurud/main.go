// Copyright 2016 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"github.com/tsuru/config"
	"github.com/tsuru/tsuru/api"
	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/provision"
	_ "github.com/tsuru/tsuru/provision/docker"
	_ "github.com/tsuru/tsuru/repository/gandalf"
)

const defaultConfigPath = "/etc/tsuru/tsuru.conf"

var configPath = defaultConfigPath

func buildManager() *cmd.Manager {
	m := cmd.NewManager("tsurud", api.Version, "", os.Stdout, os.Stderr, os.Stdin, nil)
	m.Register(&tsurudCommand{Command: &apiCmd{}})
	m.Register(&tsurudCommand{Command: tokenCmd{}})
	m.Register(&tsurudCommand{Command: &migrateCmd{}})
	m.Register(&tsurudCommand{Command: gandalfSyncCmd{}})
	m.Register(&tsurudCommand{Command: createRootUserCmd{}})
	m.Register(&migrationListCmd{})
	err := registerProvisionersCommands(m)
	if err != nil {
		log.Fatalf("unable to register commands: %v", err)
	}
	return m
}

func registerProvisionersCommands(m *cmd.Manager) error {
	provisioners, err := provision.Registry()
	if err != nil {
		return err
	}
	for _, p := range provisioners {
		if c, ok := p.(cmd.Commandable); ok {
			commands := c.Commands()
			for _, cmd := range commands {
				m.Register(&tsurudCommand{Command: cmd})
			}
		}
	}
	return nil
}

func listenSignals() {
	ch := make(chan os.Signal, 2)
	go func() {
		for sig := range ch {
			switch sig {
			case syscall.SIGUSR1:
				pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
			case syscall.SIGHUP:
				config.ReadConfigFile(configPath)
			}
		}
	}()
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGUSR1)
}

func main() {
	config.ReadConfigFile(configPath)
	listenSignals()
	m := buildManager()
	m.Run(os.Args[1:])
}
