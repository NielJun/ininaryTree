package main

import "fmt"

type Data struct {
	string1 string
	string2 string
	string3 string

	int1 int
	int2 int
}

func NewData(str1, str2, str3 string, int1, int2 int) (d *Data) {
	return &Data{
		string1: str1,
		string2: str2,
		string3: str3,
		int1:    int1,
		int2:    int2,
	}
}

// option patten

type Option func(d *Data)

func initData(opts ...Option) (d *Data) {

	d = & Data{}
	for _, opt := range opts {
		opt(d)
	}
	return
}

func insertStr1(str1 string) Option {
	return func(d *Data) {
		d.string1 = str1
	}
}

func insertStr2(str2 string) Option {
	return func(d *Data) {
		d.string2 = str2
	}
}

func main() {
	d := initData(insertStr1("我爱你"), insertStr2("中国"))
	fmt.Printf("str1 is %s,str2 is %s", d.string1, d.string2)
}
