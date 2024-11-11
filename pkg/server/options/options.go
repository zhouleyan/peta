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
	"peta.io/peta/pkg/server/auditing"
	"peta.io/peta/pkg/utils/iputils"
	"strings"
)

const (
	defaultConfigName = "peta"

	defaultConfigPath = "/etc/peta"

	envPrefix = "PETA"
)

type APIServerOptions struct {
	*ServerRunOptions
	*Options
	ConfigFile string
	DebugMode  bool
}

func NewAPIServerOptions() *APIServerOptions {
	o := &APIServerOptions{
		ServerRunOptions: NewServerRunOptions(),
		Options:          NewOptions(),
	}
	return o
}

func (s *APIServerOptions) Merge(conf *Options) {
	if conf == nil {
		return
	}
	if s.AuditingOptions == nil {
		s.AuditingOptions = conf.AuditingOptions
	}
}

func (s *APIServerOptions) Flags() (nfs NamedFlagSets) {
	fs := nfs.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "enable debug mode")
	fs.StringVar(&s.ConfigFile, "config", "/etc/default/peta", "config file path")
	s.ServerRunOptions.AddFlags(fs)
	s.AuditingOptions.AddFlags(nfs.FlagSet("auditing"))

	fs = nfs.FlagSet("klog")
	local := flag.NewFlagSet("local", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return nfs
}

type ServerRunOptions struct {
	// server bind address
	BindAddress string

	// insecure port number
	InsecurePort int

	// secure port number
	SecurePort int

	// tls cert file
	TLSCertFile string

	// tls private key file
	TLSPrivateKey string
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
		errs = append(errs, fmt.Errorf("insecure and secure port can not be disabled at the same time"))
	}

	if iputils.IsValidPort(s.SecurePort) {
		if s.TLSCertFile == "" {
			errs = append(errs, fmt.Errorf("tls cert file is empty while secure serving"))
		} else {
			if _, err := os.Stat(s.TLSCertFile); err != nil {
				errs = append(errs, err)
			}
		}

		if s.TLSPrivateKey == "" {
			errs = append(errs, fmt.Errorf("tls private key file is empty while secure serving"))
		} else {
			if _, err := os.Stat(s.TLSPrivateKey); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {

	fs.StringVar(&s.BindAddress, "bind-address", s.BindAddress, "server bind address")
	fs.IntVar(&s.InsecurePort, "insecure-port", s.InsecurePort, "insecure port number")
	fs.IntVar(&s.SecurePort, "secure-port", s.SecurePort, "secure port number")
	fs.StringVar(&s.TLSCertFile, "tls-cert-file", s.TLSCertFile, "tls cert file")
	fs.StringVar(&s.TLSPrivateKey, "tls-private-key", s.TLSPrivateKey, "tls private key")
}

type Options struct {
	AuditingOptions *auditing.Options `json:"auditing,omitempty" yaml:"auditing,omitempty" mapstructure:"auditing"`
}

func NewOptions() *Options {
	return &Options{
		AuditingOptions: auditing.NewOptions(),
	}
}
