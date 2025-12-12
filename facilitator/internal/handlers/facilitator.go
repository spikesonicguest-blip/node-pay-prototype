package handlers

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"nodepay-facilitator/internal/models"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Config holds the facilitator configuration
type Config struct {
	RPCURL         string
	PrivateKey     string // Hex string without 0x
	USDCAddress    string
	ChainID        int64
}

type Handler struct {
	Client *ethclient.Client
	Config Config
}

func New() *Handler {
	// Load from env, fallback to defaults/mock
	rpcURL := os.Getenv("FACILITATOR_RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://sepolia.base.org"
	}

	privateKey := strings.TrimSpace(os.Getenv("FACILITATOR_PRIVATE_KEY"))
	// If empty, remains empty -> triggers mock mode

	cfg := Config{
		RPCURL:      rpcURL,
		PrivateKey:  privateKey,
		USDCAddress: "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		ChainID:     84532,
	}

	client, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		fmt.Printf("Failed to connect to ETH client: %v\n", err)
		// For demo robustness, we don't panic here, but Settle will fail
	}

	return &Handler{
		Client: client,
		Config: cfg,
	}
}

// TransferWithAuthorization represents the EIP-3009 payload
type TransferWithAuthorization struct {
	From        common.Address `json:"from"`
	To          common.Address `json:"to"`
	Value       *hexutil.Big   `json:"value"`
	ValidAfter  *hexutil.Big   `json:"validAfter"`
	ValidBefore *hexutil.Big   `json:"validBefore"`
	Nonce       hexutil.Bytes  `json:"nonce"`
	V           uint8          `json:"v"`
	R           hexutil.Bytes  `json:"r"`
	S           hexutil.Bytes  `json:"s"`
}

