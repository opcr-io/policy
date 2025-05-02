package clui

// WithTable prints a new table.
func (u *Message) WithTable(headers ...string) *Message {
	u.table.headers = append(u.table.headers, headers)
	u.table.data = append(u.table.data, [][]string{})

	return u
}

// WithTableRow adds a row in the latest table.
func (u *Message) WithTableRow(values ...string) *Message {
	if len(u.table.headers) < 1 {
		return u.WithTable(make([]string, len(values))...).WithTableRow(values...)
	}

	u.table.data[len(u.table.data)-1] = append(u.table.data[len(u.table.data)-1], values)

	return u
}

func (u *Message) WithTableNoAutoWrapText() *Message {
	u.table.noAutoWrap = true
	return u
}
