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

	dist := seedfake.NewMixedDistribution(rand.New(source), []seedfake.NumberDistribution{seedfake.Min{}, seedfake.Max{}, flat}, []float64{1, 1, 3})
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
	// Min: map[seed.CodeName]interface {}{"bool":false, "bytes_10":[]uint8{}, "datetime_sec_9999":time.Date(0, time.December, 31, 11, 1, 0, 0, time.Location("-12:59")), "integer_js":-9007199254740991, "text_10":""}
	// error 2 <nil>
	// Max: map[seed.CodeName]interface {}{"bool":true, "bytes_10":[]uint8{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, "datetime_sec_9999":time.Date(10000, time.January, 1, 12, 59, 59, 999999000, time.Location("+13:00")), "integer_js":9007199254740991, "text_10":"~~~~~~~~~~"}
	// error 3 <nil>
	// Flat: map[seed.CodeName]interface {}{"bool":false, "bytes_10":[]uint8{0xc3, 0x54, 0xf3, 0xed, 0xe2, 0xd6, 0xbe}, "datetime_sec_9999":time.Date(5988, time.August, 19, 16, 30, 40, 837874000, time.Location("-9:10")), "integer_js":-7185322384382250, "text_10":"(9;#x[1X"}
	// error 4 <nil>
	// map[seed.CodeName]interface {}{"bool":false, "bytes_10":[]uint8{0xd4, 0xa6, 0x53, 0x14, 0x76, 0x8d, 0x68, 0x91}, "datetime_sec_9999":time.Date(5630, time.February, 27, 22, 5, 38, 592223000, time.Location("+0:07")), "integer_js":-6466883365446492, "text_10":"~  ~)iM L{"}
	// map[seed.CodeName]interface {}{"bool":true, "bytes_10":[]uint8{0x9d, 0xd2, 0xce, 0xb4, 0x9c, 0x24, 0x74, 0x30, 0x3c, 0xbb}, "datetime_sec_9999":time.Date(0, time.December, 31, 20, 57, 0, 0, time.Location("-3:03")), "integer_js":-9007199254740991, "text_10":"82sO~U~ Bt"}
	// map[seed.CodeName]interface {}{"bool":false, "bytes_10":[]uint8{}, "datetime_sec_9999":time.Date(8837, time.August, 18, 10, 16, 6, 337786000, time.Location("-1:35")), "integer_js":-888819755794019, "text_10":"Z2 y ao~G"}
	// map[seed.CodeName]interface {}{"bool":false, "bytes_10":[]uint8{0x6d, 0xec, 0x71, 0xcb, 0x7a, 0x4b, 0xca, 0xed, 0x91, 0x30}, "datetime_sec_9999":time.Date(7117, time.April, 30, 11, 35, 20, 271231000, time.Location("+2:38")), "integer_js":9007199254740991, "text_10":".; "}
	// map[seed.CodeName]interface {}{"bool":true, "bytes_10":[]uint8{0x7e, 0xc8, 0x69, 0xa7, 0x57, 0xb7, 0xc4}, "datetime_sec_9999":time.Date(0, time.December, 31, 11, 38, 0, 0, time.Location("-12:22")), "integer_js":9007199254740991, "text_10":"q WH\\>~6Ai"}
}
