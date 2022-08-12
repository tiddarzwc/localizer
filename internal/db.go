package internal

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/dgraph-io/badger/v3"
)

var (
	ErrFileExist = errors.New("filename existed on server")
	ErrNotFound  = errors.New("meta(s) not found")
)

const (
	FilenamePrefix          = "f_"
	FileDownloadTimesPrefix = "fd_"
	ModNamePrefix           = "m_"
	UsernamePrefix          = "u_"
)

type Meta struct {
	ModName     string `json:"mod_name,omitempty"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description,omitempty"`
	Filename    string `json:"filename,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Culture     string `json:"culture,omitempty"`
	Version     string `json:"version,omitempty"`
}

type DB interface {
	// todo user namespace

	// file meta namespace
	PutMeta(meta Meta, closure func() error) error
	ListMetas(modName string) ([]Meta, error)
	GetMeta(filename string) (*Meta, error)
}

type database struct {
	db *badger.DB
}

func putOnFilenameIndex(txn *badger.Txn, meta Meta) error {
	filename := meta.Filename
	key := []byte(FilenamePrefix + filename)
	_, err := txn.Get(key)
	if err == nil {
		return ErrFileExist
	}
	buff, err := json.Marshal(&meta)
	if err != nil {
		return err
	}
	return txn.Set(key, buff)
}

func putOnModNameIndex(txn *badger.Txn, meta Meta) error {
	modName := meta.ModName
	metas := make([]Meta, 0, 8)
	key := []byte(ModNamePrefix + modName)
	item, err := txn.Get(key)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return err
		}
	} else {
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &metas)
		})
		if err != nil {
			return err
		}
	}
	metas = append(metas, meta)
	buff, err := json.Marshal(metas)
	if err != nil {
		return err
	}
	return txn.Set(key, buff)
}

func (d *database) PutMeta(meta Meta, closure func() error) error {
	db := d.db
	return db.Update(func(txn *badger.Txn) error {
		err := putOnFilenameIndex(txn, meta)
		if err != nil {
			return err
		}
		err = putOnModNameIndex(txn, meta)
		if err != nil {
			return err
		}
		return closure()
	})
}
func (d *database) ListMetas(modName string) ([]Meta, error) {
	db := d.db
	metas := make([]Meta, 0, 8)
	err := db.View(func(txn *badger.Txn) error {
		if modName == "" {
			iter := txn.NewIterator(badger.IteratorOptions{
				PrefetchSize:   10,
				PrefetchValues: true,
				Prefix:         []byte(ModNamePrefix),
			})
			defer iter.Close()
			for iter.Rewind(); iter.Valid(); iter.Next() {
				curMetas := make([]Meta, 0, 8)
				item := iter.Item()
				err := item.Value(func(val []byte) error {
					return json.Unmarshal(val, &curMetas)
				})
				if err != nil {
					return err
				}
				metas = append(metas, curMetas...)
			}
			return nil
		}
		key := []byte(ModNamePrefix + modName)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &metas)
		})
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			err = ErrNotFound
		}
		return nil, err
	}
	return metas, nil
}

func (d *database) GetMeta(filename string) (*Meta, error) {
	db := d.db
	meta := &Meta{}
	err := db.View(func(txn *badger.Txn) error {
		key := []byte(FilenamePrefix + filename)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, meta)
		})
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			err = ErrNotFound
		}
		return nil, err
	}
	return meta, nil
}

func NewDB(dbPath string) DB {
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		log.Fatal("启动数据库失败:", err)
	}
	return &database{
		db: db,
	}
}
