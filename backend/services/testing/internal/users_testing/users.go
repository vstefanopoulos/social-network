package users_test

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"

	"social-network/services/testing/internal/configs"
	"social-network/services/testing/internal/utils"
	"social-network/shared/gen-go/users"
	"social-network/shared/go/ct"
	"social-network/shared/go/gorpc"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var UsersService users.UserServiceClient

func StartTest(ctx context.Context, cfgs configs.Configs) error {
	var err error
	UsersService, err = gorpc.GetGRpcClient(
		users.NewUserServiceClient,
		cfgs.UsersGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to users service: %s", err.Error())
	}

	var wg sync.WaitGroup
	wg.Go(func() { utils.HandleErr("users random register", ctx, randomRegister) })
	wg.Go(func() { utils.HandleErr("users random login", ctx, randomLogin) })
	wg.Go(func() { utils.HandleErr("users auth flow", ctx, registerLogin) })
	wg.Wait()
	return nil
}

var fail = "FAIL TEST: err ->"

func randomRegister(ctx context.Context) error {
	fmt.Println("users-service starting register test")
	for range 100 {
		req := newRegisterReq()
		resp, err := UsersService.RegisterUser(ctx, req)
		if err != nil {
			return errors.New(fail + err.Error())
		}

		if resp.UserId < 1 {
			return errors.New(fail)
		}

	}

	fmt.Println("users-service random register test passed")
	return nil
}

func randomLogin(ctx context.Context) error {
	fmt.Println("users-service starting Login test")
	for range 100 {
		req := newLoginReq()
		_, err := UsersService.LoginUser(ctx, req)
		if err != nil && !strings.Contains(err.Error(), "wrong credentials") {
			return errors.New(fail + "wrong type of error!: " + err.Error())
		}
		if err == nil {
			return errors.New(fail + "expected error! cause these random logins should all be failing!")
		}
	}

	fmt.Println("users-service random login test passed")
	return nil
}

func registerLogin(ctx context.Context) error {
	reg := newRegisterReq()

	_, err := UsersService.RegisterUser(ctx, reg)
	if err != nil {
		return errors.New(fail + err.Error())
	}

	log := newLoginReq()
	log.Identifier = reg.Email
	log.Password = reg.Password

	resp, err := UsersService.LoginUser(ctx, log)
	if err != nil {
		return errors.New(fail + "should have worked! err should be nil:" + err.Error())
	}
	if resp.Username != reg.Username {
		return errors.New(fail + "incorrect login, these two should be the same: `" + resp.Username + "` <-> `" + reg.Username + "`")
	}

	if resp.UserId == 0 || resp.Username == "" {
		return errors.New("found empty values")
	}

	fmt.Println("users-service passed simple reg login test")
	return nil
}

func newRegisterReq() *users.RegisterUserRequest {
	req := users.RegisterUserRequest{
		Username:    utils.Title(utils.RandomString(10, false)),
		FirstName:   utils.Title(utils.RandomString(10, false)),
		LastName:    utils.Title(utils.RandomString(10, false)),
		DateOfBirth: timestamppb.New(time.Unix(rand.Int64N(1000000), 0)),
		Avatar:      0,
		About:       utils.RandomString(300, true),
		Public:      false,
		Email:       utils.RandomString(10, false) + "@hotmail.com",
		Password:    utils.RandomString(10, true),
	}
	return &req
}

func newLoginReq() *users.LoginRequest {
	req := users.LoginRequest{
		Identifier: utils.Title(utils.RandomString(10, false)),
		Password:   utils.Title(utils.RandomString(10, false)),
	}
	return &req
}
