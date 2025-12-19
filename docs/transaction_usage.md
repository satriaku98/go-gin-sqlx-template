# Transaction Manager Usage Guide

## Overview

Transaction Manager memungkinkan Anda untuk menjalankan multiple repository operations dalam satu transaksi atomik. Semua operasi akan berhasil bersama-sama (commit) atau gagal bersama-sama (rollback).

## Kapan Menggunakan Transaction?

Gunakan transaction ketika:

1. **Multiple Repository Operations**: Anda perlu memodifikasi data di multiple tables/repositories
2. **Atomicity Required**: Semua operasi harus berhasil atau gagal bersama-sama
3. **Data Consistency**: Anda perlu menjaga konsistensi data antar tables

**Contoh Use Cases:**
- Membuat user dan profile-nya secara bersamaan
- Transfer saldo antar akun (debit satu akun, credit akun lain)
- Membuat order dengan multiple order items
- Update inventory dan create audit log

## Cara Menggunakan

### 1. Setup di Usecase

Inject `database.Transactor` ke dalam usecase Anda:

```go
type userUsecase struct {
    userRepo     repository.UserRepository
    txManager    database.Transactor  // Add this
    // ... other dependencies
}

func NewUserUsecase(
    userRepo repository.UserRepository,
    txManager database.Transactor,  // Add this parameter
    // ... other parameters
) usecase.UserUsecase {
    return &userUsecase{
        userRepo:  userRepo,
        txManager: txManager,
        // ... other fields
    }
}
```

### 2. Gunakan WithTransaction

Wrap operasi repository Anda dalam `WithTransaction`:

```go
func (u *userUsecase) CreateUserWithProfile(ctx context.Context, req CreateUserRequest) error {
    return u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
        // Semua operasi di dalam function ini menggunakan transaction yang sama
        
        // Operation 1: Create user
        user := &model.User{...}
        if err := u.userRepo.Create(txCtx, user); err != nil {
            return err // Auto rollback
        }
        
        // Operation 2: Create profile
        profile := &model.Profile{UserID: user.ID, ...}
        if err := u.profileRepo.Create(txCtx, profile); err != nil {
            return err // Auto rollback (user creation juga di-rollback)
        }
        
        // Jika return nil, semua operasi akan di-commit
        return nil
    })
}
```

### 3. Dependency Injection di Container

Update `container.go` untuk membuat dan inject transaction manager:

```go
func NewContainer(cfg config.Config, log *logger.Logger, db *database.Database) *Container {
    // Create transaction manager
    txManager := db.NewTransactionManager()
    
    // Inject ke repository
    userRepo := postgres.NewUserRepository(db.DB, txManager)
    
    // Inject ke usecase
    userUsecase := impl.NewUserUsecase(userRepo, txManager, ...)
    
    // ...
}
```

## Best Practices

### ✅ DO

1. **Keep Transactions Short**: Hanya operasi database yang perlu atomicity
   ```go
   err := u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
       // ✅ Database operations only
       if err := u.userRepo.Create(txCtx, user); err != nil {
           return err
       }
       if err := u.profileRepo.Create(txCtx, profile); err != nil {
           return err
       }
       return nil
   })
   
   // ✅ Non-critical operations outside transaction
   u.pubsubClient.Publish(ctx, "user.created", data)
   ```

2. **Use txCtx for All Repository Calls**: Gunakan context dari transaction function
   ```go
   u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
       // ✅ Use txCtx
       u.userRepo.Create(txCtx, user)
       u.profileRepo.Create(txCtx, profile)
   })
   ```

3. **Return Errors Immediately**: Error akan trigger automatic rollback
   ```go
   if err := u.userRepo.Create(txCtx, user); err != nil {
       return err // ✅ Auto rollback
   }
   ```

### ❌ DON'T

1. **Don't Put Non-Database Operations in Transaction**
   ```go
   // ❌ BAD: External calls inside transaction
   u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
       u.userRepo.Create(txCtx, user)
       u.emailService.SendWelcomeEmail(user.Email) // ❌ External call
       u.pubsubClient.Publish(ctx, "user.created") // ❌ External call
       return nil
   })
   
   // ✅ GOOD: External calls outside transaction
   err := u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
       return u.userRepo.Create(txCtx, user)
   })
   if err == nil {
       u.emailService.SendWelcomeEmail(user.Email)
       u.pubsubClient.Publish(ctx, "user.created")
   }
   ```

2. **Don't Use Original ctx Inside Transaction**
   ```go
   u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
       u.userRepo.Create(ctx, user)      // ❌ Using ctx instead of txCtx
       u.profileRepo.Create(txCtx, profile) // ✅ Using txCtx
   })
   ```

3. **Don't Ignore Errors**
   ```go
   u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
       u.userRepo.Create(txCtx, user) // ❌ Not checking error
       return nil
   })
   ```

