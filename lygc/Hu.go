package lygc

type EHuType int

const (
	HuTypeZiMo   EHuType = 1
	HuTypeMingCi EHuType = 2
	HuTypeAnCi   EHuType = 3
	HuTypePiCi   EHuType = 4
	HuTypeBaoCi  EHuType = 5
	HuTypeEnd    EHuType = 6
	HuTypeNone   EHuType = -1
)

type EFanType int

const (
	FanTypeQiDui      EFanType = 1
	FanTypePengPengHu EFanType = 2
	FanTypeQingYiSe   EFanType = 3
	FanTypeGangShang  EFanType = 4
)
