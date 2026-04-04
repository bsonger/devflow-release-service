package model

type Configuration struct {
	BaseModel `bson:",inline" json:",inline"`
	Name      string  `bson:"name" json:"name"`
	Files     []*File `bson:"files,omitempty" json:"files,omitempty"`
}

type File struct {
	Name    string `bson:"name" json:"name"`
	Content string `bson:"content" json:"content"`
}

func (Configuration) CollectionName() string { return "configuration" }
