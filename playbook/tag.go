package playbook

import "strings"

func TagFilter(confTagStr, tagStr string) bool {
	if confTagStr == "" || tagStr == "" {
		return true
	}
	confTags := strings.Split(confTagStr, ",")
	tags := strings.Split(tagStr, ",")
	for _, confTag := range confTags {
		for _, tag := range tags {
			if confTag == tag {
				return true
			}
		}
	}
	return false
}
