export function formatUSD(value: string | number): string {
  const n = typeof value === "string" ? parseFloat(value) : value;
  return n.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function formatCrypto(value: string | number, decimals = 4): string {
  const n = typeof value === "string" ? parseFloat(value) : value;
  return n.toLocaleString("en-US", {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

export function formatPnL(value: string | number): string {
  const n = typeof value === "string" ? parseFloat(value) : value;
  const sign = n >= 0 ? "+" : "";
  return `${sign}${formatUSD(n)}`;
}

export function formatPercent(value: string | number): string {
  const n = typeof value === "string" ? parseFloat(value) : value;
  return `${(n * 100).toFixed(2)}%`;
}

export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}
