# NodePay SDK Specification

This document defines the interface and behavior of the official NodePay Client and Server SDKs. These SDKs abstract the complexity of the x402 protocol, making it easy for developers to integrate payment walls and agentic payments.

## 1. Server SDK

The Server SDK is responsible for protecting resources, determining pricing, communication with the Facilitator, and verifying payments/identities.

### 1.1. Core Middleware `NodepayMiddleware`

The core component is an HTTP middleware (e.g., specific for Express, Koa, or standard `http.Server`).

#### Configuration Interface

```typescript
interface NodepayConfig {
  /**
   * The NodePay API Key (connects to Merchant Dashboard/Facilitator)
   */
  apiKey: string;

  /**
   * Configuration for the Facilitator connection
   */
  facilitator: {
    url: string; // e.g. "https://api.nodepay.ai/v1"
  };

  /**
   * Configuration for where to receive payments.
   */
  settlement?: {
    address: string; // The merchant's wallet address
    network?: string; // e.g. "eip155:84532"
    asset?: string; // e.g. "0x..."
  };

  /**
   * Default pricing rules. Can be overridden per route.
   */
  defaultPricing?: PricingRule;
}

interface PricingRule {
  amount: string; // "0.01"
  currency: string; // "USD"
}
```

#### Middleware Usage (Express Example)

```typescript
// Import
import { nodepay } from '@nodepay/server';

// Initialize
const paywall = nodepay({
  apiKey: "sk_...",
  facilitator: { url: "..." }
});

// Use globally or per-route
app.get('/premium-content', 
  paywall.requirePayment({ amount: "1.00", currency: "USD" }), 
  (req, res) => {
    res.json({ data: "Premium Data" });
  }
);
```

### 1.2. Internal Logic Flow

1.  **Receive Request**: Middleware intercepts incoming request.
2.  **Check Headers**:
    *   If `SIGN-IN-WITH-X` is present: Verify signature against Facilitator/Internal Cache. If valid, call `next()`.
    *   If `PAYMENT-SIGNATURE` is present: Verify payment transaction on-chain (via Facilitator). If verified, call `next()`.
3.  **Payment Required**:
    *   If no valid headers, create a `Charge` via Facilitator API.
    *   Respond with `402 Payment Required`.
    *   Header `PAYMENT-REQUIRED`: JSON string of the Charge.

---

## 2. Client SDK

The Client SDK allows web frontends and AI agents to consume x402-protected resources effortlessly.

### 2.1. `NodepayClient`

#### Interface

```typescript
interface ClientConfig {
  /**
   * Wallet provider (e.g., EIP-1193 window.ethereum, or a private key signer)
   */
  wallet: WalletProvider;
}

class NodepayClient {
  constructor(config: ClientConfig);

  /**
   * A wrapper around the standard fetch API.
   * Automatically handles 402 responses, signing, and payments.
   */
  async fetch(url: string, init?: RequestInit): Promise<Response>;

  /**
   * Explicitly login to establish a session without making a request first.
   */
  async login(url: string): Promise<void>;
}
```

### 2.2. `fetch` Interceptor Logic

1.  **Initial Request**: Send request as normal.
2.  **Handle response**:
    *   **200 OK**: Return response.
    *   **402 Payment Required**:
        1.  Parse `PAYMENT-REQUIRED` header.
        2.  **Strategy Check**:
            *   *Can I sign in?* Check if wallet address has a valid subscription/session. Attempt to sign challenge.
            *   *Must I pay?* Prompt user (or auto-pay if agent) to pay the `amount` to `pay_to` address.
        3.  **Action**:
            *   If Signing: Sign message, retry request with `SIGN-IN-WITH-X`.
            *   If Paying: Broadcast transaction, retry request with `PAYMENT-SIGNATURE` (or wait for confirmation).
3.  **Retry**: Send the request again with the new credentials.

### 2.3. Agentic Usage

For AI agents, the `wallet` provider can be a programmatic signer with a spending limit.

```typescript
const agent = new NodepayClient({
  wallet: new PrivateKeyWallet(process.env.PRIVATE_KEY) // Dangerous in browser, safe in backend-for-frontend
});

// This will auto-pay if the budget permits
const response = await agent.fetch('https://api.example.com/data');
```

## 3. Discovery Protocol

The SDKs also implement the Discovery protocol defined in `discovery.yaml`.

*   **Server**: Automatically adds `GET /v1/discovery` to the app to expose pricing and assets.
*   **Client**: Before first request, optional call to `/v1/discovery` to check supported chains (e.g., switch wallet network to Base if required).
