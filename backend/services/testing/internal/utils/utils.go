package utils

import (
	"context"
	"fmt"
	"math/rand/v2"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func RandomString(length int, withSpecialChars bool) string {
	runes := make([]rune, length)
	for i := range length {
		runes[i] = RandomRune(withSpecialChars)
	}
	return string(runes)
}

func RandomPassword() string {
	runes := []rune{'!', 'A', 'b', '1'}
	for range 15 {
		runes = append(runes, RandomRune(true))
	}
	return string(runes)
}

var runes = []rune("abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789" + "_-.")
var lettersOnly = []rune("abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandomRune(withSpecialChars bool) rune {
	if withSpecialChars {
		return runes[rand.IntN(len(runes))]
	}
	return lettersOnly[rand.IntN(len(lettersOnly))]
}

func Title(str string) string {
	caser := cases.Title(language.English)
	return caser.String(str)
}

func HandleErr(testTarget string, ctx context.Context, f func(context.Context) error) {
	err := f(ctx)
	if err != nil {
		fmt.Println("------------  FAIL TEST: err ->", testTarget, "-> error:", err.Error())
	} else {
		fmt.Println("------------  SUCCESS -> ", testTarget)
	}
}
