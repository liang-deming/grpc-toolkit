// client パッケージは Echo サービスの各 RPC を呼び出すヘルパー関数を提供する。
// ファイル入出力を伴うストリーミング系は echo-client/file 配下のパスを参照する。
package client

import (
	"context"
	"fmt"
	"grpcv2/echo"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CallUnary は UnaryEcho を 1 回呼び出す。
// コンテキストは 1 秒タイムアウト。PerRPCCredentials で Authorization を付与し、サーバー側検証と整合させる。
func CallUnary(client echo.EchoServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req := &echo.EchoRequest{
		Message:   "client send message",
		Timestamp: timestamppb.New(time.Now()),
	}

	auth := GetPerRPCCredentials(FetchToken())

	res, err := client.UnaryEcho(ctx, req, grpc.PerRPCCredentials(auth))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("client recv: %v\n", res.Message)
}

// CallServerStreamingEcho は ServerStreamingEcho を呼び、サーバーから送られる画像チャンクを
// タイムスタンプ付きファイルに書き込む。ストリーム終了は io.EOF で検知する。
func CallServerStream(client echo.EchoServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req := &echo.EchoRequest{
		Message:   "client ask for image",
		Timestamp: timestamppb.New(time.Now()),
	}
	stream, err := client.ServerStreamingEcho(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	filepath := "echo-client/file/" + strconv.FormatInt(time.Now().UnixMilli(), 10) + ".jpg"
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
			break
		}

		file.Write(res.Bytes[:len(res.Bytes)])
		fmt.Printf("client recv: %v\n", res.Message)
	}
}

// CallClientStream はクライアントストリーミング RPC。client.jpg を読み、チャンクを Send し、
// CloseAndRecv でサーバーの最終レスポンスを受け取る。
func CallClientStream(client echo.EchoServiceClient) {
	filepath := "echo-client/file/client.jpg"
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := client.ClientStreamingEcho(ctx)
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if n == 0 {
			break
		}
		err = stream.Send(&echo.EchoRequest{
			Message:   "client send image",
			Bytes:     buf[:n],
			Timestamp: timestamppb.New(time.Now()),
			Length:    int32(n),
		})
		if err != nil {
			log.Fatal(err)
			break
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("client recv: %v\n", res.Message)
}

// CallBidirectional は双方向ストリーミング。送信ゴルーチンで client.jpg を送り、
// 受信ゴルーチンでサーバーからの画像をファイルに保存する。CloseSend で送信側を閉じる。
func CallBidirectional(client echo.EchoServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.BidirectionalStreamingEcho(ctx)
	if err != nil {
		log.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		filepath := "echo-client/file/client.jpg"
		file, err := os.Open(filepath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if err != nil && err != io.EOF {
				log.Fatal(err)
			}
			if n == 0 {
				break
			}
			err = stream.Send(&echo.EchoRequest{
				Message:   "client send image",
				Bytes:     buf[:n],
				Timestamp: timestamppb.New(time.Now()),
				Length:    int32(n),
			})
			if err != nil {
				log.Fatal(err)
				break
			}
		}
		stream.CloseSend()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		filepath := "echo-client/file/" + strconv.FormatInt(time.Now().UnixMilli(), 10) + ".jpg"
		file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
				break
			}
			file.Write(res.Bytes[:len(res.Bytes)])
			fmt.Printf("client recv: %v\n", res.Message)
		}
	}()

	wg.Wait()

}
