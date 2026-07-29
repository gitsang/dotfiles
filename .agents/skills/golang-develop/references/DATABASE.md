<!-- MarkdownTitleNumber auto -->

# 数据库操作

使用 [GORM](https://github.com/go-gorm/gorm) 和 [GORM gen](https://github.com/go-gorm/gen)。

## 1. GORM vs gen 选择

| 场景               | 选择                           |
| ------------------ | ------------------------------ |
| 简单 CRUD          | GORM 基础用法                  |
| 复杂查询、类型安全 | GORM gen（代码生成）           |
| 高频查询、团队协作 | GORM gen（生成类型安全的 DAO） |

## 2. GORM 基础用法

```go
import "gorm.io/gorm"

// Model 定义
type User struct {
    ID        uint           `gorm:"primaryKey"`
    Name      string         `gorm:"size:100"`
    Email     string         `gorm:"uniqueIndex"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"` // 软删除
}

// CRUD 操作
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
    var user User
    err := r.db.WithContext(ctx).First(&user, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, ErrNotFound  // 封装为 repository 层错误
    }
    return &user, err
}

func (r *UserRepository) Update(ctx context.Context, user *User) error {
    return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepository) Delete(ctx context.Context, id uint) error {
    return r.db.WithContext(ctx).Delete(&User{}, id).Error  // 软删除
}

// 分页查询
func (r *UserRepository) List(ctx context.Context, page, pageSize int) ([]User, int64, error) {
    var users []User
    var total int64

    r.db.WithContext(ctx).Model(&User{}).Count(&total)
    err := r.db.WithContext(ctx).
        Offset((page - 1) * pageSize).
        Limit(pageSize).
        Find(&users).Error
    return users, total, err
}
```

## 3. GORM gen 使用

```go
// 1. 生成代码配置 (gen.go)
import "gorm.io/gen"

func main() {
    g := gen.NewGenerator(gen.Config{
        OutPath: "./internal/repository/query",  // 输出路径
        Mode:    gen.DefaultMode,
    })

    g.UseDB(db)

    // 生成模型和 DAO
    g.ApplyBasic(
        model.User{},
        model.Order{},
    )

    // 自定义查询方法
    g.ApplyInterface(func(repository.UserQuerier) {}, model.User{})

    g.Execute()
}

// 2. 定义自定义查询接口
type UserQuerier interface {
    // WhereName query by name
    WhereName(name string) []*model.User
    // FindByEmail query by email
    FindByEmail(email string) (*model.User, error)
}

// 3. 使用生成的 DAO（类型安全）
func (s *UserService) GetByEmail(ctx context.Context, email string) (*User, error) {
    u := query.User
    user, err := u.WithContext(ctx).FindByEmail(email)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

## 4. 事务处理

```go
func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 创建订单
        order := &Order{UserID: req.UserID, Amount: req.Amount}
        if err := tx.Create(order).Error; err != nil {
            return err  // 自动回滚
        }

        // 更新用户余额
        if err := tx.Model(&User{}).
            Where("id = ?", req.UserID).
            Update("balance", gorm.Expr("balance - ?", req.Amount)).Error; err != nil {
            return err  // 自动回滚
        }

        return nil  // 提交
    })
}
```
