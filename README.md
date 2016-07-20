# wallgen

little wallpaper generator that uses [https://unsplash.com/](unsplash.com) to download an random wallpaper and adds some text to it.

## install

```bash
$ go get github.com/microo8/wallgen
```

## wallpaper generation

```bash
$ wallgen -t "hello github"
```

```
Usage of wallgen:
  -h int
        height of the image (default 1080)
  -o string
        output file (default "wallpaper.png")
  -t string
        printed text (default "MEH")
  -w int
        width of the image (default 1920)
```

## example

![wallpaper example](https://github.com/microo8/wallgen/raw/master/src/wallpaper.png "Wallpaper")
