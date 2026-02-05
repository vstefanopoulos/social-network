# Custom Types Documentation

This package provides custom types for the social network backend, ensuring type safety, validation, and consistent serialization.

## Overview

All custom types implement the `Validator` interface, which requires a `Validate() error` method. Use `ValidateStruct()` to validate entire structs containing these types.

### Validator Interface

```go
type Validator interface {
    Validate() error
}
```

### ValidateStruct Function

Validates structs by iterating over exported fields and checking those that implement the `Validator` interface.

**Behavior**:
- Calls `Validate()` on fields implementing `Validator`.
- For fields without the `validate:"nullable"` tag, zero values are treated as errors.
- Nullable fields skip validation if empty.
- Primitives are excluded except slices of custom types.
- If a field is a slice of a custom type, if a null value is found in that slice validation error is returned.
- Allows zero values in type that are included in 'alwaysAllowZero' map.

**Tags**:
- `validate:"nullable"`: Marks the field as optional; zero values are allowed and skip validation.

**Example**:

```go
type RegisterRequest struct {
    Username  ct.Username  `json:"username,omitempty" validate:"nullable"` // optional
    FirstName ct.Name      `json:"first_name,omitempty" validate:"nullable"` // optional
    LastName  ct.Name      `json:"last_name"` // required
    About     ct.About     `json:"about"`     // required
    Email     ct.Email     `json:"email,omitempty" validate:"nullable"` // optional
}

err := ct.ValidateStruct(req)
if err != nil {
    // handle validation errors
}
```

**Notes**: Slice validation code is commented out in the implementation. Unexported fields are skipped.

## Types


### Id

**Description**: Represents an encrypted ID (int64). Allows null values in JSON but encrypts to a hash using hashids.

**Validation**: Must be > 0.

**Marshal/Unmarshal**: Marshals to encrypted string using hashids (requires `ENC_KEY` env var). Unmarshals from hash back to int64.

**Usage**: For secure ID transmission in APIs. Implements `Scan()` and `Value()` methods for use in postgress database calls.


### UnsafeId

**Description**: Represents an unencrypted ID (int64).

**Validation**: Must be > 0.

**Marshal/Unmarshal**: Standard JSON marshal/unmarshal as int64.

**Usage**: For internal use where encryption is not needed. Implements `Scan()` and `Value()` methods for use in postgress database calls.


### Ids

**Description**: Slice of `Id`.

**Validation**: Non-empty slice, all IDs must be valid.

**Marshal/Unmarshal**: As []int64.

**Usage**: For lists of IDs. Implements `Scan()` and `Value()` methods for use in postgress database calls.

**Extra Features**: Implements `Unique()` method that returns a copy of type 'Ids' only containing the unique entries of the given instance.


### About

**Description**: Bio or description text.

**Validation**: 3-300 characters, no control characters (except \n, \r, \t).

**Marshal/Unmarshal**: Standard string.

**Usage**: User bios.


### Audience

**Description**: Visibility level for posts/comments/events.

**Validation**: Must be one of: "everyone", "group", "followers", "selected" (case-insensitive), no control characters.

**Marshal/Unmarshal**: Standard string.

**Usage**: Content visibility.


### PostBody

**Description**: Body text for posts.

**Validation**: 3-500 characters, no control characters.

**Marshal/Unmarshal**: Standard string.

**Usage**: Post content.


### CommentBody

**Description**: Body text for comments.

**Validation**: 3-400 characters, no control characters.

**Marshal/Unmarshal**: Standard string.

**Usage**: Comment content.


### EventBody

**Description**: Body text for events.

**Validation**: 3-400 characters, no control characters.

**Marshal/Unmarshal**: Standard string.

**Usage**: Event descriptions.


### MsgBody

**Description**: Body text for messages.

**Validation**: 3-400 characters, no control characters.

**Marshal/Unmarshal**: Standard string.

**Usage**: Chat messages. Implements `Scan()` and `Value()` methods for use in postgress database calls.


### CtxKey

**Description**: Type alias for context keys to enforce naming conventions.

**Validation**: N/A.

**Marshal/Unmarshal**: N/A.

**Usage**: Context keys like `ClaimsKey`, `UserId`, etc.


### DateOfBirth

**Description**: User's date of birth.

**Validation**: Not zero, not in future, age 13-120.

**Marshal/Unmarshal**: "2006-01-02" format.

**Usage**: User profiles.


### EventDateTime

**Description**: Date and time for events.

**Validation**: Not zero, not in past, within 6 months ahead.

**Marshal/Unmarshal**: RFC3339.

**Usage**: Event scheduling.


### GenDateTime

**Description**: Generic nullable datetime.

**Validation**: Not zero.

**Marshal/Unmarshal**: RFC3339.

**Usage**: Timestamps like created_at. Implements `Scan()` and `Value()` methods for use in postgress database calls.


### Email

**Description**: Email address.

**Validation**: Valid email regex, no control characters.

