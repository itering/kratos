package config

import (
	"encoding/json"

	"github.com/ghodss/yaml"
	"github.com/pelletier/go-toml"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// ApplyJSON unmarshals a JSON string into a proto message.
// Unknown fields are allowed.
func ApplyJSON(b []byte, v interface{}) error {
	if pb, ok := v.(proto.Message); ok {
		o := protojson.UnmarshalOptions{}
		if err := o.Unmarshal(b, pb); err != nil {
			return err
		}
		return nil
	}
	return json.Unmarshal(b, v)
}

// ApplyYAML unmarshals a YAML string into a proto message.
// Unknown fields are allowed.
func ApplyYAML(b []byte, v interface{}) error {
	b, err := yaml.YAMLToJSON(b)
	if err != nil {
		return err
	}
	return ApplyJSON(b, v)
}

// ApplyTOML unmarshals a TOML string into a proto message.
func ApplyTOML(b []byte, v interface{}) error {
	tree, err := toml.Load(string(b))
	if err != nil {
		return err
	}
	b, err = json.Marshal(tree.ToMap())
	if err != nil {
		return err
	}
	return ApplyJSON(b, v)
}
