package db

func Reset(name string) (err error) {
	_ = Destroy(name)
	err = Create(name)
	if err != nil {
		return err
	}
	return Migrate(name)
}
