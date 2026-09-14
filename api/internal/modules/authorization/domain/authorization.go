package domain

type Context struct {
	Subject         string
	ApplicationCode string
	ManifestVersion int64
	Permissions     map[string]struct{}
}

func (c Context) HasPermission(code string) bool {
	_, ok := c.Permissions[code]
	return ok
}
