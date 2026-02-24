package scanner

type ScanResult struct {
	OrderedRefs []string
	UniqueRefs  []string
}

func CollectSampleRefs(node any) ScanResult {
	out := ScanResult{
		OrderedRefs: make([]string, 0),
		UniqueRefs:  make([]string, 0),
	}
	seen := map[string]struct{}{}
	collect(node, &out, seen)
	return out
}

func collect(node any, out *ScanResult, seen map[string]struct{}) {
	switch t := node.(type) {
	case map[string]any:
		for key, value := range t {
			if key == "library_sample_path" {
				if s, ok := value.(string); ok && s != "" {
					out.OrderedRefs = append(out.OrderedRefs, s)
					if _, exists := seen[s]; !exists {
						seen[s] = struct{}{}
						out.UniqueRefs = append(out.UniqueRefs, s)
					}
				}
			}
			collect(value, out, seen)
		}
	case []any:
		for _, v := range t {
			collect(v, out, seen)
		}
	}
}

func RewriteSampleRefs(node any, replacement map[string]string) {
	switch t := node.(type) {
	case map[string]any:
		for key, value := range t {
			if key == "library_sample_path" {
				if s, ok := value.(string); ok {
					if updated, exists := replacement[s]; exists {
						t[key] = updated
					}
				}
			}
			RewriteSampleRefs(value, replacement)
		}
	case []any:
		for _, v := range t {
			RewriteSampleRefs(v, replacement)
		}
	}
}
