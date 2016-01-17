// Copyright 2016 polaris. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	polaris@studygolang.com

package dbutil_test

import (
	"testing"
	"time"

	"github.com/polaris1119/dbutil"
)

func init() {
	dbutil.InitDB("root:@tcp(localhost:3306)/studygolang?charset=utf8")
}

type Test struct {
	Id    uint   `json:"id"`
	Title string `json:"title"`
}

func (t *Test) Table() string {
	return "test"
}

// 社区主题信息
type Topic struct {
	Tid           int       `json:"tid" pk:"1"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Nid           int       `json:"nid"`
	Uid           int       `json:"uid"`
	Flag          uint8     `json:"flag"`
	Lastreplyuid  int       `json:"lastreplyuid"`
	Lastreplytime time.Time `json:"lastreplytime"`
	EditorUid     int       `json:"editor_uid"`
	Top           bool      `json:"istop" db:"top"`
	Ctime         time.Time `json:"ctime"`
	Mtime         time.Time `json:"mtime"`

	// 为了方便，加上Node（节点名称，数据表没有）
	// Node string
}

func (*Topic) Table() string {
	return "topics"
}

// 社区主题扩展（计数）信息
type TopicEx struct {
	Tid   int       `json:"tid" pk:"1"`
	View  int       `json:"view"`
	Reply int       `json:"reply"`
	Like  int       `json:"like"`
	Mtime time.Time `json:"mtime"`
}

func (*TopicEx) Table() string {
	return "topics_ex"
}

func TestFindBySql(t *testing.T) {
	dao := dbutil.NewDao()
	strSql := "SELECT * FROM topics t LEFT JOIN topics_ex ex ON t.tid=ex.tid"
	rows, err := dao.FindBySql(strSql)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	var (
		topics   = make([]*Topic, 10)
		topicExs = make([]*TopicEx, 10)
	)

	err = dao.ScanRows(rows, topics, topicExs)
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal(topics[0].Uid, "==", topicExs[0].Mtime)
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
