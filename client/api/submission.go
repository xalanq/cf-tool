package api

type Submission struct {
	ContestID    float64
	SubmissionID float64
	ProblemID    string
	Verdict      string
	Lang         string
	Timestamp    float64
}

func NewSubmission(contestID, submissionID float64, problemID, verdict, lang string, timestamp float64) *Submission {
	return &Submission{
		ContestID:    contestID,
		SubmissionID: submissionID,
		ProblemID:    problemID,
		Verdict:      verdict,
		Lang:         lang,
		Timestamp:    timestamp,
	}
}
