package cmd

// Eval args
func Eval(args map[string]interface{}) error {
	if args["config"].(bool) {
		return Config(args)
	} else if args["submit"].(bool) {
		return Submit(args)
	} else if args["list"].(bool) {
		return List(args)
	} else if args["parse"].(bool) {
		return Parse(args)
	} else if args["gen"].(bool) {
		return Gen(args)
	} else if args["test"].(bool) {
		return Test(args)
	}
	return nil
}
