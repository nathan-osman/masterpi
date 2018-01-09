## masterpi

This simple Go app (compiled for ARMv6) performs the following functions:

- monitor w1 temperature sensors
- display clock and temperature on OLED
- operate a mechanical button for toggling a lamp
- provide an HTTP server for controlling a relay
- upload temperature readings to InfluxDB

As you can imagine, this app is written specifically for my hardware setup. Therefore, you may need to adapt the code to make it work for you.
