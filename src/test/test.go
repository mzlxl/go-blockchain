package test

import (
	"fmt"
	"strconv"
	"time"
)

func main() {
	fmt.Println("test")

	timestamp := time.Now().Unix()
	fmt.Println(timestamp)
	fmt.Println(strconv.FormatInt(timestamp, 10))

}
