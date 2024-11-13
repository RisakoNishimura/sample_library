package sample_library

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"

	"github.com/quic-go/quic-go"
)

func ConnectToQUICServer(address string) error {
	// TLS設定。証明書の検証をスキップ（テスト環境用）。
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	// コンテキストを作成
	ctx := context.Background()

	// QUICサーバーに接続
	connection, err := quic.DialAddr(ctx, address, tlsConfig, nil) // nilはデフォルトのQUIC設定を指定
	if err != nil {
		return fmt.Errorf("failed to dial server at %s: %v", address, err)
	}
	defer connection.CloseWithError(0, "client closing") // 接続をクリーンに閉じる

	// ストリームをオープン
	stream, err := connection.OpenStreamSync(ctx) // コンテキストを渡してストリームを開く
	if err != nil {
		return fmt.Errorf("failed to open stream: %v", err)
	}

	// サーバーにメッセージを送信
	message := "Hello, QUIC server!"
	_, err = stream.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}
	log.Println("Client sent:", message)

	// サーバーからのレスポンスを受信
	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil {
		if err == io.EOF {
			log.Println("Server closed the stream")
		} else {
			return fmt.Errorf("failed to read response: %v", err)
		}
	}
	log.Println("Client received:", string(buffer[:n]))

	return nil
}
