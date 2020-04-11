package fail

import "log"

func Check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
