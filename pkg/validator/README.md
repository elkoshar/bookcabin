# Validator

- Prepare Struct
```go
type User struct {
    Name    string  `validate:"string,required"`
    Age     int     `validate:"numeric,required"`
    Sex     string  `validate:"-"` //this will not validated
}

u := User{
    Name : "", // this will fail the validator since it required
    Age : 10,
    Sex : "F",
}
```
- Pass struct
```go
result, err := validator.ValidateStruct(u)
if err != nil {
    return err
}
```

More Info about validator : https://github.com/go-playground/validator