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
        // Get connected account
        const accounts = await this.wallet.request({ method: 'eth_requestAccounts' });
        const account = accounts[0];

        // Ensure we are on the correct network if specified
        if (req.network && req.network.startsWith('eip155:')) {
            const chainId = parseInt(req.network.split(':')[1]);
            const hexChainId = '0x' + chainId.toString(16);

            try {
                await this.wallet.request({
                    method: 'wallet_switchEthereumChain',
                    params: [{ chainId: hexChainId }],
                });
            } catch (error: any) {
                // Ethers.js or other providers might wrap the error.
                const code = error.code || error?.error?.code || error?.data?.originalError?.code;
                if (code === 4902 || (error.message && error.message.includes("4902"))) {
                    if (chainId === 84532) {
                        try {
                            await this.wallet.request({
                                method: 'wallet_addEthereumChain',
                                params: [{
                                    chainId: hexChainId,
                                    chainName: 'Base Sepolia',
                                    nativeCurrency: { name: 'ETH', symbol: 'ETH', decimals: 18 },
                                    rpcUrls: ['https://sepolia.base.org'],
                                    blockExplorerUrls: ['https://sepolia-explorer.base.org']
                                }],
                            });
                        } catch (addError) {
                            console.error('Failed to add network', addError);
                            throw new Error(`Please add network ${chainId} to your wallet manually.`);
                        }
                    } else {
                        throw new Error(`Network ${chainId} not configured in wallet.`);
                    }
                } else {
                    console.error('Failed to switch network', error);
                    throw error;
                }
            }
        }

        // EIP-3009 (TransferWithAuthorization) Signing
        const chainId = req.network ? parseInt(req.network.split(':')[1]) : 84532;

        const domain = {
            name: 'USDC',
            version: '2',
            chainId: chainId,
            verifyingContract: req.asset || '0x036CbD53842c5426634e7929541eC2318f3dCF7e', // USDC on Base Sepolia
        };

        const types = {
            EIP712Domain: [
                { name: 'name', type: 'string' },
                { name: 'version', type: 'string' },
                { name: 'chainId', type: 'uint256' },
                { name: 'verifyingContract', type: 'address' },
            ],
            TransferWithAuthorization: [
                { name: 'from', type: 'address' },
                { name: 'to', type: 'address' },
                { name: 'value', type: 'uint256' },
                { name: 'validAfter', type: 'uint256' },
                { name: 'validBefore', type: 'uint256' },
                { name: 'nonce', type: 'bytes32' },
            ],
        };

        const now = Math.floor(Date.now() / 1000);
        const validAfter = 0;
        const validBefore = now + 3600; // 1 hour
        // Generate random nonce (32 bytes)
        const nonce = '0x' + Array.from(crypto.getRandomValues(new Uint8Array(32)))
            .map(b => b.toString(16).padStart(2, '0')).join('');

        const message = {
            from: account,
            to: req.payTo,
            value: req.amount, // amount in base units (e.g. Wei/USDC atomic units)
            validAfter: validAfter,
            validBefore: validBefore,
            nonce: nonce,
        };

        const data = JSON.stringify({
            types: {
                EIP712Domain: types.EIP712Domain,
                TransferWithAuthorization: types.TransferWithAuthorization,
            },
            domain: domain,
            primaryType: 'TransferWithAuthorization',
            message: message,
        });

        const signature = await this.wallet.request({
            method: 'eth_signTypedData_v4',
            params: [account, data],
        });

        // Split Signature (r, s, v)
        // Sig is 0x + 65 bytes (130 hex chars)
        const stripped = signature.startsWith('0x') ? signature.slice(2) : signature;
        const r = '0x' + stripped.substring(0, 64);
        const s = '0x' + stripped.substring(64, 128);
        const v = parseInt(stripped.substring(128, 130), 16);

        // Construct final payload expected by Facilitator
        // Helper to convert to 0x-prefixed hex string for hexutil compatibility
        const toHex = (val: any) => {
            // Check if it's already a hex string
            if (typeof val === 'string' && val.startsWith('0x')) return val;
            const num = BigInt(val);
            return '0x' + num.toString(16);
        };

        const payload = {
            from: message.from,
            to: message.to,
            value: toHex(message.value),
            validAfter: toHex(message.validAfter),
            validBefore: toHex(message.validBefore),
            nonce: message.nonce,
            v: v,
            r: r,
            s: s,
        };

        return JSON.stringify(payload);
    }
}
