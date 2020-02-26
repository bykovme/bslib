package bslib

func (storage *StorageSingleton) checkReadiness() error {
	if storage.encObject == nil || !storage.encObject.IsReady() {
		return formError("Encrypter is not initialized", "checkReadiness")
	}

	if storage.dbObject == nil {
		return formError("Database is not initialized", "checkReadiness")
	}

	return nil
}

func (storage *StorageSingleton) DeleteItem(deleteItemParms JSONInputUpdateItem) (response JSONResponseCommon, err error) {
	err = storage.checkReadiness()
	if err != nil {
		return response, err
	}
	err = storage.dbObject.StartTX()
	if err != nil {
		return response, err
	}

	err = storage.dbObject.dbDeleteItem(deleteItemParms.ItemID)
	if err != nil {
		errEndTX := storage.dbObject.RollbackTX()
		if errEndTX != nil {
			return response, formError(BSERR00016DbDeleteFailed, err.Error(), errEndTX.Error())
		}
		return response, err
	}

	err = storage.dbObject.CommitTX()
	if err != nil {
		return response, err
	}

	response.Status = ConstSuccessResponse

	return response, nil
}

// AddNewItem - adds new item
func (storage *StorageSingleton) AddNewItem(addItemParms JSONInputUpdateItem) (response JSONResponseItemAdded, err error) {

	err = storage.checkReadiness()
	if err != nil {
		return response, err
	}

	if !CheckIfExistsFontAwesome(addItemParms.ItemIcon) {
		return response, formError(BSERR00006DbInsertFailed)
	}

	encryptedItemName, err := storage.encObject.Encrypt(addItemParms.ItemName)
	if err != nil {
		return response, err
	}

	encryptedIconName, err := storage.encObject.Encrypt(addItemParms.ItemIcon)
	if err != nil {
		return response, err
	}

	err = storage.dbObject.StartTX()
	if err != nil {
		return response, err
	}

	response.ItemID, err = storage.dbObject.dbInsertItem(encryptedItemName, encryptedIconName)
	if err != nil {
		errEndTX := storage.dbObject.RollbackTX()
		if errEndTX != nil {
			return response, formError(BSERR00006DbInsertFailed, err.Error(), errEndTX.Error())
		}
		return response, err
	}

	err = storage.dbObject.CommitTX()
	if err != nil {
		return response, err
	}

	response.Status = ConstSuccessResponse

	return response, nil
}

// ReadAllItems - read all not deleted items from the database and decrypt them
func (storage *StorageSingleton) ReadAllItems() (items []BSItem, err error) {
	err = storage.checkReadiness()
	if err != nil {
		return items, err
	}

	items, err = storage.dbObject.dbSelectAllItems()
	if err != nil {
		return items, err
	}

	for i := range items {
		items[i].Name, err = storage.encObject.Decrypt(items[i].Name)
		if err != nil {
			return items, err
		}
		items[i].Icon, err = storage.encObject.Decrypt(items[i].Icon)
		if err != nil {
			return items, err
		}
	}

	return items, nil
}

// ReadAllItems - read all not deleted items from the database and decrypt them
func (storage *StorageSingleton) ReadItemById(itemId int64) (item BSItem, err error) {
	err = storage.checkReadiness()
	if err != nil {
		return item, err
	}

	item, err = storage.dbObject.dbGetItemById(itemId)
	if err != nil {
		return item, err
	}

	item.Name, err = storage.encObject.Decrypt(item.Name)
	if err != nil {
		return item, err
	}
	item.Icon, err = storage.encObject.Decrypt(item.Icon)
	if err != nil {
		return item, err
	}

	item.Fields, err = storage.ReadFieldsByItemID(itemId)
	if err != nil {
		return item, err
	}

	return item, nil
}
