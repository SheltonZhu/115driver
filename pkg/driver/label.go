package driver

var (
	LabelColors = []string{
		// No Color
		"#000000",
		// Red
		"#FF4B30",
		// Orange
		"#F78C26",
		// Yellow
		"#FFC032",
		// Green
		"#43BA80",
		// Blue
		"#2670FC",
		// Purple
		"#8B69FE",
		// Gray
		"#CCCCCC",
	}

	LabelColorMap = map[string]int{
		"#000000": 0,
		"#FF4B30": 1,
		"#F78C26": 2,
		"#FFC032": 3,
		"#43BA80": 4,
		"#2670FC": 5,
		"#8B69FE": 6,
		"#CCCCCC": 7,
	}
)

type Label struct {
	ID    string
	Name  string
	Color LabelColor
}

type LabelColor int
