# PixelMek 3D
PixelMek 3D is an unofficial BattleTech raycasted game using community contributed pixel mech artwork.
It is written in the [Go programming language](https://go.dev/) using the
[Ebitengine 2D game engine](https://ebitengine.org/).

## This is still a work in progress...

To see it in action, visit the [YouTube Playlist](https://www.youtube.com/playlist?list=PLOINtzQqJWIjJazpjglLLukTZF3KBNghR)

![Screenshot](docs/images/screenshot.png?raw=true)

## Running PixelMek 3D

PixelMek 3D can be run from pre-compiled binaries or directly from the source code.

### Release binaries

The easiest way to run PixelMek 3D is to download the appropriate pre-compiled binary for your
operating system from the [Releases](https://github.com/pixelmek-3d/pixelmek-3d/releases) page.
From the latest release entry on the page, expand the `Assets` section and download
the correct binary file or archive (`.tar.gz`) containing the binary file.
No installation is necessary.

- Linux - `pixelmek-3d-lnx.tar.gz`
- MacOS - `pixelmek-3d-mac.tar.gz`
- Windows - `pixelmek-3d.exe`

### Source code

To run the program from source, you will first need to download or use git to clone
the source code. The source code is also available in the `Assets` section of the
[Releases](https://github.com/pixelmek-3d/pixelmek-3d/releases) page.

Required Software to run from source:

- Git - <https://git-scm.com/>
- Go - <https://go.dev/>

> [!NOTE]
> The current minimum required version of Go can be found in [go.mod](./go.mod)
> near the top of the file (e.g. `go X.Y`).

> [!IMPORTANT]
> Some operating systems may require additional dependencies to be installed to run from source code.
> Refer to the Ebitengine installation documentation: <https://ebitengine.org/en/documents/install.html>

Run the following command from the root project directory containing the `main.go` file:

```bash
go run main.go
```

## Copyright and License Information

PixelMek - Copyright (C) 2016 Eric Harbeston ([harbdog](https://github.com/harbdog))

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

_MechWarrior, BattleMech, ‘Mech, and AeroTech are registered trademarks of
The Topps Company, Inc. Original BattleTech material Copyright by Catalyst Game Labs
All Rights Reserved. Used without permission._
