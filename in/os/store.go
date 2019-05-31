package os

import "os"
import "path/filepath"
import "errors"

import qlovastore "github.com/qlova/store"

type Store struct {
	err error
	
	Path string
}

func (store Store) Error() error {
	return store.err
}

func (store Store) Data(name string) qlovastore.Data {
	return qlovastore.Data{
		Raw: &Data{
			Path: filepath.Join(store.Path, name),
		},
	}
}

func (store Store) Goto(path string) qlovastore.Store {
	return Store{
		Path: store.Path+"/"+path,
	}
}

func (store Store) Create() error {
	return os.MkdirAll(store.Path, 0700)
}

func (store Store) String() string {
	return store.Path
}

func (store Store) List() qlovastore.Cursor {
	
	dir, err := os.Open(store.Path)
	if err != nil {
		store.err = err
		return nil
	}
	
	return &Cursor{
		directory: dir,
		Path: store.Path,
	}
}

func (store Store) Exists() bool {
	_, err := os.Stat(store.Path)
	return !os.IsNotExist(err)
}

func (store Store) Delete() error {
	return errors.New("Unimplemented")
}

//Open the given local directory as a store.
func Open(directory string) Store {
	
	err := os.MkdirAll(directory, 0700)
	
	return Store{
		Path: directory,
		err: err,
	}
}
