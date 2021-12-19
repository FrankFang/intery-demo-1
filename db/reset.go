package db

// type Container struct {
// 	mu sync.Mutex
// }

// var c = Container{}

func Reset() (err error) {
	// c.mu.Lock()
	// defer c.mu.Unlock()
	_ = Drop()
	err = Create()
	if err != nil {
		return err
	}
	return Migrate()
}
