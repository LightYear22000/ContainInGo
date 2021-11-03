package utils

/*
	This is the format of our imageDB file where we store the
	list of images we have on the system.
	{
		"ubuntu" : {
						"18.04": "[image-hash]",
						"18.10": "[image-hash]",
						"19.04": "[image-hash]",
						"19.10": "[image-hash]",
					},
		"centos" : {
						"6.0": "[image-hash]",
						"6.1": "[image-hash]",
						"6.2": "[image-hash]",
						"7.0": "[image-hash]",
					}
	}
*/

type (
	ImageEntries map[string]string
	ImagesDB map[string]ImageEntries
	Manifest []struct {
		Config string
		RepoTags []string
		Layers []string
	}
	ImageConfigDetails struct {
		Env []string	`json:"Env"`
		Cmd []string	`json:"Cmd"`
	}
	ImageConfig struct {
		Config ImageConfigDetails `json:"Config"`
	}
)