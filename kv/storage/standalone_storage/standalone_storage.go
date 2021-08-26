package standalone_storage

import (
	"github.com/Connor1996/badger"
	"github.com/pingcap-incubator/tinykv/kv/config"
	"github.com/pingcap-incubator/tinykv/kv/storage"
	"github.com/pingcap-incubator/tinykv/kv/util/engine_util"
	"github.com/pingcap-incubator/tinykv/proto/pkg/kvrpcpb"
	"path"
)

// StandAloneStorage is an implementation of `Storage` for a single-node TinyKV instance. It does not
// communicate with other nodes and all data is stored locally.
type StandAloneStorage struct {
	engine *engine_util.Engines
}

func NewStandAloneStorage(conf *config.Config) *StandAloneStorage {
	kvPath := path.Join(conf.DBPath, "kv")
	raftPath := path.Join(conf.DBPath, "raft")
	kvEngine := engine_util.CreateDB(kvPath, false)
	raftEngine := engine_util.CreateDB(raftPath, true)

	engine := engine_util.NewEngines(kvEngine, raftEngine, kvPath, raftPath)
	return &StandAloneStorage{engine: engine}
}

func (s *StandAloneStorage) Start() error {
	// Your Code Here (1).
	return nil
}

func (s *StandAloneStorage) Stop() error {
	// Your Code Here (1).
	return s.engine.Close()
}

func (s *StandAloneStorage) Reader(ctx *kvrpcpb.Context) (storage.StorageReader, error) {
	// Your Code Here (1).
	txn := s.engine.Kv.NewTransaction(false)
	return &StandAloneReader{txn: txn}, nil
}

func (s *StandAloneStorage) Write(ctx *kvrpcpb.Context, batch []storage.Modify) error {
	// Your Code Here (1).
	// all or nothing
	return s.engine.Kv.Update(func(txn *badger.Txn) error {
		for _, op := range batch {
			switch op.Data.(type) {
			case storage.Put:
				err := txn.Set(engine_util.KeyWithCF(op.Cf(), op.Key()), op.Value())
				if err != nil {
					return err
				}
			case storage.Delete:
				err := txn.Delete(engine_util.KeyWithCF(op.Cf(), op.Key()))
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

type StandAloneReader struct {
	txn *badger.Txn
}

func (s *StandAloneReader) GetCF(cf string, key []byte) ([]byte, error) {
	result, err := engine_util.GetCFFromTxn(s.txn, cf, key)
	if err == badger.ErrKeyNotFound {
		return nil, nil
	}
	return result, nil
}

func (s *StandAloneReader) IterCF(cf string) engine_util.DBIterator {
	return engine_util.NewCFIterator(cf, s.txn)
}

func (s *StandAloneReader) Close() {
	s.txn.Discard()
}