**Marshal/Unmarshal**: Standard string.

**Usage**: User emails.


### Username

**Description**: Username.

**Validation**: 3-32 chars, alphanumeric + underscore, no control characters.

**Marshal/Unmarshal**: Standard string.

**Usage**: Usernames.


### Identifier

**Description**: Username or email.

**Validation**: Matches username or email regex.

**Marshal/Unmarshal**: Standard string.

**Usage**: Login identifiers.


### Name

**Description**: First or last name.

**Validation**: Min 2 chars, Unicode letters + apostrophes/hyphens/spaces, no control characters.

**Marshal/Unmarshal**: Standard string.

**Usage**: User names.


### Limit

**Description**: Pagination limit.

**Validation**: 1-500.

**Marshal/Unmarshal**: As int32.

**Usage**: API pagination. Implements `Scan()` and `Value()` methods for use in postgress database calls.


### Offset

**Description**: Pagination offset.

**Validation**: >= 0.

**Marshal/Unmarshal**: As int32.

**Usage**: API pagination. Implements `Scan()` and `Value()` methods for use in postgress database calls.


### Password

**Description**: Plain password.

**Validation**: 8-64 chars, requires uppercase, lowercase, digit, symbol, no control characters.

**Marshal/Unmarshal**: Marshals to "********".

**Usage**: Password input.


### HashedPassword

**Description**: Hashed password.

**Validation**: Non-empty, no control characters.

**Marshal/Unmarshal**: Marshals to "********".

**Usage**: Stored passwords.


### SearchTerm

**Description**: Search query term.

**Validation**: Min 2 chars, alphanumeric + spaces/hyphens, no control characters.

**Marshal/Unmarshal**: Standard string.

**Usage**: Search inputs.


### Title

**Description**: Title for groups/chats.

**Validation**: 1-50 chars, no control characters.

**Marshal/Unmarshal**: Standard string.

**Usage**: Group/chat titles.


### Error

**Description**: Custom error type that includes classification, cause, and context. It implements the error interface and supports error wrapping and classification.

**Validation**: N/A (not a Validator).

**Marshal/Unmarshal**: N/A.

**Usage**: For structured error handling with kinds like ErrNotFound, ErrInternal, etc. Supports errors.Is and errors.As.

**Methods**: 
#### Wrap():
Wrap creates a MediaError that classifies and optionally wraps an existing error.

Usage:
  - kind: the classification of the error (e.g., ErrFailed, ErrNotFound). If nil, ErrUnknownClass is used.
  - err: the underlying error to wrap; if nil, Wrap returns nil.
  - msg: optional context message describing where or why the error occurred.
Behavior:
  - If `err` is already a MediaError and `kind` is nil, it preserves the original Kind and optionally adds a new message.
  - Otherwise, it creates a new MediaError with the specified Kind, Err, and message.
  - The resulting MediaError supports errors.Is (matches Kind) and errors.As (type assertion) and preserves the wrapped cause.
  - If kind is nil and the err is not media error or lacks kind then kind is set to ErrUnknownClass.

It is recommended to only use nil kind if the underlying error is of type MediaError and its kind is not nil.   
   
Example:
```go
if err != nil {
    return Wrap(ErrReqValidation, err, "upload image:")
}
```

#### Public:
Returns a string containing only the kind field of Error. Recommeded use for APIs and handler respones
```go
err.(*ct.Error).Public()
```

#### Error():
Verbose and detailed string of error.
Behavior:
  - If `err` is already a MediaError and `kind` is nil, it preserves the original Kind and optionally adds a new message.
  - Otherwise, it creates a new MediaError with the specified Kind, Err, and message.
  - The resulting MediaError supports errors.Is (matches Kind) and errors.As (type assertion) and preserves the wrapped cause.
  - If kind is nil and the err is not media error or lacks kind then kind is set to ErrUnknownClass.

*It is recommended to only use nil kind if the underlying error is of type MediaError and its kind is not nil.*


### FileVisibility

**Description**: Represents the visibility level of a file. It can be either "private" or "public".

**Validation**: Must be "private" or "public".

**Marshal/Unmarshal**: Standard string. Implements JSON marshal/unmarshal and database Scan/Value.

**Usage**: File access control. Private files expire in 3 minutes, public in 6 hours.


### FileVariant

**Description**: Represents the variant or size of a file. It can be "original", "thumb", "small", "medium", or "large".

**Validation**: Must be one of the defined variants.

**Marshal/Unmarshal**: Standard string. Implements JSON marshal/unmarshal and database Scan/Value.

**Usage**: Image resizing variants for media files.


### UploadStatus

**Description**: Represents the status of a file upload process. It can be "pending", "processing", "complete", or "failed".

**Validation**: Must be one of the defined statuses.

**Marshal/Unmarshal**: Standard string. Implements JSON marshal/unmarshal and database Scan/Value.

**Usage**: Tracking file upload progress.