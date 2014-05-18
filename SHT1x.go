// SHT1x project SHT1x.go
// re implementation in go of https://bitbucket.org/lunobili/rpisht1x by Luca Nobili

// This package reads Humidity and Temperature from a Sensirion SHT1x and SHT7x sensors. It has been tested
// with an SHT71.

// It is meant to be used in a Raspberry Pi.

// Requires root privileges, therefore, to run this module you need to run your script as root.

// example usage:
//		 sht := SHT1x.New(rpi.GPIO_P1_11, rpi.GPIO_P1_07)

//		 temp := sht.ReadTemperature()
//		 fmt.Printf("temp (°C): %.2f\n", temp)

//		 humid := sht.ReadHumidity()
//		 fmt.Printf("Humid (rel%%): %.2f\n", humid)

//		 temp2, humid2 := sht.ReadTempAndHumidity()

//		 fmt.Printf("temp (°C): %.2f\n", temp2)
//		 fmt.Printf("Humid (rel%%): %.2f\n", humid2)

//		 sht.CleanUp()

package SHT1x

import (
	"log"
	"time"

	"github.com/griffina/gpio"
	_ "github.com/griffina/gpio/rpi"
)

// Sensor type called SHT1x after the Sht1x class in rpisht1x.
// Can be used with SHT71/SHT75/SHT15/SHT11/SHT10
type SHT1x struct {
	dataPin  gpio.Pin
	clockPin gpio.Pin
}

// Consts from the datasheet for comunicating withe the sensor
// These numbers are from the SHT7x datasheet
// http://www.sensirion.com/fileadmin/user_upload/customers/sensirion/Dokumente/Humidity/Sensirion_Humidity_SHT7x_Datasheet_V5.pdf
// If using SHT1x the refer to:
// http://www.sensirion.com/fileadmin/user_upload/customers/sensirion/Dokumente/Humidity/Sensirion_Humidity_SHT1x_Datasheet_V5.pdf
const (
	d1 float32 = -40.1
	d2 float32 = 0.01

	c1 float32 = -2.0468    // for 12 Bit
	c2 float32 = 0.0367     // for 12 Bit
	c3 float32 = -1.5955E-6 // for 12 Bit
	t1 float32 = 0.01       // for 12 Bit @ 5V
	t2 float32 = 0.00008    // for 12 Bit @ 5V

	//bin for "00000101"
	humidCmd uint8 = 5
	//bin for "00000011"
	tempCmd uint8 = 3
)

// Create a new sensor, supplying the clock and data pins,
// returns a pointer to a sensor
func New(P1_dataPin, P1_clockPin int) *SHT1x {
	//create two gpio pins

	pinData, pin1_err := gpio.OpenPin(P1_dataPin, gpio.ModeOutput)

	if pin1_err != nil {
		log.Println("error opening data pin:", P1_dataPin, pin1_err)
	}

	pinClock, pin2_err := gpio.OpenPin(P1_clockPin, gpio.ModeOutput)

	if pin2_err != nil {
		log.Println("error opening clock pin:", P1_clockPin, pin2_err)
	}

	return &SHT1x{dataPin: pinData, clockPin: pinClock}
}

// Reads th humidity from the sensor and returns the relative humidity %
// to do this the temperature also has to be read
func (sht *SHT1x) ReadHumidity() float32 {
	// not interested in the temp returned, but the
	// temp is needed to read the relative humidity
	_, humidity := sht.ReadTempAndHumidity()
	return humidity
}

// Read the temperature from the sensor and returns °C
// Like rpisht1x,:
//  "I deliberately will not implement read_temperature_F because I believe in the
//   in the Metric System (http://en.wikipedia.org/wiki/Metric_system)"
func (sht *SHT1x) ReadTemperature() float32 {

	sht.sendCommand(tempCmd)
	sht.waitForResult()
	val := sht.getData16()
	sht.skipCRC()
	// Maths from data sheet
	return (float32(val) * d2) + d1
}

