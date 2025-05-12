import csv

def summarize(path):
    total = 0
    sum_rtt = 0.0
    lost = 0
    ord_err = 0

    with open(path, newline='') as f:
        reader = csv.DictReader(f)
        for row in reader:
            total   += 1
            # read the float seconds column
            sum_rtt += float(row['rtt_s'])
            lost     += int(row['lost'])
            ord_err  += int(row['order_error'])

    avg_rtt   = sum_rtt / total if total else 0.0
    loss_pct  = lost / total * 100 if total else 0.0
    ord_pct   = ord_err / total * 100 if total else 0.0

    return total, avg_rtt, loss_pct, ord_pct

# summarize both runs
base = summarize('metrics_baseline.csv')
imp  = summarize('metrics_impaired.csv')

# print side-by-side
print(f"{'Metric':<15}{'Baseline':>12}{'Impaired':>12}")
print(f"{'-'*39}")
print(f"{'total_msgs':<15}{base[0]:>12}{imp[0]:>12}")
print(f"{'avg_rtt_s':<15}{base[1]:>12.6f}{imp[1]:>12.6f}")
print(f"{'loss_%':<15}{base[2]:>11.1f}%{imp[2]:>11.1f}%")
print(f"{'order_err_%':<15}{base[3]:>11.1f}%{imp[3]:>11.1f}%")