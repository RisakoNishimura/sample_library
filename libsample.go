package sample_library

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/quic-go/quic-go"
)

// StartServer はQUICサーバーを起動し、リスナーを返します。
func StartServer(address, certFile, keyFile string) (*quic.Listener, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13, // 最低TLSバージョンをTLS1.3に設定
	}

	listener, err := quic.ListenAddr(address, tlsConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on address %s: %w", address, err)
	}

	log.Printf("QUIC server started and listening on %s", address)
	return listener, nil
}

// StopServer はQUICサーバーを停止し、リスナーを閉じます。
func StopServer(listener *quic.Listener) error {
	// リスナーがnilでないことを確認
	if listener == nil {
		return fmt.Errorf("listener is nil, cannot stop server")
	}

	// リスナーを閉じる
	err := (*listener).Close()
	if err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}

	log.Println("QUIC server stopped")
	return nil
}

// ConnectToServer は指定されたアドレスにQUIC接続を行います。
func ConnectToServer(address string) (quic.Connection, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // テスト用。運用では信頼済み証明書を使用
		MinVersion:         tls.VersionTLS13,
	}

	conn, err := quic.DialAddr(context.Background(), address, tlsConfig, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server at %s: %w", address, err)
	}

	log.Printf("Connected to server at %s", address)
	return conn, nil
}

// SendMessage は指定された接続のストリームにメッセージを書き込む関数です。
func SendMessage(conn quic.Connection, message string) error {
	// ストリームを開く
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}

	// メッセージを送信
	_, err = stream.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write to stream: %w", err)
	}

	log.Printf("Sent message: %s", message)

	// ストリームを閉じる
	err = stream.Close()
	if err != nil {
		log.Printf("Failed to close stream: %v", err)
	}

	return nil
}

// ReceiveMessage は指定された接続のストリームからメッセージを受信する関数です。
func ReceiveMessage(conn quic.Connection) (string, error) {
	// ストリームを受け入れる
	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to accept stream: %w", err)
	}

	// ストリームからデータを読み取る
	var builder strings.Builder
	buffer := make([]byte, 1024)

	for {
		n, err := stream.Read(buffer)
		if err != nil && err != io.EOF {
			return "", fmt.Errorf("failed to read from stream: %w", err)
		}
		builder.Write(buffer[:n])

		if err == io.EOF {
			break
		}
	}

	message := builder.String()
	log.Printf("Received message: %s", message)

	// ストリームを閉じる
	err = stream.Close()
	if err != nil {
		log.Printf("Failed to close stream: %v", err)
	}

	return message, nil
}

// GetMessageFromInput はキーボード入力からメッセージを取得し、そのまま返します。
func GetMessageFromInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your message: ") // 固定プロンプト

	// ユーザーの入力を取得
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	// 入力のトリミング（前後の空白や改行を削除）
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("input cannot be empty")
	}

	return input, nil
}

// CloseServer はQUICサーバーを閉じます。
func CloseServer(listener quic.Listener) error {
	err := listener.Close()
	if err != nil {
		return fmt.Errorf("failed to close server: %w", err)
	}

	log.Println("QUIC server closed successfully")
	return nil
}
