package config

type Config struct{}

type Selection struct {
	ItemsPath    string `mapstructure:"itemsPath" json:"itemsPath,omitempty"`
	MatchField   string `mapstructure:"matchField" json:"matchField,omitempty"`
	MatchValue   string `mapstructure:"matchValue" json:"matchValue,omitempty"`
	ValuePath    string `mapstructure:"valuePath" json:"valuePath,omitempty"`
	OutputKey    string `mapstructure:"outputKey" json:"outputKey,omitempty"`
	IncludeItem  bool   `mapstructure:"includeItem" json:"includeItem,omitempty"`
	RequireMatch *bool  `mapstructure:"requireMatch" json:"requireMatch,omitempty"`
}

type Inputs struct {
	Input         any        `mapstructure:"input" json:"input"`
	Select        *Selection `mapstructure:"select" json:"select"`
	ParseJSONText *bool      `mapstructure:"parseJSONText" json:"parseJSONText,omitempty"`
}
