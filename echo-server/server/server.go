// server パッケージは Echo gRPC サービスの具体実装を提供する。
//
// UnaryEcho は単純な文字列エコー。ストリーミング系は主に画像ファイルの
// 読み書きを通じたデモとして実装されており、クライアント側の対応する呼び出しとセットで動作する。
package server

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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EchoServer は echo.EchoServiceServer インターフェースの実装。
// 埋め込んだ UnimplementedEchoServiceServer により、将来 RPC が追加されても
// 未実装メソッドでコンパイルが壊れにくい形になっている。
type EchoServer struct {
	echo.UnimplementedEchoServiceServer
}

// UnaryEcho は 1 リクエスト 1 レスポンスの RPC。
// リクエストの message をプレフィックス付きで返すだけの最小実装。
func (EchoServer) UnaryEcho(ctx context.Context, req *echo.EchoRequest) (*echo.EchoResponse, error) {
	return &echo.EchoResponse{Message: "server got your message: " + req.Message}, nil
}

// ServerStreamingEcho はサーバーがクライアントへ複数フレームを送るストリーミング RPC。
// 固定パス echo-server/file/server.jpg を開き、1024 バイトずつ読み出して EchoResponse.Bytes に載せて送信する。
// ファイルが無い場合や読み取りエラー時は gRPC の Internal ステータスで返す。
func (EchoServer) ServerStreamingEcho(req *echo.EchoRequest, stream echo.EchoService_ServerStreamingEchoServer) error {
	fmt.Printf("server recv: %v\n", req.Message)
	filepath := "echo-server/file/server.jpg"
	file, err := os.Open(filepath)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to open file: %v", err)
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
		stream.Send(&echo.EchoResponse{
			Bytes:   buf[:n],
			Message: "server send image",
		})
	}
	return nil
}

// ClientStreamingEcho はクライアントから複数リクエストを受け取り、最後に 1 レスポンスを返す RPC。
// 受信した各 EchoRequest の Bytes を、タイムスタンプ付きファイル名の画像ファイルに追記保存する。
// ストリーム終端（io.EOF）後に SendAndClose で応答メッセージを返す。
func (EchoServer) ClientStreamingEcho(stream echo.EchoService_ClientStreamingEchoServer) error {
	fmt.Println("server start to recv client streaming")
	filepath := "echo-server/file/" + strconv.FormatInt(time.Now().UnixMilli(), 10) + ".jpg"
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to create file: %v", err)
	}
	defer file.Close()
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
			break
		}
		file.Write(req.Bytes[:len(req.Bytes)])
		fmt.Printf(
			"server recv\nMessage: %v\nTimestamp: %v\nLength: %v\n", req.Message, req.Timestamp, req.Length)
	}
	err = stream.SendAndClose(&echo.EchoResponse{Message: "server got your image"})
	return err
}

// BidirectionalStreamingEcho は送受信が同時に行われる双方向ストリーミング RPC。
//
// ゴルーチン1: クライアントから受信した画像チャンクをタイムスタンプ付きファイルに書き込む。
// ゴルーチン2: server.jpg を読み、チャンクをクライアントへ送信する。
// sync.WaitGroup で両方の処理が完了するまで待機する。
func (EchoServer) BidirectionalStreamingEcho(stream echo.EchoService_BidirectionalStreamingEchoServer) error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		filepath := "echo-server/file/" + strconv.FormatInt(time.Now().UnixMilli(), 10) + ".jpg"
		file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			file.Write(req.Bytes[:len(req.Bytes)])
			fmt.Printf("server recv\nMessage: %v\nTimestamp: %v\nLength: %v\n", req.Message, req.Timestamp, req.Length)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		filepath := "echo-server/file/server.jpg"
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
			stream.Send(&echo.EchoResponse{
				Message: "server send image",
				Bytes:   buf[:n],
			})
		}
	}()

	wg.Wait()
	return nil
}
