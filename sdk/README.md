# NodePay Client SDK (@nodepay/client)

The official TypeScript SDK for integrating NodePay and x402 payments into web applications and agents.

## Features

-   **x402 Client**: A wrapper around `fetch` that automatically handles `402 Payment Required` responses.
-   **Auto-Payment**: Intercepts 402s, requests user signature/payment via the provided wallet, and retries the request with `Payment-Signature`.
-   **Type Safety**: Full TypeScript definitions for x402 wire protocol types (`PaymentRequired`, `PaymentPayload`).

## Installation

```bash
npm install @nodepay/client
```

## Usage

```typescript
import { NodepayClient } from '@nodepay/client';

// 1. Initialize with a Wallet Provider (e.g., window.ethereum)
const client = new NodepayClient({
  wallet: window.ethereum
});

// 2. Make requests as normal
// The client handles the 402 flow under the hood
async function getData() {
  try {
    const res = await client.fetch('http://localhost:8080/premium-data');
    const data = await res.json();
    console.log(data);
  } catch (err) {
    console.error("Payment failed", err);
  }
}
```

## Development

### Build

```bash
npm run build
```

### Test

```bash
npm test
```
