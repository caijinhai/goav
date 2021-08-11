Golang 1.16 work with ffmpeg 4.1.6.   Forked from [here](https://github.com/asticode/goav)

## run example
download ffmpep:
https://github.com/caijinhai/ffmpeg-bin


```shell
export FFMPEG_ROOT=/home/caijinhai/tmp/ffmpeg
export CGO_LDFLAGS="-L$FFMPEG_ROOT/lib/ -lavcodec -lavformat -lavutil -lswscale -lswresample -lavdevice -lavfilter"
export CGO_CFLAGS="-I$FFMPEG_ROOT/include"
export PKG_CONFIG_PATH="$FFMPEG_ROOT/lib/pkgconfig"

export LD_LIBRARY_PATH=$FFMPEG_ROOT/lib:$LD_LIBRARY_PATH

cd example
go run decode.go ${test_mp4_path}
```
