# Coalesced Inifier

A tool to fiddle with the binary encoded INI files of Unreal Engine 3 games such as Gears of War 2.

I wanted to change the graphics settings on my Xbox 360 version of Gears of War 2, so I wrote this little tool that
extracts all the INI files that are binary encoded in the file `Coalesced_int.bin`.

## Supported games

It has been tested to read and write the files of the following games:

- Gears of War 2
- Lollipop Chainsaw
- Thief
- DMC: Devil May Cry

You may need a modified XEX depending on the game.

## Web UI

The application includes a web UI:

```shell
./coalesced-inifier web
```

You can now access it on [http://localhost:8080](http://localhost:8080).

Alternatively go to: [https://coalesced.getscab.net](https://coalesced.getscab.net) for a hosted version that may
or may not be up-to-date.

## Example:  Unpacking the INI files

To unpack the ini files from a gow2 coalesced file, run the following:

```shell
./coalesced-inifier unpack -i /path/to/Coalesced_int.bin -o  my-output-dir
```

You now should have a folder structure containing the INI files:

```text
ls  output/GearGame/Config 
GearAI.ini  GearCamera.ini  GearPawn.ini  GearPawnMP.ini  GearPlaylist.ini  GearWeapon.ini  GearWeaponMP.ini  Xe-GearEngine.ini  Xe-GearGame.ini  Xe-GearInput.ini  Xe-GearUI.ini
```

## Example:  Packing the INI files

To coalesce the INI files run the following:

```shell
./coalesced-inifier pack -i my-output-dir -o /path/to/coalesced_int.bin -g gow2
```

It will output something like:

```text
[...]
Packing 'my-output-dir/GearGame/Localization/INT/locust_skorge_chatter.INT' as '..\GearGame\Localization\INT\locust_skorge_chatter.INT'... done.
Packing 'my-output-dir/GearGame/Localization/INT/locust_theron_chatter.INT' as '..\GearGame\Localization\INT\locust_theron_chatter.INT'... done.

File length: 1305998
  File hash: 3cab0a4f2237b00875ab02061bdd46a6

Successfully packed 63 files to 'coalesced.bin'.
```

You can now update the `Xbox360TOC.txt` with the hash and length.

## Usage

### `unpack` command usage

```text
Coalesced Inifier
0.1.0
Usage: coalesced-inifier unpack --input <coalesced bin> --output <ini folder>

Options:
  --input <coalesced bin>, -i <coalesced bin>
                         The coalesced bin file such as Coalesced_int.bin
  --output <ini folder>, -o <ini folder>
                         The path containing your INI files
  --help, -h             display this help and exit
  --version              display version and exit
```


### `pack` command usage

```text
Coalesced Inifier
0.1.0
Usage: coalesced-inifier pack --input <ini folder> --output <coalesced bin> --game GAME

Options:
  --input <ini folder>, -i <ini folder>
                         The path containing your INI files
  --output <coalesced bin>, -o <coalesced bin>
                         The coalesced bin file such as Coalesced_int.bin
  --game GAME, -g GAME   game to pack data for, one of: gow2, lollipop
  --help, -h             display this help and exit
  --version              display version and exit
```
