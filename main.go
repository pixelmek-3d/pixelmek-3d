// Copyright 2022 Eric H.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"github.com/pixelmek-3d/pixelmek-3d/game"
	log "github.com/sirupsen/logrus"
)

func main() {
	// setup logging
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	formatter := game.LogFormat{}
	formatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(&formatter)

	// run the game
	g := game.NewGame()
	g.Run()
}
