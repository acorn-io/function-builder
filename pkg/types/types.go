package types

import (
	"encoding/json"

	"github.com/google/shlex"
)

type Template struct {
	Name       string            `json:"name,omitempty"`
	Targets    map[string]Build  `json:"targets,omitempty"`
	LocalFiles map[string]string `json:"localFiles,omitempty"`
	Detect     []DetectFile      `json:"detect,omitempty"`
}

type DetectFile struct {
	Priority *int     `json:"priority,omitempty"`
	Files    []string `json:"files,omitempty"`
}

type Shlex []string

func (s *Shlex) UnmarshalJSON(data []byte) error {
	if data[0] == '[' {
		type shlex Shlex
		return json.Unmarshal(data, (*shlex)(s))
	}
	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	parts, err := shlex.Split(v)
	if err != nil {
		return err
	}
	*s = parts
	return nil
}

type Build struct {
	From         string       `json:"from,omitempty"`
	Shell        Shlex        `json:"shell,omitempty"`
	StopSignal   string       `json:"stopSignal,omitempty"`
	CacheVolumes []string     `json:"cacheVolumes,omitempty"`
	Entrypoint   Shlex        `json:"entrypoint,omitempty"`
	Workdir      string       `json:"workdir,omitempty"`
	Acornfile    string       `json:"acornfile,omitempty"`
	User         *int         `json:"user,omitempty"`
	Run          Instructions `json:"run,omitempty"`
}

type Instructions []Instruction

func (i *Instructions) UnmarshalJSON(data []byte) error {
	if data[0] == '[' {
		return json.Unmarshal(data, (*[]Instruction)(i))
	}

	var v string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*i = []Instruction{
		{
			Run: Run{
				Command: v,
			},
		},
	}
	return nil
}

type Run struct {
	Command string `json:"command,omitempty"`
	Mount   Mounts `json:"mount,omitempty"`
}

type Mounts []Mount

func (m *Mounts) UnmarshalJSON(data []byte) error {
	if data[0] == '[' {
		return json.Unmarshal(data, (*[]Mount)(m))
	}

	var mnt Mount
	if err := json.Unmarshal(data, &mnt); err != nil {
		return err
	}
	*m = []Mount{
		mnt,
	}
	return nil
}

type Copy struct {
	Source string `json:"source,omitempty"`
	Dest   string `json:"dest,omitempty"`
	From   string `json:"from,omitempty"`
	Link   bool   `json:"link,omitempty"`
}

type Volume struct {
	Volume string `json:"volume,omitempty"`
}

type Workdir struct {
	Workdir string `json:"workdir,omitempty"`
}

type Instruction struct {
	Run     `json:",inline"`
	Volume  `json:",inline"`
	Workdir `json:",inline"`
	Copy    Copy `json:"copy,omitempty"`
}

func (i *Instruction) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		var v string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*i = Instruction{
			Run: Run{
				Command: v,
			},
		}
		return nil
	}

	type instruction Instruction
	return json.Unmarshal(data, (*instruction)(i))
}

type CacheMount struct {
	Target string `json:"target,omitempty"`
}

func (c *CacheMount) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		var v string
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*c = CacheMount{
			Target: v,
		}
		return nil
	}

	type cacheMount CacheMount
	return json.Unmarshal(data, (*cacheMount)(c))
}

type Mount struct {
	Cache *CacheMount `json:"cache,omitempty"`
}
