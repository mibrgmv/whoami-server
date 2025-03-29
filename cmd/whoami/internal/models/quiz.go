package models

import pb "whoami-server/protogen/golang/quiz"

type Quiz struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Results []string `json:"results"`
}

func (q Quiz) ToProto() *pb.Quiz {
	return &pb.Quiz{
		Id:      q.ID,
		Title:   q.Title,
		Results: q.Results,
	}
}
