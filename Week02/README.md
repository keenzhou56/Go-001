学习笔记

问题：

1. 我们在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。为什么，应该怎么做请写出代码？

ErrNoRows 定义在 "database/sql" 包，是个sentinel error

    sql.go link     -- https://github.com/golang/go/blob/go1.15.5/src/database/sql/sql.go
    sql.go 388 line -- var ErrNoRows = errors.New("sql: no rows in result set")

对于在dao层对数据库进行操作，属于底层封装，类似于工具、通用包，在此层的err不建议直接wrap或者降级处理，直接将该err往上抛，由业务层进行wrap或降级处理。
