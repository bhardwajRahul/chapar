package domain

import "github.com/google/uuid"

const EnvKind = "Environment"

type Environment struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	MetaData   MetaData `yaml:"metadata"`
	Spec       EnvSpec  `yaml:"spec"`
	FilePath   string   `yaml:"-"`
}

type EnvSpec struct {
	Values []KeyValue `yaml:"values"`
}

func (e *EnvSpec) Clone() EnvSpec {
	clone := EnvSpec{
		Values: make([]KeyValue, len(e.Values)),
	}

	for i, v := range e.Values {
		clone.Values[i] = v
	}

	return clone
}

func NewEnvironment(name string) *Environment {
	return &Environment{
		ApiVersion: ApiVersion,
		Kind:       EnvKind,
		MetaData: MetaData{
			ID:   uuid.NewString(),
			Name: name,
		},
		Spec: EnvSpec{
			Values: make([]KeyValue, 0),
		},
		FilePath: "",
	}
}

func CompareEnvValue(a, b KeyValue) bool {
	// compare length of the values
	if len(a.Key) != len(b.Key) || len(a.Value) != len(b.Value) || len(a.ID) != len(b.ID) {
		return false
	}

	if a.Key != b.Key || a.Value != b.Value || a.Enable != b.Enable || a.ID != b.ID {
		return false
	}

	return true
}

func (e *Environment) Clone() *Environment {
	clone := &Environment{
		ApiVersion: e.ApiVersion,
		Kind:       e.Kind,
		MetaData:   e.MetaData,
		Spec:       e.Spec.Clone(),
		FilePath:   e.FilePath,
	}

	return clone
}

func CompareEnvValues(a, b []KeyValue) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if !CompareEnvValue(v, b[i]) {
			return false
		}
	}

	return true
}