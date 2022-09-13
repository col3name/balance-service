package handler

type responseWithoutData struct {
	Code    uint32 `json:"code"`
	Message string `json:"message"`
}

type response struct {
	Data interface{} `json:"data"`
}

type Error struct {
	Status   int
	Response responseWithoutData
}

func NewError(status, code int, err error) *Error {
	return &Error{
		Status: status,
		Response: responseWithoutData{
			Code:    uint32(code),
			Message: err.Error(),
		},
	}
}

type moneyTransferRequest struct {
	From        string `json:"from"`
	To          string `json:"to"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}

type moneyRequest struct {
	Account     string `json:"account"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
}
