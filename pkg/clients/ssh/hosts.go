/*
 *  This file is part of PETA.
 *  Copyright (C) 2025 The PETA Authors.
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

package ssh

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func KnownHosts(file string) (ssh.HostKeyCallback, error) {
	return knownhosts.New(file)
}

// DefaultKnownHostsPath returns default user knows a host file.
func DefaultKnownHostsPath() (string, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".ssh", "known_hosts"), nil
}

// VerifyHost check new connections public keys if the key not trusted you should return an error.
func VerifyHost(knownFile string, askAddKnownHost bool) ssh.HostKeyCallback {
	return func(host string, remote net.Addr, key ssh.PublicKey) error {
		hostFound, err := CheckKnownHost(host, remote, key, knownFile)

		var keyErr *knownhosts.KeyError

		// Host in known hosts but key mismatch!
		// Maybe because of MAN IN THE MIDDLE ATTACK!
		if hostFound && err != nil {
			if errors.As(err, &keyErr) && len(keyErr.Want) > 0 {
				return AddKnownHost(host, remote, key, knownFile, askAddKnownHost)
			}
			return err
		}

		// public key not found.
		if !hostFound && err != nil && errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			return AddKnownHost(host, remote, key, knownFile, askAddKnownHost)
		}

		return err
	}
}

// CheckKnownHost checks is host in a known hosts file.
// It returns the host found in a known_hosts file and error, if the host found in
// a known_hosts file and error not nil that means public key mismatch.
func CheckKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string) (found bool, err error) {
	var keyErr *knownhosts.KeyError

	if knownFile == "" {
		path, err := DefaultKnownHostsPath()
		if err != nil {
			return false, err
		}
		knownFile = path
	}

	// Get host key callback.
	callback, err := KnownHosts(knownFile)

	if err != nil {
		return false, err
	}

	// check if the host already exists.
	err = callback(host, remote, key)

	if err != nil {
		// Make sure that the error returned from the callback is host not in file error.
		// If keyErr.Want is greater than length 0, that means the host is in file with a different key.
		if errors.As(err, &keyErr) && len(keyErr.Want) > 0 {
			return true, keyErr
		}

		if errors.As(err, &keyErr) && len(keyErr.Want) == 0 {
			return false, keyErr
		}

		return true, err
	}

	// Key is not trusted because it is not in the file.
	return true, nil
}

// AddKnownHost add a host to a known hosts file.
func AddKnownHost(host string, remote net.Addr, key ssh.PublicKey, knownFile string, askAddKnownHost bool) (err error) {
	if knownFile == "" {
		path, err := DefaultKnownHostsPath()
		if err != nil {
			return err
		}

		knownFile = path
	}

	f, err := os.OpenFile(knownFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		dErr := f.Close()
		if dErr != nil && err == nil {
			err = dErr
		}
	}(f)

	remoteNormalized := knownhosts.Normalize(remote.String())
	hostNormalized := knownhosts.Normalize(host)
	addresses := []string{remoteNormalized}

	if hostNormalized != remoteNormalized {
		addresses = append(addresses, hostNormalized)
	}

	if askAddKnownHost {
		if askIsHostTrusted(host, key) == false {
			return errors.New("you typed no, aborted")
		}
	}

	_, err = f.WriteString(knownhosts.Line(addresses, key) + "\n")

	return err
}

func askIsHostTrusted(host string, key ssh.PublicKey) bool {

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Unknown Host: %s \nFingerprint: %s \n", host, ssh.FingerprintSHA256(key))
	fmt.Print("Would you like to add it? type yes or no: ")

	a, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(strings.TrimSpace(a)) == "yes"
}
