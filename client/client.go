// Client for a simple TCP echo server with test mode

package main

import (
    "bufio"
    "encoding/csv"
    "flag"
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"
    "time"
)

var (
    host         = flag.String("host", "127.0.0.1", "server host")
    port         = flag.String("port", "9000", "server port")
    name         = flag.String("name", "anon", "client display name")
    testMode     = flag.Bool("test", false, "run automated test sequence")
    testCount    = flag.Int("count", 100, "number of messages to send in test mode")
    testInterval = flag.Duration("interval", 1*time.Second, "interval between test messages")
    timeout      = flag.Duration("timeout", 5*time.Second, "read timeout per message")
    csvPath      = flag.String("csv", "metrics.csv", "output CSV file for test metrics")
)

func main() {
    flag.Parse()
    addr := *host + ":" + *port
    if *testMode {
        runTest(addr, *name)
    } else {
        startClient(addr, *name)
    }
}

func startClient(addr, name string) {
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Dial error:", err)
        return
    }
    defer conn.Close()
    logf("Connected to %s", addr)

    go func() {
        scanner := bufio.NewScanner(conn)
        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
        if err := scanner.Err(); err != nil {
            fmt.Fprintln(os.Stderr, "Receive error:", err)
        }
        os.Exit(0)
    }()

    input := bufio.NewScanner(os.Stdin)
    for input.Scan() {
        text := strings.TrimSpace(input.Text())
        if text == "" {
            continue
        }
        msg := fmt.Sprintf("%s: %s", name, text)
        if _, err := conn.Write([]byte(msg + "\n")); err != nil {
            fmt.Fprintln(os.Stderr, "Send error:", err)
            break
        }
    }
}

type metric struct {
    seq      int
    rttSec   float64
    lost     bool
    orderErr bool
}

func runTest(addr, name string) {
    fmt.Printf("Running test mode: %d messages to %s every %s (timeout %s)\n",
        *testCount, addr, testInterval.String(), timeout.String())

    conn, err := net.Dial("tcp", addr)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Dial error:", err)
        return
    }
    defer conn.Close()

    reader := bufio.NewReader(conn)
    var metrics []metric
    expected := 1

    for i := 1; i <= *testCount; i++ {
        t0 := time.Now()
        payload := fmt.Sprintf("%d|%d", i, t0.UnixNano())
        if _, err := conn.Write([]byte(payload + "\n")); err != nil {
            fmt.Fprintf(os.Stderr, "Send error at #%d: %v\n", i, err)
            metrics = append(metrics, metric{i, 0, true, false})
            continue
        }

        conn.SetReadDeadline(time.Now().Add(*timeout))
        line, err := reader.ReadString('\n')
        if err != nil {
            fmt.Fprintf(os.Stderr, "Timeout/loss at #%d\n", i)
            metrics = append(metrics, metric{i, 0, true, false})
        } else {
            idx := strings.LastIndex(line, ": ")
            if idx < 0 {
                metrics = append(metrics, metric{i, 0, false, true})
            } else {
                recv := strings.TrimSpace(line[idx+2:])
                sub := strings.SplitN(recv, "|", 2)
                if len(sub) != 2 {
                    metrics = append(metrics, metric{i, 0, false, true})
                } else {
                    seqRecv, err1 := strconv.Atoi(sub[0])
                    sendNano, err2 := strconv.ParseInt(sub[1], 10, 64)
                    if err1 != nil || err2 != nil {
                        metrics = append(metrics, metric{i, 0, false, true})
                    } else {
                        rtt := time.Since(time.Unix(0, sendNano))
                        orderErr := seqRecv != expected
                        metrics = append(metrics, metric{i, rtt.Seconds(), false, orderErr})
                        expected = seqRecv + 1
                    }
                }
            }
        }

        time.Sleep(*testInterval)
    }

    file, err := os.Create(*csvPath)
    if err != nil {
        fmt.Fprintln(os.Stderr, "CSV create error:", err)
        return
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()
    // note header changed to rtt_s
    writer.Write([]string{"sequence", "rtt_s", "lost", "order_error"})
    for _, m := range metrics {
        lostFlag := "0"
        if m.lost {
            lostFlag = "1"
        }
        ordFlag := "0"
        if m.orderErr {
            ordFlag = "1"
        }
        writer.Write([]string{
            strconv.Itoa(m.seq),
            // format with 6 decimal places
            strconv.FormatFloat(m.rttSec, 'f', 6, 64),
            lostFlag,
            ordFlag,
        })
    }

    fmt.Println("Test complete — metrics written to", *csvPath)
}

func logf(format string, args ...interface{}) {
    ts := time.Now().Format(time.RFC3339)
    fmt.Printf("[%s] %s\n", ts, fmt.Sprintf(format, args...))
}
