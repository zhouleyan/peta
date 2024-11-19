/*
 *  This file is part of PETA.
 *  Copyright (C) 2024 The PETA Authors.
 *  PETA is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  PETA is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with PETA. If not, see <https://www.gnu.org/licenses/>.
 */

package options

import (
	"flag"
	"fmt"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"os"
	"peta.io/peta/pkg/persistence"
	"peta.io/peta/pkg/server/auditing"
	"peta.io/peta/pkg/server/metrics"
	"peta.io/peta/pkg/utils/iputils"
	"strings"
)

const (
	defaultConfigName = "peta"
	defaultConfigPath = "/etc/default/peta"

	envPrefix = "PETA"

	BindAddress   = "bind-address"
	InsecurePort  = "insecure-port"
	SecurePort    = "secure-port"
	TLSCertFile   = "tls-cert-file"
	TLSPrivateKey = "tls-private-key"
)

type APIServerOptions struct {
	ConfigFile        string
	DebugMode         bool
	*ServerRunOptions `json:"server,omitempty" yaml:"server,omitempty" mapstructure:"server"`
	AuditingOptions   *auditing.Options    `json:"auditing,omitempty" yaml:"auditing,omitempty" mapstructure:"auditing"`
	MetricsOptions    *metrics.Options     `json:"metrics,omitempty" yaml:"metrics,omitempty" mapstructure:"metrics"`
	DatabaseOptions   *persistence.Options `json:"database,omitempty" yaml:"database,omitempty" mapstructure:"database"`
}

func NewAPIServerOptions() *APIServerOptions {
	o := &APIServerOptions{
		ServerRunOptions: NewServerRunOptions(),
		AuditingOptions:  auditing.NewOptions(),
		MetricsOptions:   metrics.NewOptions(),
		DatabaseOptions:  persistence.NewOptions(),
	}
	return o
}

func (s *APIServerOptions) Merge(fs *pflag.FlagSet, conf *APIServerOptions) {
	s.AuditingOptions.Merge(fs, conf.AuditingOptions)
	s.ServerRunOptions.Merge(fs, conf.ServerRunOptions)
	s.MetricsOptions.Merge(fs, conf.MetricsOptions)
	s.DatabaseOptions.Merge(fs, conf.DatabaseOptions)
}

func (s *APIServerOptions) Flags() *NamedFlagSets {
	nfs := new(NamedFlagSets)
	fs := nfs.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "enable debug mode")
	fs.StringVar(&s.ConfigFile, "config", defaultConfigPath, "config file path")
	s.ServerRunOptions.AddFlags(fs)

	fs = nfs.FlagSet("klog")
	local := flag.NewFlagSet("local", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return nfs
}

func (s *APIServerOptions) AddFlags(nfs *NamedFlagSets) {
	s.AuditingOptions.AddFlags(nfs.Insert("auditing", 1))
	s.MetricsOptions.AddFlags(nfs.Insert("metrics", 1))
	s.DatabaseOptions.AddFlags(nfs.Insert("database", 1))
}

type ServerRunOptions struct {
	// server bind address
	BindAddress string `json:"bindAddress,omitempty" yaml:"bindAddress,omitempty" mapstructure:"bindAddress"`

	// insecure port number
	InsecurePort int `json:"insecurePort,omitempty" yaml:"insecurePort,omitempty" mapstructure:"insecurePort"`

	// secure port number
	SecurePort int `json:"securePort,omitempty" yaml:"securePort,omitempty" mapstructure:"securePort"`

	// tls cert file
	TLSCertFile string `json:"tlsCertFile,omitempty" yaml:"tlsCertFile,omitempty" mapstructure:"tlsCertFile"`

	// tls private key file
	TLSPrivateKey string `json:"tlsPrivateKey,omitempty" yaml:"tlsPrivateKey,omitempty" mapstructure:"tlsPrivateKey"`
}

func NewServerRunOptions() *ServerRunOptions {
	// create default server run options
	s := ServerRunOptions{
		BindAddress:   "0.0.0.0",
		InsecurePort:  9090,
		SecurePort:    0,
		TLSCertFile:   "",
		TLSPrivateKey: "",
	}

	return &s
}

func (s *ServerRunOptions) Validate() []error {
	var errs []error

	if s.SecurePort == 0 && s.InsecurePort == 0 {
		errs = append(errs, fmt.Errorf("* insecure-port and secure-port can not be disabled at the same time"))
	}

	if iputils.IsValidPort(s.SecurePort) {
		if s.TLSCertFile == "" {
			errs = append(errs, fmt.Errorf("* tls-cert-file is empty while secure serving"))
		} else {
			if _, err := os.Stat(s.TLSCertFile); err != nil {
				errs = append(errs, err)
			}
		}

		if s.TLSPrivateKey == "" {
			errs = append(errs, fmt.Errorf("* tls-private-key file is empty while secure serving"))
		} else {
			if _, err := os.Stat(s.TLSPrivateKey); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func (s *ServerRunOptions) Merge(fs *pflag.FlagSet, conf *ServerRunOptions) {
	if f := fs.Lookup(BindAddress); f != nil && !f.Changed {
		s.BindAddress = conf.BindAddress
	}
	if f := fs.Lookup(InsecurePort); f != nil && !f.Changed {
		s.InsecurePort = conf.InsecurePort
	}
	if f := fs.Lookup(SecurePort); f != nil && !f.Changed {
		s.SecurePort = conf.SecurePort
	}
	if f := fs.Lookup(TLSCertFile); f != nil && !f.Changed {
		s.TLSCertFile = conf.TLSCertFile
	}
	if f := fs.Lookup(TLSPrivateKey); f != nil && !f.Changed {
		s.TLSPrivateKey = conf.TLSPrivateKey
	}
}

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {

	fs.StringVar(&s.BindAddress, BindAddress, s.BindAddress, "server bind address")
	fs.IntVar(&s.InsecurePort, InsecurePort, s.InsecurePort, "insecure port number")
	fs.IntVar(&s.SecurePort, SecurePort, s.SecurePort, "secure port number")
	fs.StringVar(&s.TLSCertFile, TLSCertFile, s.TLSCertFile, "tls cert file")
	fs.StringVar(&s.TLSPrivateKey, TLSPrivateKey, s.TLSPrivateKey, "tls private key")
}
