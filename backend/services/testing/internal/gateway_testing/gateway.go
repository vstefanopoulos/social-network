package gateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/cookiejar"
	"slices"
	"social-network/services/testing/internal/configs"
	"social-network/services/testing/internal/utils"
	"social-network/shared/go/ct"
	"sync"
	"time"
)

func StartTest(ctx context.Context, cfgs configs.Configs) {
	var wg sync.WaitGroup
	wg.Go(func() { utils.HandleErr("api-gateway auth flow", ctx, testAuthFlow) })
	time.Sleep(time.Second) //sleeping so that ratelimiting caused by the next tests doesn't affect the above test
	wg.Go(func() { utils.HandleErr("api-gateway posts flow", ctx, testPostsFlow) })
	time.Sleep(time.Second)
	wg.Go(func() { utils.HandleErr("api-gateway random register", ctx, randomRegister) })
	wg.Go(func() { utils.HandleErr("api-gateway random login", ctx, randomLogin) })

	wg.Wait()

}

func testAuthFlow(ctx context.Context) error {
	fmt.Println("api-gateway Start test auth flow")

	// Create client with cookie jar to persist cookies between requests
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("failed to create cookie jar: %w", err)
	}
	client := &http.Client{Jar: jar}

	// 1. Register
	registerData := newRegisterReq()

	resp, err := postJSON(client, "http://api-gateway:8081/register", registerData)
	if err != nil {
		return fmt.Errorf("register failed: %w, body: %s", err, "bodyToString(resp)")
	}

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("register failed: bad status, %w, body: %s", err, bodyToString(resp))
	}

	email, _ := (*registerData)["email"].(string)
	pass, _ := (*registerData)["password"].(string)

	// 2. Login
	loginData := map[string]any{
		"identifier": email,
		"password":   pass,
	}
	resp, err = postJSON(client, "http://api-gateway:8081/login", loginData)
	if err != nil {
		return fmt.Errorf("login failed: %w, body: %s", err, bodyToString(resp))
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("login failed: bad status, %w, body: %s", err, bodyToString(resp))
	}

	// 4. Logout
	resp, err = postJSON(client, "http://api-gateway:8081/logout", nil)
	if err != nil {
		return fmt.Errorf("logout failed: %w, body: %s", err, bodyToString(resp))
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("logout failed: bad status, %w, body: %s", err, bodyToString(resp))
	}

	fmt.Println("api-gateway Finished test auth flow")
	return nil
}

func postJSON(client *http.Client, url string, data any) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}

func newRegisterReq() *map[string]any {
	registerData := map[string]any{
		"username":      utils.Title(utils.RandomString(10, false)),
		"first_name":    utils.Title(utils.RandomString(10, false)),
		"last_name":     utils.Title(utils.RandomString(10, false)),
		"date_of_birth": fmt.Sprintf("%d-%02d-%02d", rand.IntN(50)+1950, rand.IntN(11)+1, rand.IntN(20)+1),
		"about":         utils.RandomString(300, true),
		"public":        rand.IntN(3) == 0,
		"email":         utils.RandomString(10, false) + "@hotmail.com",
		"password":      utils.RandomPassword(),
	}

	return &registerData
}

func newLoginReq() *map[string]any {
	login := map[string]any{
		"identifier": utils.Title(utils.RandomString(10, false)),
		"password":   utils.RandomString(10, true),
	}
	return &login
}

