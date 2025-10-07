package mjsc

const (
	HuTypePingHu      = iota //平胡 0
	HuTypeKaDang             //卡当 1
	HuTypeZiMo               //自摸 2
	HuTypeMoBao              //摸宝 3
	HuTypeLouBao             //搂宝 4
	HuTypeGuaDaFeng          //刮大风 5
	HuTypeBaoZhongBao        //宝中宝 6
)

var multiples = map[int32]int64{HuTypePingHu: 1, HuTypeKaDang: 2, HuTypeZiMo: 2, HuTypeMoBao: 3, HuTypeLouBao: 3, HuTypeGuaDaFeng: 6, HuTypeBaoZhongBao: 12}

func totalMuti(huTypes []int32) int64 {
	totalMuti := int64(1)
	for _, huType := range huTypes {
		if multiple, ok := multiples[huType]; ok {
			totalMuti *= multiple
		}
	}
	return totalMuti
}
