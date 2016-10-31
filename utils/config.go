/* {{{ Copyright (c) Paul R. Tagliamonte <paultag@debian.org>, 2016
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE. }}} */

package utils

import (
	"io"
	"strconv"

	"pault.ag/go/debian/control"
)

type SuiteConfig struct {
	control.Paragraph

	Suite       string
	Components  []string `delim:"," strip:"\n \r\t"`
	Description string
	Origin      string
	Version     string
}

type GlobalConfig struct {
	control.Paragraph

	Archive  string
	Database string
	Keyring  string
	KeyID    string
}

func (g GlobalConfig) KeyIDInt() (uint64, error) {
	return strconv.ParseUint(g.KeyID, 16, 64)
}

type Configuration struct {
	Global GlobalConfig
	Suites []SuiteConfig
}

func ParseConfiguration(reader io.Reader) (*Configuration, error) {
	decoder, err := control.NewDecoder(reader, nil)
	if err != nil {
		return nil, err
	}
	config := Configuration{
		Global: GlobalConfig{},
		Suites: []SuiteConfig{},
	}
	if err := decoder.Decode(&config.Global); err != nil {
		return nil, err
	}
	for {
		suite := SuiteConfig{}
		err := decoder.Decode(&suite)
		if err == io.EOF {
			break
		}
		config.Suites = append(config.Suites, suite)
	}
	return &config, nil
}

// vim: foldmethod=marker
