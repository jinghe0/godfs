package db

import (
    "container/list"
    "sync"
    "time"
    "util/logger"
    "fmt"
    "errors"
)

// TODO bug: 大量数据测试发现，数据库连接存在泄露机会，导致程序大量等待连接。

type IDbConnPool interface {
    InitPool(connSize int)
    GetDB() *DAO
    ReturnDB(dao *DAO)
}

type DbConnPool struct {
    connSize int
    dbList *list.List
    fetchLock *sync.Mutex
}

func NewPool(poolSize int) *DbConnPool {
    pool := &DbConnPool{}
    pool.InitPool(poolSize)
    return pool
}

// init db connection pool
func (pool *DbConnPool) InitPool(poolSize int) {
    pool.connSize = poolSize
    pool.dbList = list.New()
    pool.fetchLock = new(sync.Mutex)
    for i := 0; i < poolSize; i++ {
        dao := &DAO{}
        dao.InitDB(i)
        pool.dbList.PushBack(dao)
    }
}

//fetch dao
func (pool *DbConnPool) GetDB() (*DAO, error) {
    pool.fetchLock.Lock()
    defer pool.fetchLock.Unlock()
    for {
        dao := pool.dbList.Remove(pool.dbList.Front())
        waits := 0
        if dao == nil {
            if waits > 30 {
                return nil, errors.New("cannot fetch db connection from pool: wait time out")
            }
            fmt.Print("\n\n等待数据库连接..........\n\n\n")
            logger.Debug("no connection available")
            time.Sleep(time.Millisecond * 100)
            waits++
        } else {
            logger.Trace("using db connection of index:", dao.(*DAO).index)
            return dao.(*DAO), nil
        }
    }
}

// return dao
func (pool *DbConnPool) ReturnDB(dao *DAO) {
    if dao != nil {
        logger.Trace("return db connection of index:", dao.index)
        pool.dbList.PushBack(dao)
    }
}

