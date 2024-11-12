package internal

func PanicError(err error) {
	if err != nil {
		panic(err)
	}
}
