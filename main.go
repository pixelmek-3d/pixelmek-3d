/*
 * Copyright (c) 2024 - Eric Harbeston. All Rights Reserved.
 *
 * PixelMek 3D is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 *
 * PixelMek 3D is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with PixelMek 3D. If not, see <http://www.gnu.org/licenses/>.
 */
package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/pixelmek-3d/pixelmek-3d/cmd"
	"github.com/pixelmek-3d/pixelmek-3d/game"
	log "github.com/sirupsen/logrus"
)

func main() {
	// setup logging
	log.SetLevel(log.InfoLevel)
	if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		log.SetOutput(os.Stdout)
	} else {
		f, err := os.OpenFile("pixelmek-3d.log", os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			fmt.Printf("error opening log file: %v\n", err)
		} else {
			fmt.Printf("writing log to file: %s\n", f.Name())
		}
		defer f.Close()

		os.Stdout = f
		os.Stderr = f
		log.SetOutput(f)
	}

	formatter := game.LogFormat{}
	formatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(&formatter)

	cmd.Execute()
}
