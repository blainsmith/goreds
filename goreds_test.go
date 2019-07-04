package goreds_test

import (
	"fmt"
	"testing"

	"github.com/garyburd/redigo/redis"

	"blainsmith.com/go/goreds"
)

func TestNewClient(t *testing.T) {
	conn, err := redis.DialURL("redis://localhost:6379")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	t.Run("with namespace", func(t *testing.T) {
		ns := goreds.NewClient(conn, "WITH_NAMESPACE")
		if ns.Namespace != "WITH_NAMESPACE" {
			t.Error("wrong namespace is set")
		}
	})
	t.Run("without namespace", func(t *testing.T) {
		ns := goreds.NewClient(conn, "")
		if ns.Namespace != "goreds" {
			t.Error("wrong default namespace is set")
		}
	})
}

func ExampleNewClient() {
	redis, _ := redis.DialURL("redis://localhost:6379")
	search := goreds.NewClient(redis, "namespace")

	// index some text and assign to their id
	search.Index("example index text", "1")
	search.Index("example text being indexed for the sake of this example", "2")
	search.Index("example testing with a lot more text", "3")
	search.Index("index some more sample text data", "4")

	// remove an item from the search index
	search.Remove("2")

	// query the search index that should match ids 1 and 4
	ids, _ := search.Query("index text", goreds.AND)
	fmt.Println(ids)
	// Output: [4 1]
}

func TestIndex(t *testing.T) {
	conn, err := redis.DialURL("redis://localhost:6379")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	search := goreds.NewClient(conn, "testing")

	testtext := "Some text I would like to index for searching in the future."

	err = search.Index(testtext, "12345")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemove(t *testing.T) {
	conn, err := redis.DialURL("redis://localhost:6379")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	search := goreds.NewClient(conn, "testing")

	err = search.Remove("12345")
	if err != nil {
		t.Fatal(err)
	}
}

func TestQuery(t *testing.T) {
	conn, err := redis.DialURL("redis://localhost:6379")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	search := goreds.NewClient(conn, "testing")

	search.Index("baller search text", "111")
	search.Index("some more baller text to find", "222")
	search.Index("this text not be found", "333")

	t.Run("AND", func(t *testing.T) {
		ids, err := search.Query("baller text", goreds.AND)
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) != 2 {
			t.Error("expected 2 ids")
		}
		if ids[0] != "222" {
			t.Error("222 not in result set")
		}
		if ids[1] != "111" {
			t.Error("111 not in result set")
		}
	})

	t.Run("OR", func(t *testing.T) {
		ids, err := search.Query("baller text", goreds.OR)
		if err != nil {
			t.Fatal(err)
		}
		if len(ids) != 3 {
			t.Error("expected 3 ids")
		}
		if ids[0] != "222" {
			t.Error("222 not in result set")
		}
		if ids[1] != "111" {
			t.Error("111 not in result set")
		}
		if ids[2] != "333" {
			t.Error("333 not in result set")
		}
	})
}
