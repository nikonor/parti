# lockfile #

Пакет реализует простой автоматический метод блокировки по PID

### Пример
```go
    if lock, err := lockfile.Lock(filename); err != nil {
        panic(err)
    } else {
        defer lock.Unlock()
    }
```
