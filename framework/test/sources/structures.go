package sources

import (
	"fmt"
	"log"
	"math/rand"
	"time"
)

type MyType struct {
	Num int
	Str string
}

func createMyTypeWithRandom() (MyType, error) {
	randInt := rand.Intn(100)
	if randInt > 50 {
		return MyType{}, fmt.Errorf("panic")
	}
	return MyType{
		Num: randInt,
		Str: "Test",
	}, nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	a, err := createMyTypeWithRandom()
	if err != nil {
		log.Println(err)
	} else {
		log.Println(a.Num)
	}
}
