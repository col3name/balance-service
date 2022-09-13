package currency

import (
	"fmt"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	tNow := time.Now()

	//time.Time to Unix Timestamp
	tUnix := tNow.Unix()
	fmt.Printf("timeUnix %d\n", tUnix)

	//Unix Timestamp to time.Time
	timeT := tNow.String()
	parse, err := time.Parse(timeT, timeT)
	fmt.Println(err)

	fmt.Printf("time.Time: %s\n", timeT)
	fmt.Println(parse.String())
}
