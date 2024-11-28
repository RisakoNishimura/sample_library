package sample_library

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"

	"github.com/quic-go/quic-go"
)

func Connect(address string) (quic.Connection, error) {
	// TLS設定。証明書の検証をスキップ（テスト環境用）。
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	// コンテキストを作成
	ctx := context.Background()

	// QUICサーバーに接続
	connection, err := quic.DialAddr(ctx, address, tlsConfig, nil) // nilはデフォルトのQUIC設定を指定
	if err != nil {
		return nil, fmt.Errorf("failed to dial server at %s: %v", address, err)
	}

	log.Println("Connected to server:", address)
	return connection, nil
}

func SendMessage(conn quic.Connection, message string) (string, error) {
	// コンテキストを作成
	ctx := context.Background()

	// ストリームをオープン
	stream, err := conn.OpenStreamSync(ctx) // コンテキストを渡してストリームを開く
	if err != nil {
		return "", fmt.Errorf("failed to open stream: %v", err)
	}
	defer stream.Close() // ストリームをクリーンに閉じる

	// サーバーにメッセージを送信
	_, err = stream.Write([]byte(message))
	if err != nil {
		return "", fmt.Errorf("failed to write message: %v", err)
	}
	log.Println("Client sent:", message)

	// サーバーからのレスポンスを受信
	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil {
		if err == io.EOF {
			log.Println("Server closed the stream")
		} else {
			return "", fmt.Errorf("failed to read response: %v", err)
		}
	}

	response := string(buffer[:n])
	log.Println("Client received:", response)
	return response, nil
}

// Disconnect は指定された接続を切断します。
func Disconnect(conn quic.Connection) error {
	err := conn.CloseWithError(0, "client closing") // 接続をクリーンに閉じる
	if err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}

	log.Println("Disconnected from server")
	return nil
}
