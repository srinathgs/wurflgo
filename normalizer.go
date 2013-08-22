package wurflgo


type Normalizer interface{
	Normalize(string) string
}

type Null struct{

}

func (null *Null) Normalize(ua string) string {
	return ua
}

type UserAgentNormalizer struct{
	normalizers []Normalizer
}

func NewUserAgentNormalizer(normalizers []Normalizer) *UserAgentNormalizer{
	uaNorm := new(UserAgentNormalizer)
	uaNorm.normalizers = normalizers
	return uaNorm
}

func (UANorm *UserAgentNormalizer) AddNormalizer(norm Normalizer) *UserAgentNormalizer{
	return NewUserAgentNormalizer(append(UANorm.normalizers,norm))
}

func (UANorm *UserAgentNormalizer) Normalize(ua string) string{
	normalizedUA := ua 
	for i := range UANorm.normalizers{
		normalizedUA = UANorm.normalizers[i].Normalize(normalizedUA)
	}
	return normalizedUA
}