package seedfake_test

import (
	"fmt"
	"math/rand"

	"github.com/xiegeo/seed/demo/testdomain"
	"github.com/xiegeo/seed/seedfake"
)

func ExampleValueGen() {
	obj := testdomain.ObjLevel0()

	values, err := seedfake.NewValueGen(seedfake.Min{}).ValuesForObject(obj, 1)
	fmt.Println("error 1", err)
	fmt.Printf("Min: %#v\n", values[0])

	values, err = seedfake.NewValueGen(seedfake.Max{}).ValuesForObject(obj, 1)
	fmt.Println("error 2", err)
	fmt.Printf("Max: %#v\n", values[0])

	source := rand.NewSource(0)
	flat := seedfake.NewFlat(source)
	values, err = seedfake.NewValueGen(flat).ValuesForObject(obj, 1)
	fmt.Println("error 3", err)
	fmt.Printf("Flat: %#v\n", values[0])

	dist := seedfake.NewMixedDistribution(rand.New(source), []seedfake.NumberDistribution{seedfake.Min{}, seedfake.Max{}, flat}, []float64{1, 1, 4})
	gen := seedfake.NewValueGen(dist)
	values, err = gen.ValuesForObject(obj, 5)
	fmt.Println("error 4", err)
	fmt.Printf("%#v\n", values[0])
	fmt.Printf("%#v\n", values[1])
	fmt.Printf("%#v\n", values[2])
	fmt.Printf("%#v\n", values[3])
	fmt.Printf("%#v\n", values[4])

	// Output:
	// error 1 <nil>
	// Min: map[seed.CodeName]interface {}{"bool":false, "bytes_10":[]uint8{}, "datetime_sec_9999":time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), "integer_js":-9007199254740991, "text_10":""}
	// error 2 <nil>
	// Max: map[seed.CodeName]interface {}{"bool":true, "bytes_10":[]uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "datetime_sec_9999":time.Date(9999, time.December, 31, 23, 59, 59, 999999000, time.UTC), "integer_js":9007199254740991, "text_10":"~~~~~~~~~~"}
	// error 3 <nil>
	// Flat: map[seed.CodeName]interface {}{"bool":false, "bytes_10":[]uint8{0xc3, 0x54, 0xf3, 0xed, 0xe2, 0xd6, 0xbe}, "datetime_sec_9999":time.Date(5988, time.August, 20, 1, 40, 40, 837874000, time.UTC), "integer_js":-4610345107248525, "text_10":"(9;#x[1X"}
	// error 4 <nil>
	// map[seed.CodeName]interface {}{"bool":true, "bytes_10":[]uint8{0xd4, 0xa6, 0x53, 0x14, 0x76, 0x8d, 0xb7, 0xd9}, "datetime_sec_9999":time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), "integer_js":547436066665727, "text_10":"JT~i`Gr'VO"}
	// map[seed.CodeName]interface {}{"bool":false, "bytes_10":[]uint8{0x84, 0x6b}, "datetime_sec_9999":time.Date(2823, time.January, 31, 10, 2, 28, 34392000, time.UTC), "integer_js":3044997194790778, "text_10":":~82"}
	// map[seed.CodeName]interface {}{"bool":true, "bytes_10":[]uint8{0x60, 0x23, 0x66}, "datetime_sec_9999":time.Date(9999, time.December, 31, 23, 59, 59, 999999000, time.UTC), "integer_js":9007199254740991, "text_10":"DBtm  ;"}
	// map[seed.CodeName]interface {}{"bool":false, "bytes_10":[]uint8{0x70, 0xba, 0xbc, 0xe3, 0xe, 0x54, 0xeb, 0xc0, 0x7f, 0xa4}, "datetime_sec_9999":time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), "integer_js":-5731509364127879, "text_10":"4~ e"}
	// map[seed.CodeName]interface {}{"bool":false, "bytes_10":[]uint8{0x91, 0xf, 0x9d, 0xdb, 0x9, 0xa0, 0xd0, 0x59, 0xc4, 0xcd}, "datetime_sec_9999":time.Date(2923, time.April, 30, 21, 32, 36, 405599000, time.UTC), "integer_js":9007199254740991, "text_10":"2K."}
}
