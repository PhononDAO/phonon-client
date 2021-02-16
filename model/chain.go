package model

type CurrencyType uint16

var (
	NotSet   CurrencyType = 0x0000
	Bitcoin  CurrencyType = 0x0001
	Ethereum CurrencyType = 0x0002
)
