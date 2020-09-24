package wallet

type API struct {
	Address string
}

func New(address string) *API {
	return &API{
		Address: address,
	}
}

func GROTH2Beam(groth uint64) float64 {
	return float64(groth) / 100000000
}