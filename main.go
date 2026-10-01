package main

import (
 "io"
 "log"
 "net"
 "os"
 "time"
)

const bufferSize = 64 * 1024

func tuneTCP(c net.Conn) {
 if tcp, ok := c.(*net.TCPConn); ok {
  _ = tcp.SetNoDelay(true)
  _ = tcp.SetKeepAlive(true)
  _ = tcp.SetKeepAlivePeriod(30 * time.Second)
 }
}

func closeWrite(c net.Conn) {
 if tcp, ok := c.(*net.TCPConn); ok {
  _ = tcp.CloseWrite()
 } else {
  _ = c.Close()
 }
}

func copyData(dst net.Conn, src net.Conn, done chan<- struct{}) {
 buf := make([]byte, bufferSize)
 _, _ = io.CopyBuffer(dst, src, buf)
 closeWrite(dst)
 done <- struct{}{}
}

func proxy(client net.Conn, target string) {
 defer client.Close()
 tuneTCP(client)

 dialer := net.Dialer{
  Timeout:   5 * time.Second,
  KeepAlive: 30 * time.Second,
 }

 remote, err := dialer.Dial("tcp", target)
 if err != nil {
  log.Printf("connect to target failed: %v", err)
  return
 }
 defer remote.Close()
 tuneTCP(remote)

 done := make(chan struct{}, 2)

 go copyData(remote, client, done)
 go copyData(client, remote, done)

 <-done
 <-done
}

func main() {
 port := os.Getenv("PORT")
 if port == "" {
  port = "8080"
 }

 ip := os.Getenv("SERVER_IP")
 if ip == "" {
  log.Fatal("SERVER_IP is empty")
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