func bodyToString(r *http.Response) string {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	return string(body)
}
func testPostsFlow(ctx context.Context) error {
	fmt.Println("posts Start test posts flow")

	client := newFakeClient("http://api-gateway:8081")

	// -2. Register
	registerData := newRegisterReq()
	_, err := client.DoRequest("POST", "/register", []int{200, 201}, "register",
		*registerData,
	)
	if err != nil {
		return err
	}

	// -1. Login
	email, _ := (*registerData)["email"].(string)
	pass, _ := (*registerData)["password"].(string)
	loginData := map[string]any{
		"identifier": email,
		"password":   pass,
	}
	loginResp, err := client.DoRequest("POST", "/login", []int{200, 201}, "login", loginData)
	if err != nil {
		return err
	}

	myId, err := fetchKey(loginResp, "id")
	if err != nil {
		return fmt.Errorf("fetch failed %w", err)
	}

	// 1. Get Non-Existent Post
	postStringId, _ := ct.EncodeId(ct.Id(97862345))
	_, err = client.DoRequest("GET", "/post/"+postStringId, []int{500, 404}, "get non-existent post", //TODO this should only be 404
		map[string]any{},
	)
	if err != nil {
		return err
	}

	// 2. Edit Non-Existent Post
	editPostData := map[string]any{
		"post_body": "new text",
	}

	_, err = client.DoRequest("POST", "/posts/"+postStringId, []int{400, 404}, "2 edit non-existent post", editPostData)
	if err != nil {
		return err
	}

	// 3. Delete Non-Existent Post
	_, err = client.DoRequest("DELETE", "/posts/"+postStringId, []int{400, 404, 500}, "3 delete non-existent post", editPostData) //TODO needs fix, no 500
	if err != nil {
		return err
	}

	// 4. Get Comments for Non-Existent Post
	_, err = client.DoRequest("GET", "/comments", []int{400, 404, 500}, "4 get comments for non-existent post", editPostData) //TOOD no 500
	if err != nil {
		return err
	}

	// 5. Create Comment on Non-Existent Post
	createCommentData := map[string]any{
		"parent_id":    ct.Id(97862345),
		"comment_body": "This is a comment on a non-existent post.",
	}

	_, err = client.DoRequest("POST", "/comments/create", []int{400, 404, 500}, "5 create comment on non-existent post", createCommentData) //TOOD no 500
	if err != nil {
		return err
	}

	// 6. Edit Non-Existent Comment
	_, err = client.DoRequest("POST", "/comments/edit", []int{400, 404, 500}, "6 edit non-existent comment", createCommentData) //TODO no 500
	if err != nil {
		return err
	}

	// 7. Delete Non-Existent Comment
	_, err = client.DoRequest("POST", "/comments/delete", []int{400, 404, 500}, "7 delete non-existent comment", createCommentData) //TODO no 500
	if err != nil {
		return err
	}

	// 8. Create Post with Invalid Data
	_, err = client.DoRequest("POST", "/posts/create", []int{400, 422}, "8 create post with invalid data", nil)
	if err != nil {
		return err
	}

	// 9. Create Post
	_, err = client.DoRequest("POST", "/posts/create", []int{200}, "9 create post",
		map[string]any{
			"post_body": utils.RandomString(100, false),
			"audience":  "everyone",
		},
	)
	if err != nil {
		return err
	}

	// 9,5. Create Post
	resp, err := client.DoRequest("POST", "/user/posts", []int{200}, "9,5 get my post feed",
		map[string]any{
			"creator_id": myId,
			"limit":      1,
			"offset":     0,
		},
	)
	if err != nil {
		return err
	}

	postIds, err := fetchKeyArray(resp)
	if err != nil {
		return fmt.Errorf("failed to fetch post id, %w", err)
	}
	slice, ok := postIds.([]any)
	if !ok {
		return fmt.Errorf("failed to cast slice for post id, %w", err)
	}
	mapAny := slice[0]
	postMap, ok := mapAny.(map[string]any)
	if !ok {
		return fmt.Errorf("failed to cast post map for post id, %w", err)
	}
	postIdAny := postMap["post_id"]

	// 10. Get Post By ID
	_, err = client.DoRequest("POST", "/post/", []int{200}, "getting post by id",
		map[string]any{
			"entity_id": postIdAny,
		},
	)
	if err != nil {
		return fmt.Errorf("get post failed: %w, body: %s", err, bodyToString(resp))
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("expected 200 for get post, got %d", resp.StatusCode)
	}

	// 11. Edit Post
	_, err = client.DoRequest("POST", "/posts/edit", []int{200}, "edit post", map[string]any{
		"post_id":   postIdAny,
		"post_body": "yoooooooooo",
		"audience":  "everyone",
	})
	if err != nil {
		return err
	}

	// 15. Create Comment on Post
	resp, err = client.DoRequest("POST", "/comments/create", []int{200}, "create comment", map[string]any{
		"parent_id":    postIdAny,
		"comment_body": "comment",
	})
	if err != nil {
		return err
	}

	// 16. Get Comments By Parent ID
	resp, err = client.DoRequest("POST", "/comments/", []int{200}, "get comment by parent id", map[string]any{
		"entity_id": postIdAny,
		"limit":     1,
		"offset":    0,
	})
	if err != nil {
		return fmt.Errorf("get comments failed: %w, body: %s", err, bodyToString(resp))
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("expected 200 for get comments, got %d", resp.StatusCode)
	}

	// comment_id"`
	// 		Body        ct.CommentBody `json:"comment_body"`

	// 19. Edit Comment
	// _, err = client.DoRequest("POST", "/comments/edit", []int{200}, "edit comment", map[string]any{
	// 	"comment_id":      commentID,
	// 	"content": "",
	// })
	// if err != nil {
	// 	return err
	// }

	// // 20. Edit Comment with Invalid Data
	// _, err = client.DoRequest("POST", "/comments/"+commentID, []int{400, 422}, "edit comment with invalid data", map[string]any{
	// 	"id": commentID,
	// })
	// if err != nil {
	// 	return err
	// }

	// // 21. Edit Nested Comment
	// _, err = client.DoRequest("POST", "/comments/"+nestedCommentID, []int{200}, "edit nested comment", map[string]any{
	// 	"id":      nestedCommentID,
	// 	"content": "",
	// })
	// if err != nil {
	// 	return err
	// }

	// // 22. Delete Nested Comment
	// _, err = client.DoRequest("POST", "/comments/delete", []int{200}, "delete nested comment", map[string]any{
	// 	"id": nestedCommentID,
	// })
	// if err != nil {
	// 	return err
	// }

	// // 23. Delete Nested Comment Again
	// _, err = client.DoRequest("POST", "/comments/delete", []int{400, 404}, "delete nested comment twice", map[string]any{
	// 	"id": nestedCommentID,
	// })
	// if err != nil {
	// 	return err
	// }

	// // 24. Edit Deleted Nested Comment
	// _, err = client.DoRequest("POST", "/comments/"+nestedCommentID, []int{400, 404}, "edit deleted nested comment", map[string]any{
	// 	"id":      nestedCommentID,
	// 	"content": "",
	// })
	// if err != nil {
	// 	return err
	// }

	// // 25. Get Comments for Comment After Deletion
	// resp, err = client.DoRequest("POST", "/comments/"+commentID+"/comments?page=1&limit=10")
	// if err != nil {
	// 	return fmt.Errorf("get nested comments after deletion failed: %w, body: %s", err, bodyToString(resp))
	// }
	// if resp.StatusCode != 200 {
	// 	return fmt.Errorf("expected 200 for get nested comments after deletion, got %d", resp.StatusCode)
	// }

	// // 26. Delete Comment
	// _, err = client.DoRequest("POST", "/comments/delete", []int{200}, "delete comment", map[string]any{
	// 	"id": commentID,
	// })
	// if err != nil {
	// 	return err
	// }

	// // 27. Delete Comment Again
	// _, err = client.DoRequest("POST", "/comments/delete", []int{400, 404}, "delete comment twice", map[string]any{
	// 	"id": commentID,
	// })
	// if err != nil {
	// 	return err
	// }

	// // 28. Edit Deleted Comment
	// _, err = client.DoRequest("POST", "/comments/"+commentID, []int{400, 404}, "edit deleted comment", map[string]any{
	// 	"id":      commentID,
	// 	"content": "",
	// })
	// if err != nil {
	// 	return err
	// }

	// // 29. Create Comment on Deleted Comment
	// _, err = client.DoRequest("POST", "/comments", []int{400, 404}, "create comment on deleted comment", map[string]any{
	// 	"parent_id": commentID,
	// 	"content":   "",
	// })
	// if err != nil {
	// 	return err
	// }

	// // 30. Get Comments for Post After Comment Deletion
	// resp, err = client.DoRequest("POST", "/posts/"+postID+"/comments?page=1&limit=10")
	// if err != nil {
	// 	return fmt.Errorf("get comments after deletion failed: %w, body: %s", err, bodyToString(resp))
	// }
	// if resp.StatusCode != 200 {
	// 	return fmt.Errorf("expected 200 for get comments after deletion, got %d", resp.StatusCode)
	// }

	// // 31. Pagination Invalid Parameters
	// _, _ = client.DoRequest("POST", "/posts/"+postID+"/comments?page=-1&limit=10")
	// _, _ = client.DoRequest("POST", "/posts/"+postID+"/comments?page=1&limit=0")
	// _, _ = client.DoRequest("POST", "/posts/"+postID+"/comments?page=1&limit=1000")

	// // 32. Delete Post
	// _, err = client.DoRequest("POST", "/posts/delete", []int{200}, "delete post", map[string]any{
	// 	"id": postID,
	// })
	// if err != nil {
	// 	return err
	// }

	// // 33. Delete Post Again
	// _, err = client.DoRequest("POST", "/posts/delete", []int{400, 404}, "delete post twice", map[string]any{
	// 	"id": postID,
	// })
	// if err != nil {
	// 	return err
	// }

	// // 34. Get Deleted Post
	// resp, err = client.DoRequest("POST", "/posts/"+postID)
	// if err != nil {
	// 	return fmt.Errorf("get deleted post failed: %w, body: %s", err, bodyToString(resp))
	// }
	// if resp.StatusCode == 200 {
	// 	return fmt.Errorf("expected error for getting deleted post, got 200")
	// }

	// // 35. Edit Deleted Post
	// _, err = client.DoRequest("POST", "/posts/"+postID, []int{400, 404}, "edit deleted post", map[string]any{
	// 	"id":      postID,
	// 	"content": "",
	// })
	// if err != nil {
	// 	return err
	// }

	// // 36. Get Comments for Deleted Post
	// _, _ = client.DoRequest("POST", "/posts/"+postID+"/comments?page=1&limit=10")

	// // 37. Create Comment on Deleted Post
	// _, err = client.DoRequest("POST", "/comments", []int{400, 404}, "create comment on deleted post", map[string]any{
	// 	"parent_id": postID,
	// 	"content":   "",
	// })

	return err
}

