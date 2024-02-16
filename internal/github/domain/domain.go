package domain

type SortBy string

const (
	SortByCreated  SortBy = "created"
	SortByUpdated  SortBy = "updated"
	SortByPushed   SortBy = "pushed"
	SortByFullNmae SortBy = "full_name"
)

func (s SortBy) String() string {
	return string(s)
}
