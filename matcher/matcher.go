package matcher


type Matcher interface{
	Match([]string, string, int) string
}