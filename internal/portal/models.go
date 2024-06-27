package portal

type StudentResult struct {
	ScopeName string `json:"ScopeName"`
	Year      string `json:"Year"`
	Ds        []struct {
		GradeName        *string `json:"GradeName"`
		Percent          *string `json:"Percent"`
		Total            *string `json:"Total"`
		StudyYearCourses []struct {
			GradeName string `json:"GradeName"`
			Parts     []struct {
				DegreesType    []string `json:"DegreesType"`
				DegreesMax     string   `json:"DegreesMax"`
				Degrees        []string `json:"Degrees"`
				CoursePartName string   `json:"CoursePartName"`
				SemasterName   string   `json:"SemasterName"`
			} `json:"Parts"`
			CourseName  string `json:"CourseName"`
			SuccessFlag string `json:"SuccessFlag"`
			Max         string `json:"Max"`
			Total       string `json:"Total"`
		} `json:"StudyYearCourses"`
	} `json:"ds"`
}

type StudentData struct {
	CollageID   string `json:"CollageID"`
	ImagePath   string `json:"ImagePath"`
	UUID        string `json:"UUID"`
	Collage     string `json:"Collage"`
	ScopeUUID   string `json:"ScopeUUID"`
	StdName     string `json:"StdName"`
	Year        string `json:"Year"`
	ShowMessage string `json:"ShowMessage"`
	ID          int64  `json:"ID"`
	StudyYear   string `json:"StudyYear"`
}
