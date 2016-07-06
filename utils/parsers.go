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
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
)

func BinaryStringToIterator(db *gorm.DB, binary string) (*BinaryIterator, error) {
	els := strings.Split(binary, "/")
	/* XXX:
	 * binary
	 * binary/version
	 * binary/verison/arch
	 */
	if len(els) < 1 {
		return nil, fmt.Errorf("Empty string passed")
	}
	if len(els) > 3 {
		return nil, fmt.Errorf("Too many binary qualifiers passed")
	}

	query := db.Table("binaries").Where("name = ?", els[0])

	if len(els) >= 2 {
		query = query.Where("version = ?", els[1])
	}

	if len(els) >= 3 {
		query = query.Where("arch = ?", els[2])
	}

	return NewBinaryIterator(query)
}

// vim: foldmethod=marker
