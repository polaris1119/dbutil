// Copyright 2016 polaris. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	polaris@studygolang.com

package dbutil_test

import (
	"time"

	"github.com/polaris1119/dbutil"
)

func init() {
	dbutil.InitDB("root:@tcp(localhost:3306)/studygolang?charset=utf8")
}

// func TestFindOne(t *testing.T) {
// 	topics := &dbutil.Topics{}
// 	err := dbutil.NewDao().OrderBy("tid DESC").FindOne(topics)
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	} else {

// 		// t.Fatal("tid==", topics.Tid, "top==", topics.Top)
// 		t.Fatal(*topics)

// 	}
// }

// func TestFindAll(t *testing.T) {
// 	topicsList := make([]*dbutil.Topics, 10)
// 	// err := dbutil.NewDao().OrderBy("tid DESC").FindOne(topics)
// 	err := dbutil.NewDao().Table("topics").Where("uid=?", 1).Limit(10).FindAll(topicsList)
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	} else {

// 		if len(topicsList) == 0 {
// 			t.Fatal("error")
// 		}

// 		t.Log(len(topicsList))
// 		t.Fatal(*topicsList[0])

// 	}
// }

// func TestFindBySql(t *testing.T) {
// 	strSql := "SELECT * FROM topics t LEFT JOIN topics_ex te ON t.tid=te.tid WHERE uid=1"
// 	rows, err := dbutil.NewDao().FindBySql(strSql)
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	} else {

// 		// t.Fatal("tid==", topics.Tid, "top==", topics.Top)
// 		t.Fatal(rows)

// 	}
// }

// func TestUpdate(t *testing.T) {
// 	affectedNum, err := dbutil.NewDao().Table("test").Set("title=?", "测试").Where("id=?", 1).Update()
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	} else {

// 		// t.Fatal("tid==", topics.Tid, "top==", topics.Top)
// 		t.Fatal(affectedNum)

// 	}
// }

// func TestPersist(t *testing.T) {
// 	topics := &dbutil.Topics{}
// 	dao := dbutil.NewDao()
// 	dao.Where("tid=?", 2).FindOne(topics)
// 	topics.Title = "测试"
// 	affectedNum, err := dao.Persist(topics, "title")
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	} else {

// 		// t.Fatal("tid==", topics.Tid, "top==", topics.Top)
// 		t.Fatal(affectedNum)

// 	}
// }

// func TestTransaction(t *testing.T) {
// 	// topics := &dbutil.Topics{}

// 	test := &Test{}

// 	dao := dbutil.NewDao()
// 	dao.Begin()

// 	err := dao.Where("id=?", 1).FindOne(test)
// 	if err != nil {
// 		dao.Rollback()
// 		t.Fatal("rollback.", err)
// 	}

// 	test.Title = "22233"
// 	affectedNum, err := dao.Persist(test, "title")
// 	if err != nil {
// 		dao.Rollback()
// 		t.Fatal("rollback..", err)
// 	}

// 	affectedNum, err = dao.Table("topics").Where("tid=?", 2).Set("title=?", "事务测试2").Update()
// 	if err != nil {
// 		dao.Rollback()
// 		t.Fatal(err.Error())
// 	} else {
// 		dao.Commit()
// 		// t.Fatal("tid==", topics.Tid, "top==", topics.Top)
// 		t.Fatal(affectedNum)

// 	}
// }

type Test struct {
	Id    uint   `json:"id"`
	Title string `json:"title"`
}

func (t *Test) Table() string {
	return "test"
}

type Topics struct {
	Tid           uint      `json:"tid" pk:"1"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Nid           int       `json:"nid"`
	Uid           uint32    `json:"uid"`
	Lastreplyuid  uint32    `json:"lastreplyuid"`
	Lastreplytime time.Time `json:"lastreplytime"`
	Top           bool      `db:"top" json:"istop"`
}

func (t *Topics) Table() string {
	return "topics"
}

// func TestInsert(t *testing.T) {
// 	test := &Test{
// 		Title: "测试",
// 	}

// 	result, err := dbutil.NewDao().Insert(test)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Fatal(result)
// }
