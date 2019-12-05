package common

func Join(a interface{}, step string) string {
	s := a.([]interface{})
	res := ""
	for i, item := range s {
		res += item.(string)
		if i < len(s)-1 {
			res += ","
		}
	}
	return res
}
