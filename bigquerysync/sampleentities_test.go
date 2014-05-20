package bigquerysync_test

var SampleEntities = `[
	{
		"__key__": ["Sample", 1],
		"Name": "Sample #1",
		"Order": 1
	},
	{
		"__key__": ["Sample", 2],
		"Name": "Sample #2",
		"Order": 2
	},
	{
		"__key__": ["Sample", 3],
		"Name": "Sample #3",
		"Order": 3
	},
	{
		"__key__": ["Log", "log-entry-1"],
		"Level": "INFO",
		"Message": "Sample log message #1"
	},
	{
		"__key__": ["Log", "log-entry-2"],
		"Level": "WARN",
		"Message": "Sample log message #2"
	}
]`
