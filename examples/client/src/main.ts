import { NodepayClient } from '@nodepay/client';
import { ethers } from 'ethers';

const MERCHANT_API = 'http://localhost:8081';

const log = (msg: string) => {
    const el = document.getElementById('console');
    if (el) {
        el.innerText += `[${new Date().toLocaleTimeString()}] ${msg}\n`;
        el.scrollTop = el.scrollHeight;
    }
    console.log(msg);
};

// 1. Setup Wallet Provider (Metamask)
const getProvider = () => {
    if ((window as any).ethereum == null) {
        log("Error: Metamask not found");
        throw new Error("Metamask not found");
    }
    return new ethers.BrowserProvider((window as any).ethereum);
}

// 2. Setup Nodepay Client
const client = new NodepayClient({
    baseUrl: MERCHANT_API,
    wallet: {
        request: async (args: { method: string, params: any[] }) => {
            log(`Wallet Request: ${args.method}`);
            const provider = getProvider();
            return await provider.send(args.method, args.params);
        }
    }
});

async function buyWidget() {
    try {
        log("Starting Purchase Flow...");

        // Ensure wallet connected
        const provider = getProvider();
        const accounts = await provider.send("eth_requestAccounts", []);
        log(`Connected: ${accounts[0]}`);

        // A. Create Charge (Optional if we just hit protected resource directly)
        // For this demo, we can just hit the protected resource.
        // But adhering to flow: User -> Merchant (Create Charge) -> User -> Pay -> Merchant (Verify)
        log("Step 1: Requesting Product (Protected Resource)...");

        // This fetch call is intercepted by NodepayClient
        // If 402 is returned, it handles payment and retries
        const response = await client.fetch(`${MERCHANT_API}/product/1`);

        if (response.ok) {
            const data = await response.json();
            log(`Success! Response: ${JSON.stringify(data)}`);
            alert("Purchase Successful!\n" + data.data);
        } else {
            log(`Error: ${response.status} ${response.statusText}`);
        }

    } catch (err: any) {
        log(`Failed: ${err.message}`);
        console.error(err);
    }
}

document.getElementById('buyBtn')?.addEventListener('click', buyWidget);

log("System Ready. Connect Metamask and click Buy.");
