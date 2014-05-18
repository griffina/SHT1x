# SHT1x

SHT1x is a Golang implementation of Luca Nobili's [rpiSht1x](https://bitbucket.org/lunobili/rpisht1x)

For using the Sensirion SHT (SHT71/SHT75/SHT15/SHT11/SHT10) series of temperature and humidity sensors on a Raspberry Pi in Go.

Using a modified version of [gpio for Go](https://github.com/griffina/gpio)

## Full example
This needs to be run as ```root```

``` go
package main

import "fmt"
import "github.com/griffina/SHT1x"
import "github.com/griffina/gpio/rpi"

func main() {

	fmt.Println("Create SHT sensor")
	sht := SHT1x.New(rpi.GPIO_P1_11, rpi.GPIO_P1_07)

	fmt.Println("read temp")
	temp := sht.ReadTemperature()
	fmt.Printf("temp (°C): %.2f\n", temp)

	humid := sht.ReadHumidity()
	fmt.Printf("Humid (rel%%): %.2f\n", humid)

	temp2, humid2 := sht.ReadTempAndHumidity()

	fmt.Printf("temp (°C): %.2f\n", temp2)
	fmt.Printf("Humid (rel%%): %.2f\n", humid2)

	sht.CleanUp()

}

```

To install the SHT1x package:
~~~
go get github.com/griffia/SHT1x
~~~

To build Go on the Raspberry Pi [follow these instructions](http://www.maketecheasier.com/build-go-from-source-on-raspberry-pi/)

# [Documentation](http://godoc.org/github.com/griffina/SHT1x)

[SHT1x Datasheet](http://www.sensirion.com/fileadmin/user_upload/customers/sensirion/Dokumente/Humidity/Sensirion_Humidity_SHT7x_Datasheet_V5.pdf)

## TODO:
* Calculate the CRC
* Return errors on failures
* Some tests


