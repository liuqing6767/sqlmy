# 项目名称
sqlmy 是一个 对官方的mysql二次封装的工具库, 它的优势在于:
- 完全兼容官方库接口,面向接口设计
- 提供 dao 命令行自动生成dao层的代码,开发者可以在完成表结构定义后一键生成dao 层代码
- 强类型访问数据库的方式.这个的具体实现依赖didi的类库,轮子已经那么多了,没必要再造了
- 引入mock工具,轻而易举的实现数据库的stub


## 面向接口的设计,完全兼容官方库

这块其实是为了能够实现单测是mock对象的注入,并且对官方库进行一些删改.包括:
- 数据库访问前：提供 SqlBuilder 接口，它完成 Sql 构造
- 数据库访问时：
    - 废弃所有没有 `Context` 后缀的官方实现，比如保留 `QueryContext`，但是废弃 `Query`
    - 扩充 `WithBuilder` 后缀，组合SqlBuilder能力，比如有 `QueryContextWitBuilder`
    - 扩充 `WithBuilderStruct` 后缀，进一步组合Orm能力，  比如有 `QueryContextWitBuilderStruct`
- 数据库访问后：提供了 ScanStruct 接口，完成 Orm 能力

### 重要对象说明
- DB：原生库 database.DB 的扩展实现（扩充内容见整体思路）
- Conn: 原始库 daabase.Conn 的扩展实现
- Tx：原始库database.Tx 的扩展实现
- Stmt：原始库database.Stmt 的扩展实现
- Rows：原始库database.Rows 的扩展实现
- Row：原始库database.Row 的扩展实现
- Hooks：日志可扩展钩子，默认实现有LogHook

具体的可以参考 [接口设计](inter.go)

## dao层代码一键生成工具

```
dao -db demo_db [-t demo_table] [-u root] [-p password] [-h 127.0.0.1] [-P 3306] [-json false] [-pkg dao]
```

它会读取数据库定义的表结构信息,完成代码自动生成,生成代码大概如下:

```
// AUTO GEN BY dao, Modify it as you want
package dao

import (
	"context"
	"database/sql"
	"time"

	"github.com/liuximu/sqlmy"
)

const cardTable = "card"

type Card struct {
	Id         int64  `ddb:"id" json:"id"`
	Name       string `ddb:"name" json:"name"`
	KnowID     int64  `ddb:"kid" json:"kid"`
}

type CardParam struct {
	Id         *int64  `ddb:"id" json:"id"`
	Ids        []int64 `ddb:"id,in" json:"id"`
	Name       *string `ddb:"name" json:"name"`
	KnowID     *int64  `ddb:"kid" json:"kid"`

	OrderBy *string `ddb:"_orderby"`
}

type CardList []*Card
type CardParamList []*CardParam

func (c *Card) Query(ctx context.Context, db sqlmy.QueryExecer, where *CardParam, columns []string) error {
	return db.QueryRowContextWithBuilderStruct(ctx, sqlmy.NewSelectBuilder(cardTable, sqlmy.Struct2Where(where), columns), c)
}

func (cp *CardParam) Create(ctx context.Context, db sqlmy.QueryExecer) (sql.Result, error) {
	return db.ExecContextWithBuilder(ctx, sqlmy.NewInsertBuilder(cardTable, sqlmy.Struct2AssignList(cp)))
}

func (cp *CardParam) Update(ctx context.Context, db sqlmy.QueryExecer, where *CardParam) (sql.Result, error) {
	return db.ExecContextWithBuilder(ctx, sqlmy.NewUpdateBuilder(cardTable, sqlmy.Struct2Where(where), sqlmy.Struct2Assign(cp)))
}

func (cp *CardParam) Delete(ctx context.Context, db sqlmy.QueryExecer) (sql.Result, error) {
	return db.ExecContextWithBuilder(ctx, sqlmy.NewDeleteBuilder(cardTable, sqlmy.Struct2Where(cp)))
}

func (cl *CardList) Query(ctx context.Context, db sqlmy.QueryExecer, where *CardParam, columns []string) error {
	return db.QueryContextWithBuilderStruct(ctx, sqlmy.NewSelectBuilder(cardTable, sqlmy.Struct2Where(where), columns), cl)
}

func (cpl CardParamList) Create(ctx context.Context, db sqlmy.QueryExecer) (sql.Result, error) {
	_cpl := make([]interface{}, len(cpl))
	for i, one := range cpl {
		_cpl[i] = one
	}
	return db.ExecContextWithBuilder(ctx, sqlmy.NewInsertBuilder(cardTable, sqlmy.Struct2AssignList(_cpl...)))
}
```

可以发现大概是这么几种类型:
- Entity: 完成单条记录Query的数据接收
- EntityParam: where 条件的构造, 更新和插入的实体载体
- EntityList: 完成多记录Query的数据接收
- EntityParamList: 完成多记录的创建


### 如何获取 dao

```
go get github.com/liuximu/sqlmy/tool/dao
```

## 单测
引入了 [sqlmock](github.com/DATA-DOG/go-sqlmock) 完成单测功能.具体用法请查看sqlmock


## 例子

[flashcard](github.com/liuximu/flashcard) 使用本类库完成数据库访问,如果需要可以参考.
