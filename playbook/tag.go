package playbook

func TagFilter(confTags, tags []string) bool {
	if len(confTags) == 0 {
		return true
	}
	for _, confTag := range confTags {
		for _, tag := range tags {
			if confTag == tag {
				return true
			}
		}
	}
	return false
}
