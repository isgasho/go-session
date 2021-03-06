// Copyright (c) 2020 HigKer
// Open Source: MIT License
// Author: SDing <deen.job@qq.com>
// Date: 2020/8/23 - 9:10 PM - UTC/GMT+08:00



// 目前本库已经实现内存存储和Redis做分布式存储
// 内存版本使用的是计算机内存 所有可能在运行的时候内存可能会大一点
// session从第一次请求过来就创建 生命周期根据你自己设置单位秒
// 每次都是按照生命周期来计算一个session的周期
// 例如一次30分钟那这个session周期就是30分钟

package session


import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// session unite standard
type Session interface {
	Get(key string) ([]byte, error)
	Set(key string, data interface{}) error
	Del(key string) error
	Clean(w http.ResponseWriter)
}

// Session for  Memory item
type MemorySession struct {
	ID      string                 // Unique id
	Safe    sync.Mutex             // Mutex lock
	Expires time.Time              // Expires time
	Data    map[string]interface{} // Save data
}

// Session for redis item
type RedisSession struct {
	ID      string     // Unique id
	Safe    sync.Mutex // Mutex lock
	Expires time.Time  // Expires time
}

//实例化 memory session
func newRSession(id string, maxAge int) *RedisSession {
	return &RedisSession{
		ID:      id,
		Expires: time.Now().Add(time.Duration(maxAge) * time.Second),
	}
}

//实例化 redis session
func newMSessionItem(id string, maxAge int) *MemorySession {
	return &MemorySession{
		Data:    make(map[string]interface{}, maxSize),
		ID:      id,
		Expires: time.Now().Add(time.Duration(maxAge) * time.Second),
	}
}

// Get get session data by key
func (rs *RedisSession) Get(key string) ([]byte, error) {
	if key == "" || len(key) <= 0 {
		return nil, ErrorKeyNotExist
	}
	//var result Value
	//result.Key = key
	// 把id和到期时间传过去方便后面使用
	cv := map[string]interface{}{contextValueID: rs.ID, contextValueKey: key}
	value := context.WithValue(context.TODO(), contextValue, cv)
	b, err := _Store.Reader(value)
	if err != nil {
		return nil, err
	}
	//result.Value = b
	return b, nil
}

// Set set session data by key
func (rs *RedisSession) Set(key string, data interface{}) error {
	if key == "" || len(key) <= 0 {
		return ErrorKeyFormat
	}
	cv := map[string]interface{}{contextValueID: rs.ID, contextValueKey: key, contextValueData: data}
	value := context.WithValue(context.TODO(), contextValue, cv)
	return _Store.Writer(value)
}

// Del delete session data by key
func (rs *RedisSession) Del(key string) error {
	if key == "" || len(key) <= 0 {
		return ErrorKeyFormat
	}
	cv := map[string]interface{}{contextValueID: rs.ID, contextValueKey: key}
	value := context.WithValue(context.TODO(), contextValue, cv)
	_Store.Remove(value)
	return nil
}

// Clean clean session data
func (rs *RedisSession) Clean(w http.ResponseWriter) {
	cv := map[string]interface{}{contextValueID: rs.ID}
	value := context.WithValue(context.TODO(), contextValue, cv)
	_Store.Clean(value)
	cookie := &http.Cookie{
		Name:     _Cfg.CookieName,
		Value:    "",
		Path:     _Cfg.Path,
		Domain:   _Cfg.Domain,
		Secure:   _Cfg.Secure,
		MaxAge:   -1,
		Expires:  time.Now().AddDate(-1, 0, 0),
		HttpOnly: _Cfg.HttpOnly,
	}
	http.SetCookie(w, cookie)
}

// Get get session data by key
func (ms *MemorySession) Get(key string) ([]byte, error) {
	if key == "" || len(key) <= 0 {
		return nil, ErrorKeyNotExist
	}
	//var result Value
	//result.Key = key
	// 把id和到期时间传过去方便后面使用
	cv := map[string]interface{}{contextValueID: ms.ID, contextValueKey: key}
	value := context.WithValue(context.TODO(), contextValue, cv)
	b, err := _Store.Reader(value)
	if err != nil {
		return nil, err
	}
	//result.Value = b
	return b, nil
}

