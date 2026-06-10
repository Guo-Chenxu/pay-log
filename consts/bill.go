package consts

type Channel int8

const (
	ChannelAlipay Channel = 1
	ChannelWechat Channel = 2
)

func IsValidChannel(v int8) bool {
	switch Channel(v) {
	case ChannelAlipay, ChannelWechat:
		return true
	default:
		return false
	}
}

func ChannelLabel(v int8) string {
	switch Channel(v) {
	case ChannelAlipay:
		return "支付宝"
	case ChannelWechat:
		return "微信"
	default:
		return "未知渠道"
	}
}

type BillType int8

const (
	BillTypeIncome  BillType = 1
	BillTypeExpense BillType = 2
	BillTypeNeutral BillType = 3
)

func IsValidBillType(v int8) bool {
	switch BillType(v) {
	case BillTypeIncome, BillTypeExpense, BillTypeNeutral:
		return true
	default:
		return false
	}
}

func BillTypeLabel(v int8) string {
	switch BillType(v) {
	case BillTypeIncome:
		return "收入"
	case BillTypeExpense:
		return "支出"
	case BillTypeNeutral:
		return "不计收支"
	default:
		return "未知类型"
	}
}
