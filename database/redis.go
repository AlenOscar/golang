/**
 * @Author: 一秋
 * @Created Date: 2019-05-17 22:24
 * @Description: TODO
 */

package database

import (
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"time"
)

const (
	RedisURL            = "redis://127.0.0.1:6379"
	redisMaxIdle        = 3      // 最大空闲连接数
	redisIdleTimeoutSec = 240    // 最大空闲连接时间
	RedisDbIndex        = 0      // 数据库索引[0 - 15]
	RedisPassword       = "1234" // redis密码
)

type redisPool struct {
	pool *redis.Pool // redis connection pool
}

var RedisCli *redisPool // redis client

func init() {
	if RedisCli == nil {
		RedisCli = &redisPool{
			pool: createRedisPool(RedisURL),
		}
	}
}

// GetRedisPool返回redis连接池
func createRedisPool(redisURL string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     redisMaxIdle,
		IdleTimeout: redisIdleTimeoutSec * time.Second,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.DialURL(redisURL, redis.DialDatabase(RedisDbIndex))
			if err != nil {
				return nil, fmt.Errorf("redis connection error: %s", err)
			}
			// 验证redis密码
			if _, authErr := conn.Do("AUTH", RedisPassword); authErr != nil {
				return nil, fmt.Errorf("redis auth password error: %s", authErr)
			}
			return conn, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				return fmt.Errorf("ping redis error: %s", err)
			}
			return nil
		},
	}
}

// ************************************** Redis keys 命令 **************************************
// 键值操作
func (cli *redisPool) SetValue() {
	conn := cli.pool.Get()
	defer conn.Close()

}

func (cli *redisPool) SetInt64(k string, v int64) {
	c := cli.pool.Get()
	defer c.Close()
	_, err := c.Do("SET", k, v)
	if err != nil {
		fmt.Println("set error", err.Error())
	}
}

func (cli *redisPool) Set(k, v string) {
	c := cli.pool.Get()
	defer c.Close()
	_, err := c.Do("SET", k, v)
	if err != nil {
		fmt.Println("set error", err.Error())
	}
}

func (cli *redisPool) GetStringValue(k string) string {
	c := cli.pool.Get()
	defer c.Close()
	username, err := redis.String(c.Do("GET", k))
	if err != nil {
		fmt.Println("Get Error: ", err.Error())
		return ""
	}
	return username
}

func (cli *redisPool) SetKeyExpire(k string, ex int) {
	c := cli.pool.Get()
	defer c.Close()
	_, err := c.Do("EXPIRE", k, ex)
	if err != nil {
		fmt.Println("set error", err.Error())
	}
}

func (cli *redisPool) CheckKey(k string) bool {
	c := cli.pool.Get()
	defer c.Close()
	exist, err := redis.Bool(c.Do("EXISTS", k))
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		return exist
	}
}

func (cli *redisPool) DelKey(k string) error {
	c := cli.pool.Get()
	defer c.Close()
	_, err := c.Do("DEL", k)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// SET if Not exists
func (cli *redisPool) SetOnce(k string, data interface{}) error {
	c := cli.pool.Get()
	defer c.Close()
	value, _ := json.Marshal(data)
	reply, _ := c.Do("SETNX", k, value)
	if reply != int64(1) {
		return errors.New("key value already existed")
	}
	return nil
}

func (cli *redisPool) GetJsonByte(key string) ([]byte, error) {
	c := cli.pool.Get()
	defer c.Close()
	jsonGet, err := redis.Bytes(c.Do("GET", key))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return jsonGet, nil
}

func (cli *redisPool) RenameKey(oldKey, newKey string) bool {
	c := cli.pool.Get()
	defer c.Close()
	if cli.CheckKey(oldKey) == true {
		ok, err := redis.String(c.Do("RENAME", oldKey, newKey))
		if err != nil {
			fmt.Println(err)
			return false
		} else {
			fmt.Println(ok)
			return true
		}
	}
	return false
}

func (cli *redisPool) AddInt64Value(key string, increment int64) error {
	c := cli.pool.Get()
	defer c.Close()
	_, err := c.Do("INCRBY", key, increment)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// ********************************************** Redis String Operation ********************************************** //
//

// *********************************************** Redis Hash Operation *********************************************** //
// Set hash value to redis.
func (cli *redisPool) SetHash(key, field, value string) error {
	conn := cli.pool.Get()
	defer conn.Close()

	_, err := redis.Bool(conn.Do("HSET", key, field, value))
	if err != nil {
		return err
	}

	return nil
}

// Get hash value from redis.
func (cli *redisPool) GetHash(key, field string) ([]byte, error) {
	conn := cli.pool.Get()
	defer conn.Close()

	result, err := redis.Bytes(conn.Do("HGET", key, field))
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Get multiHash value from redis.
func (cli *redisPool) GetHashMulti(key string, fields ...string) (map[string][]byte, error) {
	conn := cli.pool.Get()
	defer conn.Close()

	args := make([]interface{}, len(fields)+1)
	args[0] = key
	for i := range fields {
		args[i+1] = fields[i]
	}

	values, err := redis.Values(conn.Do("HMGET", args...))
	if err != nil {
		return nil, err
	}

	results := make(map[string][]byte)
	for i, field := range fields {
		if values[i] != nil {
			v, ok := values[i].([]byte)
			if !ok {
				results[field] = nil
			} else {
				results[field] = v
			}
		} else {
			results[field] = nil
		}
	}

	return results, nil
}

// Get all Hash value from redis.
func (cli *redisPool) GetHashAll(key string) (map[string][]byte, error) {
	conn := cli.pool.Get()
	defer conn.Close()

	fieldValues, err := redis.Values(conn.Do("HGETALL", key))
	if err != nil {
		return nil, err
	}

	results := make(map[string][]byte)
	for i := 0; i + 1 < len(fieldValues); i = i+2 {
		field, fok := fieldValues[i].([]byte)
		value, vok := fieldValues[i+1].([]byte)
		if fok && vok {
			results[string(field)] = value
		}
	}

	return results, nil
}