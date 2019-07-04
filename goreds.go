package goreds // import "blainsmith.com/go/goreds"

import (
	"blainsmith.com/go/goreds/words"
	"github.com/garyburd/redigo/redis"
)

const (
	AND Operator = "zinterstore"
	OR  Operator = "zunionstore"
)

type Operator string

//  Client is a goreds client that uses a Redis client and Namespace to perform searches
type Client struct {
	Redis     redis.Conn
	Namespace string
}

// NewClient will create a new search client with the given redigo connection and namespace (defaults to "goreds").
// You may create multiple clients with different namespaces for separate search indexes.
func NewClient(redis redis.Conn, namespace string) *Client {
	if namespace == "" {
		namespace = "goreds"
	}

	return &Client{redis, namespace}
}

// Index will store the `id` within the database and use the `text` as the searchable text
func (client *Client) Index(text, id string) error {
	w := words.StemsArray(words.StripStopWords(words.Split(text)))
	c := words.Count(w)
	mm := words.MetaphoneMap(w)

	client.Redis.Send("MULTI")
	for i := range mm {
		client.Redis.Send("ZADD", client.Namespace+":word:"+mm[i], c[i], id)
		client.Redis.Send("ZADD", client.Namespace+":object:"+id, c[i], mm[i])
	}

	_, err := client.Redis.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

// Remove will delete the `id` from the database so it is not longer searchable
func (client *Client) Remove(id string) error {
	reply, err := redis.Strings(client.Redis.Do("ZREVRANGEBYSCORE", client.Namespace+":object:"+id, "+inf", 0))
	if err != nil {
		return err
	}

	client.Redis.Send("MULTI")
	for _, m := range reply {
		client.Redis.Send("DEL", client.Namespace+":object:"+id)
		client.Redis.Send("ZREM", client.Namespace+":word:"+m, id)
	}

	_, err = client.Redis.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

// Query performs a search against the database and returns a slice of ids that match
func (client *Client) Query(text string, operator Operator) ([]string, error) {
	if operator == "" {
		operator = AND
	}

	w := words.StemsArray(words.StripStopWords(words.Split(text)))
	mk := words.MetaphoneKeys(client.Namespace, w)

	lmk := len(mk)
	if lmk <= 0 {
		return nil, nil
	}

	tk := client.Namespace + "temp"
	search := []interface{}{tk, lmk}
	for _, k := range mk {
		search = append(search, k)
	}

	client.Redis.Send("MULTI")
	client.Redis.Send(string(operator), search...)
	client.Redis.Send("ZREVRANGE", tk, 0, -1)
	client.Redis.Send("ZREMRANGEBYRANK", tk, 0, -1)

	values, err := redis.Values(client.Redis.Do("EXEC"))
	if err != nil {
		return nil, err
	}

	var ids []string
	count := 0
	_, err = redis.Scan(values, &count, &ids)
	if err != nil {
		return nil, err
	}

	return ids, nil
}
