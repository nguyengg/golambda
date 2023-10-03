package timestamp

type AutoGenerateFlag int

const (
	CreatedTimestamp AutoGenerateFlag = 1 << iota
	ModifiedTimestamp
	AllTimestamps = CreatedTimestamp | ModifiedTimestamp
)
