[![Report card](http://goreportcard.com/badge/microo8/wallgen)](http://goreportcard.com/report/microo8/wallgen)

# wallgen

little wallpaper generator written in go, that uses [unsplash.com](https://unsplash.com) to download an random wallpaper and adds some text to it.

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
  -dpi int
    	DPI for the text (default 100)
  -font-file string
    	path to TrueType font
  -font-size int
    	Font size for the text (default 120)
  -h int
    	height of the image (default 1080)
  -o string
    	output file (default "wallpaper.png")
  -t string
    	printed text (default "MEH")
  -w int
    	width of the image (default 1920)
```

the output saves in png or jpg, determined by the output file extension

supports multiline text eg. `$ wallgen -t "foo\nbar"`

## example

![wallpaper example](https://raw.githubusercontent.com/microo8/wallgen/master/wallpaper.png "Wallpaper")

You can use it eg. in your i3 config (generates a new wallpaper on every login):

```
exec_always --no-startup-id /home/$USER/go/bin/wallgen -t "My favorite quote" -o /home/$USER/.wallpaper.png
exec_always --no-startup-id feh --bg-fill /home/$USER/.wallpaper.png
```

##TODO

- support special characters like: ♥, ✓,♫
