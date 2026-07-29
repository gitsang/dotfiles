<!-- MarkdownTitleNumber auto -->

# 测试规范

## 1. 单元测试适用场景

不要遵循测试驱动开发（TDD），只有当某个函数可能有大量用例/边界条件时，才添加单元测试。

先写代码、再写测试。

### 1.1 什么时候写单元测试

- 核心算法：计算、算法、数据转换等纯函数
- 容易出错的边界条件：空值、零值、负数、极大值、数组越界、字符串边界、日期时间处理（闰年、时区等）
- 公共 API / 库代码：供他人调用的接口，行为必须稳定可预测

### 1.2 什么时候不写单元测试

- 胶水代码： 简单的 CRUD 框架胶水代码，纯粹调用 ORM / 框架 API，自身没有逻辑
- 框架代码：初始化，配置加载，日志等框架内容等几乎所有项目都一样的框架代码
- 调用公共函数/库/API：测试已经在库代码测试过了，没必要再测一遍
- 简单的业务逻辑：几乎没有分支的业务代码
- 不能使用 table-driven 的测试：输出难以简单描述为一个值
- 字符串输出验证：日志输出、错误 Message 输出等用于人类阅读的字符串

## 2. 核心规则

| 规则                    | 说明                                               |
| ----------------------- | -------------------------------------------------- |
| 使用 table-driven tests | 所有测试用例放在一个表格中，代码简洁               |
| 测试文件命名            | 与被测文件同名: `snapshot.go` → `snapshot_test.go` |
| 不按函数命名            | 不要用 `get_snapshot_test.go`，应与文件同名        |
| 使用 want go            | 统一语言，便于测试过滤                             |

## 3. table-driven tests 示例

```go
// snapshot_test.go
func TestSnapshotCreate(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateInput
        want    *Snapshot
        wantErr error
    }{
        {
            name:  "valid input",
            input: CreateInput{Name: "test", TTL: 3600},
            want:  &Snapshot{Name: "test", TTL: 3600},
            wantErr: nil,
        },
        {
            name:  "empty name",
            input: CreateInput{Name: "", TTL: 3600},
            want:  nil,
            wantErr: ErrInvalidInput,
        },
        {
            name:  "negative ttl",
            input: CreateInput{Name: "test", TTL: -1},
            want:  nil,
            wantErr: ErrInvalidInput,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Create(tt.input)
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("Create() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## 4. 测试文件组织

```
internal/service/
├── user.go           # 业务代码
├── user_test.go      # user.go 的所有测试
├── order.go          # 业务代码
├── order_test.go     # order.go 的所有测试
```

禁止出现以下形式：

```
internal/service/
├── user.go           # 业务代码
├── create_test.go    # 用户创建业务测试（应该写到 user_test.go）
```
