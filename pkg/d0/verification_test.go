package d0

import (
	"testing"
)

// test that errors are thrown if a submission is for a blank game
func TestBlankGame(t *testing.T) {
	sig := "gQEBERjDsnVNr4qrYkvaevguF4ypPZHq0yiXfMMKwlu7+kY3HuI8zHx2WhiYj+q26re5uamQ9r8umh54CEJ7zqZAz8IavVblWYznzee9WjIBAB1FeHwILGlKOCDpGBikoZBkMxI4MqjCPzDPAkDMrd1DK0FsWOTpWljLgNGfACTKcgKBAQGPqnGoD6GhuHLYN+Sf73ROColneBdJ7ttuVwm32FvI8LuD5aLDll7bpqfHTWhgbTW02CYvkTAYtoz2RZmIGK5ZHHaM/V6vcSXnq2ab/7mFRiag7D5OUsmIFY9E3IqcqtP7+wXSVgiNFY3DBPy27bXjk8ZJ9nUD5dQBL9sG8TzWd4EBAYrTMfF82EBgsVArIaQjeOuJC3bkPzP5b3El/ZCHkDShpu7wZ82h/82B4W5Ep3KXpgu+YAEULt+5i2WbsfRSXeVZctzD4A++MBqQx9VuN/KsxgHS/20tRiBgd1VElhRD8KJ0lbkxYNcHSkpWSMDFS+eFmizcM3/XQNQ7ukAmM3lkgQEBIZR+FpDFLoGg9mIu2RH9O7lWdifpVhqjrEnvkr4KdB6JzBXAwVPmt1NAVDjGRI/ELlTysOx1b9F2EgdJejY5LgcVxz6irwEckx0z+L10A6Ca2lsGR1E+rViFffNNIJv34dNKgaCInyUNCeBei0AF8KLXLHhRTiBvSVBi6ANb/lY="
	querystring := ""
	postdata := "hello world"

	idfp := "+r+m++fKd1mvYYr5qlZe+FsInDPcj8a2RpwwbII+/20="
	caStatus := true

	result, err := Verify(D0BlindIDKeyGen, D0BlindIDPubKey, sig, querystring, postdata)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if result.IDFP != idfp {
		t.Errorf("Unexpected IDFP value. Expected %s, got %s.", idfp, result.IDFP)
	}

	if result.CAStatus != caStatus {
		t.Errorf("Unexpected CAStatus value. Expected %t, got %t.", caStatus, result.CAStatus)
	}
}
