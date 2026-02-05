# retrieveusers

This document outlines how to integrate and use the `retrieveusers` package in any Go service to efficiently fetch user information.

## Overview

The `retrieveusers` package provides a `UserRetriever` struct that fetches user details (ID, Username, Avatar ID) and their corresponding avatar URLs. It optimizes performance by:
1.  **Caching**: Checking Redis for existing user data.
2.  **Batching**: Fetching missing data via batch gRPC calls to the Users Service.
3.  **Media Integration**: Automatically resolving avatar image URLs using the `retrievemedia` package.

## Prerequisites

To use `retrieveusers`, your service needs specific dependencies available, typically initialized during startup:

1.  **Users Service Client**: A gRPC client capable of making batch user info requests (`GetBatchBasicUserInfo`).
2.  **Redis Client**: A shared Redis client wrapper (`*redis_connector.RedisClient`).
3.  **Media Retriever**: An instance of `*retrievemedia.MediaRetriever` (used for fetching image URLs).

## Integration Steps

### 1. Import the Package

```go
import "social-network/backend/shared/go/retrieveusers"
```

### 2. Initialize `UserRetriever`

Instantiate `UserRetriever` in your application constructor or dependency injection setup. It is commonly stored alongside other service clients.

**Signature:**
```go
func NewUserRetriever(
    clients UsersBatchClient,
    cache *redis_connector.RedisClient,
    mediaRetriever *retrievemedia.MediaRetriever,
    ttl time.Duration,
) *UserRetriever
```

**Example Initialization (in `application.go` or similar):**

```go
// application.go

type Application struct {
    // ... other fields
    userRetriever *retrieveusers.UserRetriever
}

func NewApplication(
    // ... config
    redisClient *redis_connector.RedisClient,
    usersClient myservice.UsersClient, // Implements UsersBatchClient
    mediaRetriever *retrievemedia.MediaRetriever,
) *Application {
    
    // Valid cache duration, e.g., 5 minutes
    userCacheTTL := 5 * time.Minute

    return &Application{
        // ...
        userRetriever: retrieveusers.NewUserRetriever(
            usersClient, 
            redisClient, 
            mediaRetriever, 
            userCacheTTL,
        ),
    }
}
```

### 3. Usage

Use the `GetUsers` method to enrich data with user details. It takes a list of User IDs and returns a map of User objects.

**Method Signature:**
```go
func (h *UserRetriever) GetUsers(ctx context.Context, userIDs ct.Ids) (map[ct.Id]models.User, error)
```

**Example Usage (Generic Enrichment):**

```go
func (a *Application) GetItems(ctx context.Context, args AnyArgs) ([]ItemDTO, error) {
    // 1. Fetch raw items from DB/Source
    items, err := a.db.GetItems(ctx, args)
    if err != nil {
        return nil, err
    }

    // 2. Collect unique User IDs from items
    var userIds ct.Ids
    for _, item := range items {
        userIds = append(userIds, ct.Id(item.CreatorID))
    }

    // 3. Retrieve User Info (Map: UserID -> User Model)
    userMap, err := a.userRetriever.GetUsers(ctx, userIds)
    if err != nil {
        // Handle error (log it, return partial data, or fail)
        return nil, err
    }

    // 4. Enrich items with User data
    result := make([]ItemDTO, 0, len(items))
    for _, item := range items {
        dto := ItemDTO{
            ID:   item.ID,
            Data: item.Data,
        }
        
        // Attach user info if found
        if user, found := userMap[ct.Id(item.CreatorID)]; found {
            dto.CreatorName = string(user.Username)
            dto.AvatarURL   = user.AvatarURL
        }
        
        result = append(result, dto)
    }

    return result, nil
}
```

## Internal Workflow

When `GetUsers` is called:
1.  **Redis Check**: It looks for keys like `basic_user_info:<id>`.
2.  **Fetch Missing**: If keys are missing, it calls `clients.GetBatchBasicUserInfo` for those specific IDs.
3.  **Cache Update**: New user data is cached in Redis with the configured TTL.
4.  **Fetch Media**: It extracts `AvatarId` from the users and calls `mediaRetriever.GetImages` to resolve the actual image URLs (which also has its own caching layer).
5.  **Merge**: Returns the complete map of `models.User` with populated `AvatarURL`s.