package wallet

type API struct {
	Address string
}

const  (
	GrothInBEAM uint64 = 100000000
)

func New(address string) *API {
	return &API{
		Address: address,
	}
}

func GROTH2Beam(groth uint64) float64 {
	return float64(groth) / float64(GrothInBEAM)
}
