package machineinfo

func ExploreMulti(handlers ...func() error) error {
	for _, f := range handlers {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}
