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
	"os"

	"golang.org/x/crypto/ssh"
)

// Auth represents ssh auth method.
type Auth []ssh.AuthMethod

// Password returns password auth method.
func Password(pass string) ssh.AuthMethod {
	return ssh.Password(pass)
}

func Key(privateFile string, passphrase string) (ssh.AuthMethod, error) {
	signer, err := GetSigner(privateFile, passphrase)
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}

func RawKey(privateKey string, passphrase string) (ssh.AuthMethod, error) {
	signer, err := GetSignerForRawKey([]byte(privateKey), passphrase)
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}

func GetSigner(privateFile string, passphrase string) (ssh.Signer, error) {
	var (
		signer ssh.Signer
		err    error
	)

	privateKey, err := os.ReadFile(privateFile)
	if err != nil {
		return nil, err
	}

	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(privateKey)
	}

	return signer, err
}

func GetSignerForRawKey(privateKey []byte, passphrase string) (ssh.Signer, error) {
	var (
		signer ssh.Signer
		err    error
	)

	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(privateKey, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(privateKey)
	}

	return signer, err
}