// Read the temperature in °C and relative humidity from the sensor and returns
func (sht *SHT1x) ReadTempAndHumidity() (temp, humidity float32) {
	temp = sht.ReadTemperature()

	sht.sendCommand(humidCmd)
	sht.waitForResult()
	val := sht.getData16()

	sht.skipCRC()

	floatVal := float32(val)
	// Maths from data sheet
	linearHumidity := c1 + c2*floatVal + c3*floatVal*floatVal

	humidity = (temp-25.0)*(t1+t2*floatVal) + linearHumidity
	return temp, humidity
}

//Reset the sensor
func (sht *SHT1x) Reset() {
	sht.dataPin.SetMode(gpio.ModeOutput)
	sht.clockPin.SetMode(gpio.ModeOutput)
	sht.dataPin.Set()

	for i := 0; i < 10; i++ {
		sht.clockTick(true)
		sht.clockTick(false)
	}
}

// Set the gpio pins back to input for safety
func (sht *SHT1x) CleanUp() {
	sht.dataPin.SetMode(gpio.ModeInput)
	sht.clockPin.SetMode(gpio.ModeInput)
}

///// Private methods below

func (sht *SHT1x) shiftIn(numberofBits int16) uint16 {
	var ret uint16
	var i int16
	for i = 0; i < numberofBits; i++ {

		sht.clockTick(true)

		binVal := sht.dataPin.Get()

		if binVal == true {
			ret = (ret * 2) + 1
		} else {
			ret = ret * 2
		}
		sht.clockTick(false)
	}

	return ret
}

// Send the a command to the sensor and process the ACK
func (sht *SHT1x) sendCommand(command uint8) {

	sht.dataPin.SetMode(gpio.ModeOutput)
	sht.clockPin.SetMode(gpio.ModeOutput)

	sht.dataPin.Set()
	sht.clockTick(true)
	sht.dataPin.Clear()
	sht.clockTick(false)
	sht.clockTick(true)
	sht.dataPin.Set()
	sht.clockTick(false)

	var i uint8
	for i = 0; i < 8; i++ {
		var bitVal uint8
		bitVal = command & (1 << (7 - i))
		if bitVal != 0 {
			sht.dataPin.Set()
		} else {
			sht.dataPin.Clear()
		}
		sht.clockTick(true)
		sht.clockTick(false)
	}

	sht.clockTick(true)

	sht.dataPin.SetMode(gpio.ModeInput)

	ack := sht.dataPin.Get()

	if ack != false {
		log.Println("Nack 1 false, in sent command")
	}

	sht.clockTick(false)

	ack = sht.dataPin.Get()

	if ack != true {
		log.Println("Nack 2 true, in sent command")
	}

}

// if High == true set the clock line high
// else set it low
// then wait 100 nanoseconds
func (sht *SHT1x) clockTick(high bool) {
	if high {
		sht.clockPin.Set()
	} else {
		sht.clockPin.Clear()
	}
	time.Sleep(100 * time.Nanosecond)
}

// wait for the data bin to become high to signal the data is ready
func (sht *SHT1x) waitForResult() {
	var i int16
	var ack bool
	sht.dataPin.SetMode(gpio.ModeInput)

	for i = 0; i < 100; i++ {
		time.Sleep(10 * time.Millisecond)
		ack = sht.dataPin.Get()

		if ack == false {
			return
		}
	}
	log.Println("Wait exhausted")
}

// get the data from the pins
func (sht *SHT1x) getData16() uint16 {
	var val uint16
	//// Get the most significant bits
	sht.dataPin.SetMode(gpio.ModeInput)
	sht.clockPin.SetMode(gpio.ModeOutput)

	val = sht.shiftIn(8)
	val *= 256

	//// Send the required ack
	sht.dataPin.SetMode(gpio.ModeOutput)
	sht.dataPin.Set()
	sht.dataPin.Clear()

	sht.clockTick(true)
	sht.clockTick(false)

	//// Get the least significant bits
	sht.dataPin.SetMode(gpio.ModeInput)
	val |= sht.shiftIn(8)

	return val
}

// Ignore the CRC for now
func (sht *SHT1x) skipCRC() {
	sht.dataPin.SetMode(gpio.ModeOutput)
	sht.clockPin.SetMode(gpio.ModeOutput)

	sht.dataPin.Set()
	sht.clockTick(true)
	sht.clockTick(false)
}
