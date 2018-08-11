# Clipboard to youtube-dl

This Go application will monitor your clipboard for urls and automatically starts download with [youtube-dl](https://github.com/rg3/youtube-dl/) ([list of supported sites](https://github.com/rg3/youtube-dl/blob/master/docs/supportedsites.md)). 
After download has been finished you'll get a system notification with detailed information (currently only works with Linux). In your system tray you'll find a new icon to control this application.   

**IMPORTANT**

Currently I can only support running this app under Windows and Linux.
 
## Configuration

You can use all configurations of [youtube-dl](https://github.com/rg3/youtube-dl/) like video or audio format, quality, output directory and many more.

## Building from sources

### Requirements

* [Go](https://golang.org/doc/install) including [dep](https://github.com/golang/dep)
* [youtube-dl](https://github.com/rg3/youtube-dl/)
* [Docker CE](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-ce)

First you need to create a new folder under your ``$GOPATH``.

    $ mkdir -p $GOPATH/src/github.com/hebestreit

Navigate in this folder and checkout this repository.

    $ cd $GOPATH/src/github.com/hebestreit
    $ git clone https://github.com/hebestreit/clipboard-yt-dl.git

For this part you'll need docker to build application for all platforms.

    $ cd clipboard-yt-dl
    $ make all

Now you can find all binaries under ``$GOPATH/src/github.com/hebestreit/clipboard-yt-dl/bin`` and start copying over the world!

    $ ls -l bin/
    clipboard-yt-dl_darwin.app
    clipboard-yt-dl_linux
    clipboard-yt-dl_windows.exe

# Dependencies

This is a list of dependencies I'm using in this project.

* [github.com/shivylp/clipboard](https://github.com/shivylp/clipboard) for monitoring clipboard which is a fork of [github.com/atotto/clipboard](https://github.com/atotto/clipboard).
* [github.com/0xAX/notificator](https://github.com/0xAX/notificator) sending notifications
* [github.com/getlantern/systray](https://github.com/getlantern/systray) menu item in systray for user interactions