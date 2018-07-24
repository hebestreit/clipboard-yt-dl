# Clipboard to youtube-dl

This Go application will monitor your clipboard for YouTube urls and automatically starts download with [youtube-dl](https://github.com/rg3/youtube-dl/). After download has been finished you'll get a notification with detailed information.

## Configuration

Check documentation of [youtube-dl](https://github.com/rg3/youtube-dl/) for more information.

## Building from sources

### Requirements

* [Go](https://golang.org/doc/install) including [dep](https://github.com/golang/dep)
* [youtube-dl](https://github.com/rg3/youtube-dl/)

Run following commands.

    $ mkdir -p $GOPATH/src/github.com/hebestreit
    $ cd $GOPATH/src/github.com/hebestreit
    $ git clone https://github.com/hebestreit/clipboard-yt-dl.git
    $ cd clipboard-yt-dl
    $ make all

Now you can run this command and start copying over the world!

    $ ./clipboard-yt-dl

# Dependencies

This is a list of dependencies I'm using in this project.

* [github.com/shivylp/clipboard](https://github.com/shivylp/clipboard) for monitoring clipboard which is a fork of [github.com/atotto/clipboard](https://github.com/atotto/clipboard).
* [github.com/0xAX/notificator](https://github.com/0xAX/notificator) sending notifications
