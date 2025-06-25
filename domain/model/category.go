package model

type Category struct {
	Id   Id
	Name Title
}

func NewCategory(id int64, name string) (Category, error) {
	catId, err := NewId(id)
	if err != nil {
		return Category{}, err
	}
	title, err := NewTitle(name)
	if err != nil {
		return Category{}, err
	}
	return Category{Id: catId, Name: title}, nil
}