// Set set session data by key
func (ms *MemorySession) Set(key string, data interface{}) error {
	if key == "" || len(key) <= 0 {
		return ErrorKeyFormat
	}
	cv := map[string]interface{}{contextValueID: ms.ID, contextValueKey: key, contextValueData: data}
	value := context.WithValue(context.TODO(), contextValue, cv)
	return _Store.Writer(value)
}

// Del delete session data by key
func (ms *MemorySession) Del(key string) error {
	if key == "" || len(key) <= 0 {
		return ErrorKeyFormat
	}
	cv := map[string]interface{}{contextValueID: ms.ID, contextValueKey: key}
	value := context.WithValue(context.TODO(), contextValue, cv)
	_Store.Remove(value)
	return nil
}

// Clean clean session data
func (ms *MemorySession) Clean(w http.ResponseWriter) {
	cv := map[string]interface{}{contextValueID: ms.ID}
	value := context.WithValue(context.TODO(), contextValue, cv)
	_Store.Clean(value)
	cookie := &http.Cookie{
		Name:     _Cfg.CookieName,
		Value:    "",
		Path:     _Cfg.Path,
		Domain:   _Cfg.Domain,
		Secure:   _Cfg.Secure,
		MaxAge:   -1,
		Expires:  time.Now().AddDate(-1, 0, 0),
		HttpOnly: _Cfg.HttpOnly,
	}
	http.SetCookie(w, cookie)
}

//// 检测sessionID是否有效
//func IdNotExist(id string) bool {
//	return _Store.(*RedisStore).client.HGetAll(_Cfg.RedisKeyPrefix+id).Err() == nil
//
//}

func newCookie(w http.ResponseWriter, cookie *http.Cookie) (session Session) {
	// 创建一个cookie
	sid := string(Random(32, RuleKindAll))
	cookie = &http.Cookie{
		Name: _Cfg.CookieName,
		//这里是并发不安全的，但是这个方法已上锁
		Value:    url.QueryEscape(sid), //转义特殊符号@#￥%+*-等
		Path:     _Cfg.Path,
		Domain:   _Cfg.Domain,
		HttpOnly: _Cfg.HttpOnly,
		Secure:   _Cfg.Secure,
		MaxAge:   int(_Cfg.MaxAge),
		Expires:  time.Now().Add(time.Duration(_Cfg.MaxAge)),
	}
	http.SetCookie(w, cookie) //设置到响应中
	if _Cfg._st == Memory {
		item := newMSessionItem(sid, int(_Cfg.MaxAge))
		_Store.(*MemoryStore).values[sid] = item
		session = item
		return
	}
	session = newRSession(sid, int(_Cfg.MaxAge))
	return
}

// Builder build  session store
func Builder(store StoreType, conf *Config) error {
	if conf.MaxAge < DefaultMaxAge {
		return errors.New("session maxAge no less than 30min")
	}

	_Cfg = conf
	switch store {
	default:
		return errors.New("build session error, not implement type store")
	case Memory:
		_Store = newMemoryStore()
		_Cfg._st = Memory
		return nil
	case Redis:
		redisStore, err := newRedisStore()
		if err != nil {
			return err
		}
		_Store = redisStore
		_Cfg._st = Redis
		return nil
	}
}

// Ctx return request session object
func Ctx(writer http.ResponseWriter, request *http.Request) (Session, error) {

	// 检测是否有这个session数据
	// 1.全局session垃圾回收器
	// 2.请求一过来就进行检测
	// 3.如果请求的id值在内存存在并且垃圾回收也存在时间也是有效的就说明这个session是有效的

	cookie, err := request.Cookie(_Cfg.CookieName)
	if _Cfg._st == Memory {
		if err != nil || len(cookie.Value) <= 0 {
			item := newCookie(writer, cookie)
			return item, nil
		}
		// 防止浏览器关闭重新打开抛异常
		sid, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			return nil, err
		}
		session := _Store.(*MemoryStore).values[sid]
		if session == nil {
			item := newCookie(writer, cookie)
			_Store.(*MemoryStore).values[sid] = item.(*MemorySession)
			return item, nil
		}
		return _Store.(*MemoryStore).values[cookie.Value], nil
	}
	// 如果没有session数据就重新创建一个
	if err != nil || cookie == nil || len(cookie.Value) <= 0 {
		// 重新生成一个cookie 和唯一 sessionID
		rs := newCookie(writer, cookie)
		return rs, nil
	}
	sid, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return nil, err
	}
	return newRSession(sid, int(_Cfg.MaxAge)), nil
}
