package models

type Description struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Reboot      bool     `json:"reboot"`
	Files       []File   `json:"files"`
	Scripts     []Script `json:"scripts"`
}

type File struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Md5      string `json:"md5"`
	Sha256   string `json:"sha256"`
}

type Script struct {
	Filename string `json:"filename"`
	Type     string `json:"type"`
	Md5      string `json:"md5"`
	Sha256   string `json:"sha256"`
}
