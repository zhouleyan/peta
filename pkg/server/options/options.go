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
	"peta.io/peta/pkg/utils/iputils"
	"strings"
)

type APIServerOptions struct {
	*ServerRunOptions
	ConfigFile string
	DebugMode  bool
}

func NewAPIServerOptions() *APIServerOptions {
	return &APIServerOptions{
		ServerRunOptions: NewServerRunOptions(),
	}
}

func (s *APIServerOptions) Flags() (nfs NamedFlagSets) {
	fs := nfs.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "enable debug mode")
	s.ServerRunOptions.AddFlags(fs, s.ServerRunOptions)

	fs = nfs.FlagSet("klog")
	local := flag.NewFlagSet("local", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return nfs
}

type Options struct {
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

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet, c *ServerRunOptions) {

	fs.StringVar(&s.BindAddress, "bind-address", c.BindAddress, "server bind address")
	fs.IntVar(&s.InsecurePort, "insecure-port", c.InsecurePort, "insecure port number")
	fs.IntVar(&s.SecurePort, "secure-port", s.SecurePort, "secure port number")
	fs.StringVar(&s.TLSCertFile, "tls-cert-file", c.TLSCertFile, "tls cert file")
	fs.StringVar(&s.TLSPrivateKey, "tls-private-key", c.TLSPrivateKey, "tls private key")
}
