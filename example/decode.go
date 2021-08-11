package main

import "C"
import (
	"fmt"
	"os"
	"unsafe"

	"github.com/asticode/goav/avcodec"
	"github.com/asticode/goav/avformat"
	"github.com/asticode/goav/avutil"
	"github.com/asticode/goav/swscale"
)

func SaveFrame(frame *avutil.Frame, width, height, frameNumber int) {
	// Open file
	fileName := fmt.Sprintf("frame%d.ppm", frameNumber)
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error Reading")
	}
	defer file.Close()

	// Write header
	header := fmt.Sprintf("P6\n%d %d\n255\n", width, height)
	file.Write([]byte(header))

	// Write pixel data
	for y := 0; y < height; y++ {
		data0 := frame.Data()[0]
		buf := make([]byte, width*3)
		startPos := uintptr(unsafe.Pointer(data0)) + uintptr(y)*uintptr(frame.Linesize()[0])
		for i := 0; i < width*3; i++ {
			element := *(*uint8)(unsafe.Pointer(startPos + uintptr(i)))
			buf[i] = element
		}
		file.Write(buf)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a file path")
		os.Exit(1)
	}

	fileName := os.Args[1]
	fmt.Println("input: ", fileName)

	formatCtx := avformat.AvformatAllocContext()
	defer avformat.AvformatCloseInput(formatCtx)
	if ret := avformat.AvformatOpenInput(&formatCtx, fileName, nil, nil); ret != 0 {
		fmt.Println("Can not open file: ", fileName)
		os.Exit(1)
	}

	if ret := formatCtx.AvformatFindStreamInfo(nil); ret < 0 {
		fmt.Println("Can not find stream information")
		os.Exit(1)
	}

	// 打印ffmpeg日志
	formatCtx.AvDumpFormat(0, fileName, 0)

	videoStreamIndex := -1
	var codecParameters *avcodec.CodecParameters
	for index, stream := range formatCtx.Streams() {
		switch stream.CodecParameters().CodecType() {
		case avcodec.AVMEDIA_TYPE_VIDEO:
			videoStreamIndex = index
			codecParameters = stream.CodecParameters()
			break
		}
	}
	if videoStreamIndex == -1 || codecParameters == nil {
		fmt.Println("Can not find video stream")
		os.Exit(1)
	}

	codec := avcodec.AvcodecFindDecoder(codecParameters.CodecId())
	if codec == nil {
		fmt.Println("Unsupported codec: ", codec)
		os.Exit(1)
	}
	ctx := codec.AvcodecAllocContext3()
	// Copy codec parameters
	if ret := avcodec.AvcodecParametersToContext(ctx, codecParameters); ret < 0 {
		fmt.Println("avcodec.AvcodecParametersToContext failed: ", avutil.AvStrerr(ret))
		os.Exit(1)
	}

	// Open codec
	if ret := ctx.AvcodecOpen2(codec, nil); ret < 0 {
		fmt.Println("ctx.AvcodecOpen2 failed: ", avutil.AvStrerr(ret))
		os.Exit(1)
	}
	defer ctx.AvcodecClose()

	pFrame := avutil.AvFrameAlloc()
	defer avutil.AvFrameFree(pFrame)

	packet := avcodec.AvPacketAlloc()
	defer avcodec.AvPacketFree(packet)

	// Allocate an AVFrame structure
	pFrameRGB := avutil.AvFrameAlloc()
	if pFrameRGB == nil {
		fmt.Println("Unable to allocate RGB Frame")
		return
	}
	defer avutil.AvFrameFree(pFrameRGB)
	pFrameRGB.SetFormat(avutil.AV_PIX_FMT_RGB24)
	pFrameRGB.SetWidth(ctx.Width())
	pFrameRGB.SetHeight(ctx.Height())
	avutil.AvFrameGetBuffer(pFrameRGB, 0)
	// initialize SWS context for software scaling
	swsCtx := swscale.SwsGetcontext(
		ctx.Width(),
		ctx.Height(),
		ctx.PixFmt(),
		ctx.Width(),
		ctx.Height(),
		avutil.AV_PIX_FMT_RGB24,
		swscale.SWS_BILINEAR,
		nil,
		nil,
		nil,
	)

	frameNumber := 1
	for formatCtx.AvReadFrame(packet) >= 0 {
		if packet.StreamIndex() == videoStreamIndex {
			ret := ctx.SendPacket(packet)
			if ret < 0 {
				fmt.Println("Error while sending a packet to decoder: ", avutil.AvStrerr(ret))
			}
			for ret >= 0 {
				ret := ctx.ReceiveFrame(pFrame)
				if ret == avutil.AVERROR_EAGAIN || ret == avutil.AVERROR_EOF {
					break
				}
				if ret < 0 {
					fmt.Println("Error while receive a frame from the decoder: ", avutil.AvStrerr(ret))
					os.Exit(1)
				}
				swscale.SwsScale(swsCtx,
					pFrame.Data(),
					pFrame.Linesize(),
					0,
					ctx.Height(),
					pFrameRGB.Data(),
					pFrameRGB.Linesize())

				SaveFrame(pFrameRGB, ctx.Width(), ctx.Height(), frameNumber)
				frameNumber++
			}
		}
	}

}