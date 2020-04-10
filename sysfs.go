package gpio

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type direction uint

const (
	inDirection direction = iota
	outDirection
)

type Edge uint

const (
	EdgeNone Edge = iota
	EdgeRising
	EdgeFalling
	EdgeBoth
)

type LogicLevel uint

const (
	ActiveHigh LogicLevel = iota
	ActiveLow
)

type Value uint

const (
	Inactive Value = 0
	Active   Value = 1
)

func exportGPIO(p Pin) error {
	export, err := os.OpenFile("/sys/class/gpio/export", os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open gpio export file for writing\n. Error: %s", err.Error())
	}
	defer export.Close()

	_, err := export.Write([]byte(strconv.Itoa(int(p.Number))))
	if err != nil {
		return fmt.Errorf("Unable to export gpio, write failed\n. Error: %s", err.Error())
	}

	return nil
}

func unexportGPIO(p Pin) error {
	export, err := os.OpenFile("/sys/class/gpio/unexport", os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open gpio unexport file for writing\n. Error: %s", err.Error())
	}
	defer export.Close()

	_, err := export.Write([]byte(strconv.Itoa(int(p.Number))))
	if err != nil {
		return fmt.Errorf("Unable to export gpio, write failed\n. Error: %s", err.Error())
	}

	return nil
}

func setDirection(p Pin, d direction, initialValue uint) error {
	dir, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/direction", p.Number), os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("failed to open gpio %d direction file for writing\n", p.Number)
		os.Exit(1)
	}
	defer dir.Close()

	var err Error
	switch {
	case d == inDirection:
		_, err = dir.Write([]byte("in"))
	case d == outDirection && initialValue == 0:
		_, err = dir.Write([]byte("low"))
	case d == outDirection && initialValue == 1:
		_, err = dir.Write([]byte("high"))
	default:
		err = fmt.Errorf("setDirection called with invalid direction or initialValue, %d, %d", d, initialValue)
	}

	return err
}

func setEdgeTrigger(p Pin, e Edge) {
	edge, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/edge", p.Number), os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("failed to open gpio %d edge file for writing\n", p.Number)
		os.Exit(1)
	}
	defer edge.Close()

	var err error

	switch e {
	case EdgeNone:
		_, err = edge.Write([]byte("none"))
	case EdgeRising:
		_, err = edge.Write([]byte("rising"))
	case EdgeFalling:
		_, err = edge.Write([]byte("falling"))
	case EdgeBoth:
		_, err = edge.Write([]byte("both"))
	default:
		err = fmt.Errorf("setEdgeTrigger called with invalid edge %d", e)
	}

	return err
}

func setLogicLevel(p Pin, l LogicLevel) error {
	level, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/active_low", p.Number), os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer level.Close()

	var err error

	switch l {
	case ActiveHigh:
		_, err = level.Write([]byte("0"))
	case ActiveLow:
		_, err = level.Write([]byte("1"))
	default:
		err = errors.New("Invalid logic level setting.")
	}

	return err
}

func openPin(p Pin, write bool) (Pin, error) {
	flags := os.O_RDONLY
	if write {
		flags = os.O_RDWR
	}
	f, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", p.Number), flags, 0600)
	if err != nil {
		return p, fmt.Errorf("failed to open gpio %d value file for reading\n. Error: %s", p.Number, err.Error())
	}
	p.f = f
	return p, nil
}

func readPin(p Pin) (val uint, err error) {
	file := p.f
	file.Seek(0, 0)
	buf := make([]byte, 1)
	_, err = file.Read(buf)
	if err != nil {
		return 0, err
	}
	c := buf[0]
	switch c {
	case '0':
		return 0, nil
	case '1':
		return 1, nil
	default:
		return 0, fmt.Errorf("read inconsistent value in pinfile, %c", c)
	}
}

func writePin(p Pin, v uint) error {
	var buf []byte
	switch v {
	case 0:
		buf = []byte{'0'}
	case 1:
		buf = []byte{'1'}
	default:
		return fmt.Errorf("invalid output value %d", v)
	}
	_, err := p.f.Write(buf)
	return err
}
