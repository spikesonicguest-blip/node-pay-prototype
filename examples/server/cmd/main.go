	"nodepay-example-merchant/internal/handlers"
	"nodepay-example-merchant/internal/store"
	
	"nodepay-go-sdk/client"
	"nodepay-go-sdk/paywall"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	s := store.New()
	h := handlers.New(s)
	
	// Initialize SDK Client
	x402Client := client.New("http://localhost:8080")
	
	// Initialize SDK Paywall
	pw := paywall.New(x402Client)

	mux := http.NewServeMux()

	// Public Charge Creation
	mux.HandleFunc("/checkout", h.CreateCharge)

	// Protected Product
	productMux := http.NewServeMux()
	productMux.HandleFunc("/product/1", h.Product)
	
	// Apply SDK Paywall
	// 5 cents, Base Sepolia, USDC
	paymentConfig := paywall.PaymentConfig{
		Amount:      "50000", // 0.05 * 10^6
		Currency:    "USDC",
		Network:     "eip155:84532",
		Asset:       "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
		PayTo:       "0xMerchantWallet",
		Description: "Digital Widget",
	}
	
	mux.Handle("/product/1", pw.Handler(paymentConfig)(productMux))

	log.Printf("Merchant Server starting on :%s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
