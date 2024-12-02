package sample_library

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"

	"github.com/quic-go/quic-go"
)

// QUICサーバーの待ち受けを開始する
func StartQUICServer(address, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("failed to load server certificate: %v", err)
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	listener, err := quic.ListenAddr(address, tlsConfig, nil)
	if err != nil {
		return fmt.Errorf("failed to listen on address %s: %v", address, err)
	}
	defer listener.Close()

	log.Println("QUIC server listening on", address)

	// クライアントからの接続を待ち受け
	for {
		connection, err := listener.Accept(context.Background())
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}

		// 接続を処理する
		go func(conn quic.Connection) {
			// 接続終了時にエラーコード 0 と理由 "connection closed" を使って接続を終了する
			defer conn.CloseWithError(0, "connection closed")

			// メッセージを複数回受け取るループ
			for {
				err := Accept(conn, "Message received successfully")
				if err != nil {
					log.Printf("error handling connection: %v", err)
					break // メッセージの受信中にエラーがあれば接続を終了
				}
			}
		}(connection)
	}
}

// クライアントからのメッセージを受け取り、応答を送信する
func Accept(connection quic.Connection, responseMessage string) error {
	// ストリームを受け入れる
	stream, err := connection.AcceptStream(context.Background())
	if err != nil {
		return fmt.Errorf("failed to accept stream: %v", err)
	}

	// メッセージを受信する
	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read from stream: %v", err)
	}

	if err == io.EOF {
		// クライアントがストリームを閉じた場合、終了
		log.Println("Client closed the stream.")
		return nil
	}

	message := string(buffer[:n])
	log.Println("Server received:", message)

	// クライアントへ応答を送信
	_, err = stream.Write([]byte(responseMessage))
	if err != nil {
		return fmt.Errorf("failed to write to stream: %v", err)
	}

	log.Println("Server sent:", responseMessage)

	// エラーなし
	return nil
}

// クライアントがサーバーに接続する(証明書はスキップ)
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

// クライアントがサーバーにメッセージを送信する
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
	// エラーコード 0 とメッセージ "client closing" で接続を閉じる
	err := conn.CloseWithError(0, "client closing")
	if err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}

	log.Println("Disconnected from server")
	return nil
}
