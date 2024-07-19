### Deployer

В пакете `deployer` размещаются реализации интерфейса `Deployer`:

```go
type Deployer interface {
    CreatePod(name string) error
    DeletePod(name string) error
    GetPodList() ([]string, error)
}
```

<br>

При возвращении ошибок должна использоваться следующая нотация:

```go
package mydeployer

// ...

func Operation() error {
    const op = "deployer.mydeployer.Operation"

    // ...
    if err != nil {
        return fmt.Errorf("%s: %w", op, err)
        // return fmt.Errorf("%s: %w", op, deployer.ErrCustom)
    }
    return nil
}

// ...
```

<br>

Также при описании метдов следует проверять соответствие интерфейсу:

```go
package mydeployer

type MyDeployer struct{}

// ...

var _ deployer.Deployer = (*MyDeployer)(nil)

func (d *MyDeployer) CreatePod(name string) error {
    // ...
    return nil
}

// ...
```
