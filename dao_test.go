// Copyright 2016 polaris. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author：polaris	polaris@studygolang.com

package dbutil_test

import (
	"testing"

	"github.com/polaris1119/dbutil"
)

// func TestFindOne(t *testing.T) {
// 	goods := &dbutil.Goods{}
// 	err := dbutil.NewDao().OrderBy("tid DESC").FindOne(goods)
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	} else {

// 		// t.Fatal("tid==", goods.Tid, "top==", goods.Top)
// 		t.Fatal(*goods)

// 	}
// }

// func TestFindAll(t *testing.T) {
// 	goodsList := make([]*dbutil.Goods, 10)
// 	// err := dbutil.NewDao().OrderBy("tid DESC").FindOne(goods)
// 	err := dbutil.NewDao().Table("topics").Where("uid=?", 1).Limit(10).FindAll(goodsList)
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	} else {

// 		if len(goodsList) == 0 {
// 			t.Fatal("error")
// 		}

// 		t.Log(len(goodsList))
// 		t.Fatal(*goodsList[0])

// 	}
// }

// func TestFindBySql(t *testing.T) {
// 	strSql := "SELECT * FROM topics t LEFT JOIN topics_ex te ON t.tid=te.tid WHERE uid=1"
// 	rows, err := dbutil.NewDao().FindBySql(strSql)
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	} else {

// 		// t.Fatal("tid==", goods.Tid, "top==", goods.Top)
// 		t.Fatal(rows)

// 	}
// }

// func TestUpdate(t *testing.T) {
// 	affectedNum, err := dbutil.NewDao().Table("test").Set("title=?", "测试").Where("id=?", 1).Update()
// 	if err != nil {
// 		t.Fatal(err.Error())
// 	} else {

// 		// t.Fatal("tid==", goods.Tid, "top==", goods.Top)
// 		t.Fatal(affectedNum)

// 	}
// }

func TestPersist(t *testing.T) {
	goods := &dbutil.Goods{}
	dao := dbutil.NewDao()
	dao.Where("tid=?", 2).FindOne(goods)
	goods.Title = "测试"
	affectedNum, err := dao.Persist(goods, "title")
	if err != nil {
		t.Fatal(err.Error())
	} else {

		// t.Fatal("tid==", goods.Tid, "top==", goods.Top)
		t.Fatal(affectedNum)

	}
}

type Test struct {
	Id    uint   `json:"id"`
	Title string `json:"title"`
}

func (t *Test) Table() string {
	return "test"
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
