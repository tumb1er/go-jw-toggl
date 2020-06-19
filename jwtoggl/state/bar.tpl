
// bar.tpl предоставляет общие методы для справочников (которых нет у TimeEntriesDB)

func (db *FoosDB) onFooAction(action string, item foo.Foo) error {
	log.Printf("Foo %s %s, sync user", item.Name, action)
	return db.SyncUser(false)
}
