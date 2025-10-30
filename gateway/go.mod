module gateway

go 1.25

require (
	github.com/gorilla/mux v1.8.1
	ledger v0.0.0
)

replace ledger => ../ledger
