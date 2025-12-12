import { PaymentRequired } from './types';

// Simple wallet interface (e.g. window.ethereum)
export interface WalletProvider {
    request: (args: { method: string; params?: any[] }) => Promise<any>;
}

export class NodepayClient {
    private wallet: WalletProvider;

    constructor(config: { wallet: WalletProvider }) {
        this.wallet = config.wallet;
    }

    async fetch(input: RequestInfo, init?: RequestInit): Promise<Response> {
        // 1. Initial Request
        const response = await fetch(input, init);

        // 2. Check for 402
        if (response.status === 402) {
            const requirements = (await response.json()) as PaymentRequired;

            // 3. Strategy: SignIn vs Pay
            // For simplicity, we default to paying first option in 'accepts'
            // In a real agent, we would check if we already have a session

            const req = requirements.accepts[0];
            if (req.scheme === 'exact') {
                const signature = await this.pay(req);

                // 4. Retry with Payment-Signature
                const headers = new Headers(init?.headers);
                headers.set('Payment-Signature', signature);

                return fetch(input, {
                    ...init,
                    headers
                });
            }
        }

        return response;
    }

    async pay(req: any): Promise<string> {
        // Mock payment logic for 'exact' scheme
        // 1. Request user signature via wallet
        // 2. Broadcast transaction
        // 3. Return signature or tx hash

        // Example: sending transaction
        const txHash = await this.wallet.request({
            method: 'eth_sendTransaction',
            params: [{
                to: req.payTo,
                value: Number(req.amount).toString(16), // Convert to hex
                from: '0xUser', // Should get from wallet
            }]
        });

        return txHash;
    }
}