## Error Handling

Transaction manager automatically handles:

1. **Rollback on Error**: Jika function return error, transaction di-rollback
2. **Rollback on Panic**: Jika terjadi panic, transaction di-rollback dan panic di-throw ulang
3. **Commit on Success**: Jika function return nil, transaction di-commit

```go
err := u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
    if err := u.userRepo.Create(txCtx, user); err != nil {
        return err // Automatic rollback
    }
    
    if err := u.profileRepo.Create(txCtx, profile); err != nil {
        return fmt.Errorf("failed to create profile: %w", err) // Automatic rollback
    }
    
    return nil // Automatic commit
})

if err != nil {
    // Handle error - transaction already rolled back
    return err
}

// Success - transaction already committed
```

## Nested Transactions

Transaction manager supports nested `WithTransaction` calls. Jika sudah dalam transaction, nested call akan reuse transaction yang sama:

```go
func (u *userUsecase) CreateUserWithProfile(ctx context.Context) error {
    return u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
        // Create user
        if err := u.createUser(txCtx); err != nil {
            return err
        }
        
        // Create profile
        if err := u.createProfile(txCtx); err != nil {
            return err
        }
        
        return nil
    })
}

func (u *userUsecase) createUser(ctx context.Context) error {
    // This will reuse the existing transaction if called within WithTransaction
    return u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
        return u.userRepo.Create(txCtx, user)
    })
}
```

## Example: Complex Transaction

Contoh lengkap dengan multiple repositories:

```go
func (u *orderUsecase) CreateOrder(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error) {
    var order *model.Order
    
    err := u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
        // 1. Create order
        order = &model.Order{
            CustomerID: req.CustomerID,
            TotalAmount: req.TotalAmount,
            Status: "pending",
        }
        if err := u.orderRepo.Create(txCtx, order); err != nil {
            return fmt.Errorf("failed to create order: %w", err)
        }
        
        // 2. Create order items
        for _, item := range req.Items {
            orderItem := &model.OrderItem{
                OrderID: order.ID,
                ProductID: item.ProductID,
                Quantity: item.Quantity,
                Price: item.Price,
            }
            if err := u.orderItemRepo.Create(txCtx, orderItem); err != nil {
                return fmt.Errorf("failed to create order item: %w", err)
            }
        }
        
        // 3. Update inventory
        for _, item := range req.Items {
            if err := u.inventoryRepo.DecrementStock(txCtx, item.ProductID, item.Quantity); err != nil {
                return fmt.Errorf("failed to update inventory: %w", err)
            }
        }
        
        // 4. Create audit log
        auditLog := &model.AuditLog{
            Action: "order_created",
            OrderID: order.ID,
            UserID: req.CustomerID,
        }
        if err := u.auditLogRepo.Create(txCtx, auditLog); err != nil {
            return fmt.Errorf("failed to create audit log: %w", err)
        }
        
        // All operations succeed - will commit
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    
    // Transaction committed, send notifications
    u.notificationService.SendOrderConfirmation(order.ID)
    
    return order.ToResponse(), nil
}
```

## Troubleshooting

### Repository Tidak Menggunakan Transaction

**Problem**: Repository operations tidak menggunakan transaction meskipun dipanggil dalam `WithTransaction`

**Solution**: Pastikan repository menggunakan `getExecutor(ctx)` untuk mendapatkan DB atau TX:

```go
// ❌ BAD: Always uses DB
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    _, err := r.db.NamedQueryContext(ctx, query, args)
    return err
}

// ✅ GOOD: Uses TX if in transaction, DB otherwise
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
    _, err := sqlx.NamedQueryContext(ctx, r.getExecutor(ctx), query, args)
    return err
}
```

### Deadlock

**Problem**: Transaction mengalami deadlock

**Solution**: 
1. Keep transactions short
2. Access tables in consistent order
3. Avoid long-running operations in transactions

### Transaction Timeout

**Problem**: Transaction timeout karena terlalu lama

**Solution**:
1. Pindahkan non-database operations keluar dari transaction
2. Optimize database queries
3. Consider breaking into smaller transactions if possible

## Reference

Lihat contoh implementasi lengkap di:
- [transaction.go](file:///d:/Personal%20Repo/go-gin-sqlx-template/pkg/database/transaction.go) - Transaction manager implementation
- [user_repository.go](file:///d:/Personal%20Repo/go-gin-sqlx-template/internal/repository/postgres/user_repository.go) - Repository with transaction support
- [user_usecase_impl.go](file:///d:/Personal%20Repo/go-gin-sqlx-template/internal/usecase/impl/user_usecase_impl.go) - Usecase example (`CreateUserWithTransaction` method)
- [container.go](file:///d:/Personal%20Repo/go-gin-sqlx-template/cmd/api/container.go) - Dependency injection setup
