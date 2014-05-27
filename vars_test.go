package aetools

import (
	"appengine"
)

// fixture is a sample output that describes the main expected
// serializatoin format when testing.
var fixture = []byte(`[
{
	"__key__": ["Profile", 123456],
	"name": "Ronoaldo JLP",
	"height": 175,
	"active": true,
	"birthday": {
		"type": "date",
		"value": "1986-07-19T00:00:00.000-03:00"
	},
	"description": "This is a long value\nblob string",
	"htmlDesc": {
		"unindexed": true,
		"value": "<h1>This is an awesome, unindexed description"
	},
	"tags": [ "a", "b", "c" ]
}, {
	"__key__": ["IncompleteProfile", "test@example.com"],
	"name": "My Name",
	"height": null
}
]`)

// icon is a sample white png file 16x16,
// dumped as a byte array.
var icon = []byte{
	137, 80, 78, 71, 13, 10, 26, 10,
	0, 0, 0, 13, 73, 72, 68, 82,
	0, 0, 0, 16, 0, 0, 0, 16,
	8, 2, 0, 0, 0, 144, 145, 104,
	54, 0, 0, 0, 9, 112, 72, 89,
	115, 0, 0, 11, 19, 0, 0, 11,
	19, 1, 0, 154, 156, 24, 0, 0,
	0, 7, 116, 73, 77, 69, 7, 222,
	5, 8, 21, 41, 53, 225, 172, 74,
	51, 0, 0, 0, 25, 116, 69, 88,
	116, 67, 111, 109, 109, 101, 110, 116,
	0, 67, 114, 101, 97, 116, 101, 100,
	32, 119, 105, 116, 104, 32, 71, 73,
	77, 80, 87, 129, 14, 23, 0, 0,
	0, 26, 73, 68, 65, 84, 40, 207,
	99, 252, 255, 255, 63, 3, 41, 128,
	137, 129, 68, 48, 170, 97, 84, 195,
	208, 209, 0, 0, 85, 109, 3, 29,
	159, 46, 21, 162, 0, 0, 0, 0,
	73, 69, 78, 68, 174, 66, 96, 130}

// blobKey is a sample appengine.BlobKey value.
var blobKey = appengine.BlobKey("AMIfv94Ly-gFmdjqsU9IwztyA6jjiChzE8cUSwkP8EE" +
	"fo4paIuXmHiwFkoccnayuqcTmkyXfDo8SS9uetO-6h7AhqlKQFYsY1tyGjrhjqmxOYT19CC" +
	"tH5tZEL2pxtCBLe6MFProzW1fw1du_vMwPsypKMHnnpZau6F_qJNoc6yoqnYIKGDvroNk")
