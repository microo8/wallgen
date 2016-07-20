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

![wallpaper example](https://raw.githubusercontent.com/microo8/wallgen/master/wallpaper.png "Wallpaper")

You can use it eg. in your i3 config (generates a new wallpaper on every login):

```
exec_always --no-startup-id /home/$USER/go/bin/wallgen -t "My favorite quote" -o /home/$USER/.wallpaper.jpg
exec_always --no-startup-id feh --bg-fill /home/$USER/.wallpaper.jpg
```
