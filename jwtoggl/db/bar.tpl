func (db *FoosDB) onFooAction(action string, item foo.Foo) error {
	log.Printf("Foo %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}
