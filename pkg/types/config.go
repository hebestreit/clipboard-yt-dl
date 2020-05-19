package types

type Config struct {
	Profile map[string]Profile
	Default Default
}

type Default struct {
	Profile       string
}

type Profile struct {
	Title string
	Args  []string
}
