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

// OpenStream はQUICコネクションに対してストリームを開きます。
func OpenStream(conn quic.Connection) (quic.Stream, error) {
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	log.Println("Stream successfully opened")
	return stream, nil
}

// WriteToStream は指定されたストリームにメッセージを書き込みます。
func WriteToStream(stream quic.Stream, message string) error {
	_, err := stream.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write to stream: %w", err)
	}

	log.Printf("Message sent: %s", message)
	return nil
}

// ReadFromStream は指定されたストリームからメッセージを読み取ります。
func ReadFromStream(stream quic.Stream) (string, error) {
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
	log.Printf("Message received: %s", message)
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
