export interface PaymentRequired {
    x402Version: number;
    error?: string;
    resource: ResourceInfo;
    accepts: PaymentRequirements[];
    extensions?: Record<string, any>;
}

export interface ResourceInfo {
    url: string;
    description?: string;
    mimeType?: string;
}

export interface PaymentRequirements {
    scheme: string;
    network: string; // CAIP-2
    amount: string;
    asset: string; // Address
    payTo: string;
    maxTimeoutSeconds: number;
    extra?: any;
}

export interface PaymentPayload {
    x402Version: number;
    resource?: ResourceInfo;
    accepted: PaymentRequirements;
    payload: any;
    extensions?: Record<string, any>;
}
