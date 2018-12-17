package REST

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/gorilla/mux"
	"math/big"
	"net/http"
)

func GetAccountEndpoint(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	param := params["id"]

	logger.Printf("Incoming acc request for id: %v", param)

	var address [64]byte

	pubKeyInt, _ := new(big.Int).SetString(param, 16)

	if len(param) == 128 {
		copy(address[:], pubKeyInt.Bytes())
	} else {
		logger.Fatal(fmt.Sprintf("provided invalid address %x\n", param))
	}

	acc, lastTenTx, err := client.GetAccount(address)
	if err != nil {
		SendJsonResponse(w, JsonResponse{http.StatusInternalServerError, err.Error(), nil})
	} else {
		var content []Content
		content = append(content, Content{"account", acc})

		for _, tx := range lastTenTx {
			if tx != nil {
				content = append(content, Content{"inbound", tx})
			}
		}

		SendJsonResponse(w, JsonResponse{http.StatusOK, "", content})
	}
}
