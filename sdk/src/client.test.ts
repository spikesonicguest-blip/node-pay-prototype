import { NodepayClient } from './client';
import { PaymentRequired } from './types';

describe('NodepayClient', () => {
    let client: NodepayClient;
    let mockWallet: any;

    beforeEach(() => {
        mockWallet = { request: jest.fn().mockResolvedValue('0xSig') };
        client = new NodepayClient({ wallet: mockWallet });
        global.fetch = jest.fn();
    });

    it('should handle 402 and retry with payment signature', async () => {
        // 1. Mock 402 Response
        const paymentReq: PaymentRequired = {
            x402Version: 2,
            resource: { url: 'http://api.com' },
            accepts: [{
                scheme: 'exact',
                network: 'base',
                amount: '100',
                asset: '0x123',
                payTo: '0xMerch',
                maxTimeoutSeconds: 60
            }]
        };

        (global.fetch as jest.Mock)
            .mockResolvedValueOnce({
                status: 402,
                json: async () => paymentReq
            })
            .mockResolvedValueOnce({
                status: 200,
                json: async () => ({ success: true })
            });

        // 2. Execute
        const res = await client.fetch('http://api.com/data');

        // 3. Verify
        expect(res.status).toBe(200);
        expect(mockWallet.request).toHaveBeenCalled(); // Should have paid
        expect(global.fetch).toHaveBeenCalledTimes(2); // Initial + Retry
    });
});
