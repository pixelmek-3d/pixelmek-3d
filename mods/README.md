# Using custom resources with mods

Static [resource files delivered with the game](../game/resources/) can be overridden, or new resources can be
added to the game, by creating a folder named `mods` in the same directory as the `pixelmek-3d` executable.
Inside the `mods` folder you can place any number of `.tar` non-compressed archive files containing any number
of custom files to override or add to those provided with the game by default.

## Creating a custom resource mod archive

Start by locating the path(s) to the resources you would like to modify or add to, which start from the
[resources]((../game/resources/)) folder in the source code repository:

```text
audio/
fonts/
maps/
missions/
sprites/
textures/
units/
weapons/
```

For example, if you wanted to override the sprite for the [Jenner II-C](../game/resources/sprites/mechs/jenner_iic.png)
start by making a copy of the image, making changes to it using a graphics application, and placing it in the
following folder path next to the pixelmek-3d executable file:

```text
mods/sprites/mechs/jenner_iic.png
```

From within the `mods` directory created next to the pixelmek-3d executable, create the `.tar` archive
starting from `sprites` folder:

```bash
tar -cvf my_jenner.tar sprites

sprites/
sprites/mechs/
sprites/mechs/jenner_iic.png
```

Run the game with the `--debug` parameter to confirm the log output displays the expected mods files and contents.

```bash
pixelmek-3d --debug

[DEBUG] found mods file mods/my_jenner.tar
[DEBUG] [mods/my_jenner.tar] .
[DEBUG] [mods/my_jenner.tar] sprites
[DEBUG] [mods/my_jenner.tar] sprites/mechs
[DEBUG] [mods/my_jenner.tar] sprites/mechs/jenner_iic.png
```

## Known limitations or issues using mods

- Golang source files (`.go` extension) cannot currently be overridden using mods.
- Certain resources can only be overridden, not added: `fonts`, `menu`, `shaders`.
- Compressed tar archives (such as `.gz`, `.tgz`) are not currently supported, only non-compressed `.tar`.
- [Texture resources](../game/resources/textures/) for `floors`, `skies`, and `walls` must be
  exactly `256x256` pixels in size. Sprites, however, do not have this limitation.