func marshalJSON(data any) []byte {
	body, _ := json.Marshal(data)
	return body
}

func randomRegister(ctx context.Context) error {
	fmt.Println("api-gateway starting register test")
	client := &http.Client{}
	gotRateLimited := false
	for range 100 {
		registerData := newRegisterReq()
		resp, err := postJSON(client, "http://api-gateway:8081/register", registerData)
		if err != nil {
			return fmt.Errorf("spam register failed: %w", err)
		}

		if resp.StatusCode/200 != 1 && resp.StatusCode != 429 {
			return fmt.Errorf("bad status when spam registering: %d, body: %s", resp.StatusCode, bodyToString(resp))
		}

		if resp.StatusCode == 429 {
			gotRateLimited = true
		}
		time.Sleep(time.Millisecond * 50)
	}

	if gotRateLimited == false {
		return fmt.Errorf("register spam didn't get ratelimited when spamming?!")
	}

	fmt.Println("api-gateway spam api-gateway register test passed")
	return nil
}

func randomLogin(ctx context.Context) error {
	fmt.Println("api-gateway starting Login test")
	client := &http.Client{}
	gotRateLimited := false
	for range 100 {
		loginReq := newLoginReq()
		resp, err := postJSON(client, "http://api-gateway:8081/login", loginReq)
		if err != nil {
			return fmt.Errorf("spam login failed: %w", err)
		}
		if resp.StatusCode == 200 {
			return fmt.Errorf("somehow managed to login while spamming bad logins??: %d, body: %s", resp.StatusCode, bodyToString(resp))
		}
		if resp.StatusCode == 429 {
			gotRateLimited = true
		}
		time.Sleep(time.Millisecond * 50)
	}

	if gotRateLimited == false {
		return fmt.Errorf("login spam didn't get ratelimited when spamming?!")
	}

	fmt.Println("api-gateway spam login test passed")
	return nil
}

