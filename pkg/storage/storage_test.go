package storage

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

const connstr = "postgresql://[userspec@][hostspec][/dbname][?paramspec]"

func TestNew(t *testing.T) {
	_, err := New(connstr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDB_News(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	posts := []Post{
		{
			Title: "Test Post",
			Link:  strconv.Itoa(r.Intn(1_000_000_000)),
		},
	}
	db, err := New(connstr)
	if err != nil {
		t.Fatal(err)
	}
	err = db.StoreNews(posts)
	if err != nil {
		t.Fatal(err)
	}
	news, err := db.News(2)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", news)
}
