// Copyright (C) 2020 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// Takeout is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for
// more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package log

import (
	l "log"
	"os"
)

type Logger interface {
	// Print followed by Panic
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
	// Print followed by Exit
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	// Print
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

var logger = defaultLogger()

func defaultLogger() Logger {
	return l.New(os.Stdout, "", l.LstdFlags)
}

// Panic if err
func CheckError(err error) {
	if err != nil {
		logger.Panicln(err)
	}
}

func Panicf(format string, v ...interface{}) {
	logger.Panicf(format, v...)
}

func Panicln(v ...interface{}) {
	logger.Panicln(v...)
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

func Fatalln(v ...interface{}) {
	logger.Fatalln(v...)
}

func Printf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

func Println(v ...interface{}) {
	logger.Println(v...)
}