var (
	// EIP-712 Type Hashes
	EIP712DomainTypeHash = crypto.Keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))
	AuthorizationTypeHash = crypto.Keccak256([]byte("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)"))
	
	// USDC TransferWithAuthorization ABI
	usdcABI = `[{"constant":false,"inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"value","type":"uint256"},{"name":"validAfter","type":"uint256"},{"name":"validBefore","type":"uint256"},{"name":"nonce","type":"bytes32"},{"name":"v","type":"uint8"},{"name":"r","type":"bytes32"},{"name":"s","type":"bytes32"}],"name":"transferWithAuthorization","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`
)

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.PaymentPayload.X402Version != 2 {
		respondVerifyError(w, "unsupported_version", "Only x402 version 2 is supported")
		return
	}

	// 1. Unmarshal Authorization Payload
	payloadBytes, err := json.Marshal(req.PaymentPayload.Payload)
	if err != nil {
		respondVerifyError(w, "invalid_payload", "Failed to marshal payload")
		return
	}

	var auth TransferWithAuthorization
	if err := json.Unmarshal(payloadBytes, &auth); err != nil {
		respondVerifyError(w, "invalid_payload_structure", "Payload must be a valid EIP-3009 object")
		return
	}

	// 2. Recover Signer (EIP-3009 / EIP-712)
	chainID := parseBigInt("84532") // Base Sepolia
	usdcAddress := common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e")

	pad32 := func(i *hexutil.Big) []byte {
		return common.LeftPadBytes((*big.Int)(i).Bytes(), 32)
	}

	domainSeparator := crypto.Keccak256(
		EIP712DomainTypeHash,
		crypto.Keccak256([]byte("USDC")),
		crypto.Keccak256([]byte("2")),
		common.LeftPadBytes(chainID.Bytes(), 32),
		common.LeftPadBytes(usdcAddress.Bytes(), 32),
	)

	structHash := crypto.Keccak256(
		AuthorizationTypeHash,
		common.LeftPadBytes(auth.From.Bytes(), 32),
		common.LeftPadBytes(auth.To.Bytes(), 32),
		pad32(auth.Value),
		pad32(auth.ValidAfter),
		pad32(auth.ValidBefore),
		common.RightPadBytes(auth.Nonce, 32),
	)

	digest := crypto.Keccak256(
		[]byte{0x19, 0x01},
		domainSeparator,
		structHash,
	)

	v := auth.V
	if v >= 27 { v -= 27 }
	sig := make([]byte, 65)
	copy(sig[0:32], auth.R)
	copy(sig[32:64], auth.S)
	sig[64] = v

	pubKey, err := crypto.SigToPub(digest, sig)
	recoveredAddr := common.Address{}
	if err == nil {
		recoveredAddr = crypto.PubkeyToAddress(*pubKey)
	}

	if err != nil || recoveredAddr != auth.From {
		respondVerifyError(w, "invalid_signature", fmt.Sprintf("Recovered %s, expected %s", recoveredAddr.Hex(), auth.From.Hex()))
		return
	}

	// 3. Logic Validation (Timestamps, Amounts, Recipient)
	now := time.Now().Unix()
	if auth.ValidBefore != nil {
		validBefore := (*big.Int)(auth.ValidBefore).Int64()
		if now >= validBefore {
			respondVerifyError(w, "authorization_expired", "Authorization is no longer valid")
			return
		}
	}
	if auth.ValidAfter != nil {
		validAfter := (*big.Int)(auth.ValidAfter).Int64()
		if now < validAfter {
			respondVerifyError(w, "authorization_not_yet_valid", "Authorization is not yet valid")
			return
		}
	}
	
	payloadAmount := (*big.Int)(auth.Value)
	reqAmount, _ := new(big.Int).SetString(req.PaymentRequirements.Amount, 10)
	if reqAmount != nil && payloadAmount.Cmp(reqAmount) < 0 {
		respondVerifyError(w, "insufficient_amount", "Authorized amount is less than required")
		return
	}

	if auth.To.Hex() != req.PaymentRequirements.PayTo {
		respondVerifyError(w, "invalid_recipient", fmt.Sprintf("Expected %s, got %s", req.PaymentRequirements.PayTo, auth.To.Hex()))
		return
	}

	// 4. On-Chain Verification (Balance + Simulation)
	if h.Client != nil {
		// A. Balance Check
		// 70a08231 = balanceOf(address)
		balanceInput := common.Hex2Bytes("70a08231")
		balanceInput = append(balanceInput, common.LeftPadBytes(auth.From.Bytes(), 32)...)
		
		contractAddr := common.HexToAddress(h.Config.USDCAddress)
		msg := ethereum.CallMsg{
			To:   &contractAddr,
			Data: balanceInput,
		}
		
		resultBytes, err := h.Client.CallContract(context.Background(), msg, nil)
		if err == nil && len(resultBytes) == 32 {
			balance := new(big.Int).SetBytes(resultBytes)
			if balance.Cmp(payloadAmount) < 0 {
				respondVerifyError(w, "insufficient_funds", fmt.Sprintf("User balance %s < required %s", balance.String(), payloadAmount.String()))
				return
			}
		}

		// B. Transaction Simulation
		authInput, err := packTransferWithAuthorization(auth)
		if err == nil {
			simMsg := ethereum.CallMsg{
				From: auth.From,
				To:   &contractAddr,
				Data: authInput,
			}
			_, err := h.Client.CallContract(context.Background(), simMsg, nil)
			if err != nil {
				respondVerifyError(w, "simulation_failed", fmt.Sprintf("Transaction simulation failed: %v", err))
				return
			}
		}
	}

	resp := models.VerifyResponse{
		IsValid: true,
		Payer:   auth.From.Hex(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ... helper functions ...

func (h *Handler) Settle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Unmarshal Authorization Payload
	payloadBytes, _ := json.Marshal(req.PaymentPayload.Payload)
	var auth TransferWithAuthorization
	if err := json.Unmarshal(payloadBytes, &auth); err != nil {
		http.Error(w, "Invalid payload structure", http.StatusBadRequest)
		return
	}

	// 2. Prepare Transaction
	input, err := packTransferWithAuthorization(auth)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to pack arguments: %v", err), http.StatusInternalServerError)
		return
	}

	// 3. Create and Sign Transaction
	ctx := context.Background()
	privateKey, err := crypto.HexToECDSA(h.Config.PrivateKey)
	if err != nil {
		fmt.Printf("Invalid private key: %v. Key: '%s' (len: %d)\n", err, h.Config.PrivateKey, len(h.Config.PrivateKey))
		http.Error(w, "Invalid private key configuration", http.StatusInternalServerError)
		return
	}

	// Real Execution Mode
	if h.Client == nil {
		http.Error(w, "Ethereum client not initialized", http.StatusInternalServerError)
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		http.Error(w, "Error casting public key", http.StatusInternalServerError)
		return
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonceUint, err := h.Client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		http.Error(w, "Failed to get nonce", http.StatusInternalServerError)
		return
	}

	gasPrice, err := h.Client.SuggestGasPrice(ctx)
	if err != nil {
		http.Error(w, "Failed to get gas price", http.StatusInternalServerError)
		return
	}

	contractAddr := common.HexToAddress(h.Config.USDCAddress)
	
	// Create Tx
	tx := types.NewTransaction(
		nonceUint,
		contractAddr,
		big.NewInt(0), // Value (ETH) is 0
		300000,        // Gas Limit (estimated)
		gasPrice,
		input,
	)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(big.NewInt(h.Config.ChainID)), privateKey)
	if err != nil {
		http.Error(w, "Failed to sign tx", http.StatusInternalServerError)
		return
	}

	// 4. Broadcast
	err = h.Client.SendTransaction(ctx, signedTx)
	if err != nil {
		// If broadcast fails (funds, etc), return error
		http.Error(w, fmt.Sprintf("Failed to send tx: %v", err), http.StatusInternalServerError)
		return
	}
	
	fmt.Printf("[Facilitator] Broadcasted Tx: %s\n", signedTx.Hash().Hex())

	resp := models.SettleResponse{
		Success:     true,
		Payer:       auth.From.Hex(),
		Transaction: signedTx.Hash().Hex(),
		Network:     req.PaymentRequirements.Network,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func parseBigInt(s string) *big.Int {
	n, _ := new(big.Int).SetString(s, 10)
	return n
}

func (h *Handler) Supported(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := models.SupportedResponse{
		Kinds: []models.SupportedKind{
			{
				X402Version: 2,
				Scheme:      "exact",
				Network:     "eip155:84532",
			},
			{
				X402Version: 2,
				Scheme:      "exact",
				Network:     "eip155:1",
			},
		},
		Extensions: []string{},
		Signers: map[string][]string{
			"eip155:*": {"0xFacilitatorPublicKey"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func respondVerifyError(w http.ResponseWriter, reason string, logMsg string) {
	fmt.Printf("[Facilitator] Verify Failed: %s - %s\n", reason, logMsg)
	resp := models.VerifyResponse{
		IsValid:       false,
		InvalidReason: reason,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func packTransferWithAuthorization(auth TransferWithAuthorization) ([]byte, error) {
	parsedABI, err := abi.JSON(strings.NewReader(usdcABI))
	if err != nil {
		return nil, err
	}

	v := auth.V
	if v >= 27 { v -= 27 }
	
	var nonce [32]byte
	copy(nonce[:], auth.Nonce)
	var rVal [32]byte
	copy(rVal[:], auth.R)
	var sVal [32]byte
	copy(sVal[:], auth.S)

	return parsedABI.Pack("transferWithAuthorization",
		auth.From,
		auth.To,
		(*big.Int)(auth.Value),
		(*big.Int)(auth.ValidAfter),
		(*big.Int)(auth.ValidBefore),
		nonce,
		v + 27,
		rVal,
		sVal,
	)
}

