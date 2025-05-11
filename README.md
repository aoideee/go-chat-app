# Go TCP Chat App (Phase 3 Testing)

A minimal multi-client TCP chat application in Go, now extended with an automated testing mode for Phase 3: Testing & Data Collection.

## Automated Testing Mode

Phase 3 uses two test runs—**baseline** (no impairment) and **impaired** (with packet delay & loss).

### 0. Server

```bash
go run main.go -mode server -port 9000
```

### 1. Baseline Test

```bash
go run main.go \
  -mode client \
  -host 127.0.0.1 \
  -port 9000 \
  -name Tysha \
  -test \
  -count 100 \
  -interval 500ms \
  -timeout 5s \
  -csv metrics_baseline.csv
```

After completion:
```
Test complete — metrics written to metrics_baseline.csv
```

### 2. Impair the Network

In another terminal, add packet delay and loss with `tc`:
```bash
sudo tc qdisc add dev lo root netem delay 100ms loss 5%
```

### 3. Impaired Test

Back in the client terminal:
```bash
go run main.go \
  -mode client \
  -host 127.0.0.1 \
  -port 9000 \
  -name Tysha \
  -test \
  -count 100 \
  -interval 500ms \
  -timeout 5s \
  -csv metrics_impaired.csv
```

### 4. Remove the Impairment

When done testing:
```bash
sudo tc qdisc del dev lo root
```

## Comparing Metrics

Use the provided Python script to print a side‐by-side summary:

```bash
python3 compare_metrics.py
```

Or, if you prefer:
```bash
./compare_metrics.py   # ensure it has a shebang and is executable
```

You'll see metrics for `baseline` vs `impaired` runs (total messages, avg RTT, loss %, order error %).