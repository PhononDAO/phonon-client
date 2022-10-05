# API Errors

Errors are returned as an object with a key/message pair.

## Example Error Object

```
{
    key: "UNKNOWN ERROR",
    message: "Unknown Error"
}
```

# Error Keys and Messages

The following errors are returned from the API.

| HTTP Status Code | Key            | Message           | Notes                                          |
| ---------------- | -------------- | ----------------- | ---------------------------------------------- |
| 500              | UNKONWN_ERROR  | Unknown Error     |                                                |
| 400              | FIELD_REQUIRED | Field is required | This error will return for each required field |
