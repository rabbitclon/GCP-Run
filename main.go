package main

import (
 "io"
 "log"
 "net"
 "os"
 "time"
)

func closeWrite(c net.Conn) {
 if tcp, ok := c.(*net.TCPConn); ok {
  _ = tcp.CloseWrite()
 } else {
  _ = c.Close()
 }
}

func proxy(client net.Conn, target string) {
 defer client.Close()

 dialer := net.Dialer{
  Timeout:   10 * time.Second,
  KeepAlive: 30 * time.Second,
 }

 remote, err := dialer.Dial("tcp", target)
 if err != nil {
  log.Printf("connect to target failed: %v", err)
  return
 }
 defer remote.Close()

 done := make(chan struct{}, 2)

 go func() {
  _, _ = io.Copy(remote, client)
  closeWrite(remote)
  done <- struct{}{}
 }()

 go func() {
  _, _ = io.Copy(client, remote)
  closeWrite(client)
  done <- struct{}{}
 }()

 <-done
}

func main() {
 port := os.Getenv("PORT")
 if port == "" {
  port = "8080"
 }

 ip := os.Getenv("V2RAY_SERVER_IP")
 if ip == "" {
  log.Fatal("V2RAY_SERVER_IP is empty")
 }

 targetPort := os.Getenv("TARGET_PORT")
 if targetPort == "" {
  targetPort = "80"
 }

 target := net.JoinHostPort(ip, targetPort)

 listener, err := net.Listen("tcp", ":"+port)
 if err != nil {
  log.Fatalf("listen failed: %v", err)
 }
 defer listener.Close()

 log.Printf("proxy listening on :%s -> %s", port, target)

 for {
  conn, err := listener.Accept()
  if err != nil {
   log.Printf("accept failed: %v", err)
   continue
  }
  go proxy(conn, target)
 }
}