type fakeClient struct {
	baseUrl   string
	client    *http.Client
	printSucc bool
}

func newFakeClient(baseUrl string) fakeClient {
	jar, _ := cookiejar.New(nil)
	return fakeClient{
		baseUrl: baseUrl,
		client: &http.Client{
			Jar: jar,
		},
	}
}

func (cl *fakeClient) DoRequest(
	method string,
	url string,
	expectedStates []int,
	label string,
	data map[string]any,
) (*http.Response, error) {
	var resp *http.Response
	var err error

	switch method {
	case "POST":
		resp, err = postJSON(cl.client, cl.baseUrl+url, data)
	default:
		return nil, fmt.Errorf("%s at %s failed: unsupported method %s", label, url, method)
	}

	if err != nil {
		str := "[resp was nil...]"
		if resp != nil {
			str = bodyToString(resp)
		}
		return resp, fmt.Errorf("%s at %s failed: %w, body: %s", label, url, err, str)
	}

	if !slices.Contains(expectedStates, resp.StatusCode) {
		return resp, fmt.Errorf(
			"%s at %s failed: expected status %v, got %d, body: %s",
			label,
			url,
			expectedStates,
			resp.StatusCode,
			bodyToString(resp),
		)
	}

	return resp, nil
}

// fetchKey extracts the value of a specific key from the request body.
func fetchKey(r *http.Response, key string) (any, error) {
	var body map[string]any
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	value, exists := body[key]
	if !exists {
		return nil, fmt.Errorf("key %q not found", key)
	}
	return value, nil
}

// fetchKey extracts the value of a specific key from the request body.
func fetchKeyArray(r *http.Response) (any, error) {
	var body []any
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}
	return body, nil
}
