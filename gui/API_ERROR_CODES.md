# API Errors

Errors are returned as an object with a code/message pair.

## Example Error Object

```
{
    code: 0,
    message: "Unknown Error"
}
```

# Error Codes

The following error codes are returned from the API.

| HTTP Status Code | Code | Message           | Notes                                          |
| ---------------- | ---- | ----------------- | ---------------------------------------------- |
| 500              | 0    | Unknown Error     |                                                |
| 400              | 1    | Field is required | This error will return for each required field |
