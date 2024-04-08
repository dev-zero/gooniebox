package main

import (
    "log"
    "fmt"
    "strings"
    "time"
    "encoding/hex"

    "github.com/holoplot/go-evdev"

    "periph.io/x/host/v3"
	// "periph.io/x/conn/v3/physic"
	// "periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/spi/spireg"
    "periph.io/x/devices/v3/mfrc522"
    "periph.io/x/host/v3/rpi"
)


type InputEvent struct {
    evdevEvent *evdev.InputEvent
    deviceIdentifier string
}


func readEvents(devicePath string, ident string, events chan<- InputEvent) {
    device, err := evdev.Open(devicePath)
    if err != nil {
        log.Printf("Cannot read %s: %v\n", devicePath, err)
        return
    }

    defer device.Close()

    for {
        event, err := device.ReadOne()
        if err != nil {
            log.Printf("Error reading from device: %v\n", err)
            return
        }

        events <- InputEvent{event, ident}
    }
}

func doInputProc() error {
    // EVDEV
    devicePaths, err := evdev.ListDevicePaths()
    if err != nil {
        log.Printf("Cannot list device paths: %s", err)
        return err
    }

    for _, d := range devicePaths {
        log.Printf("%s:\t%s\n", d.Path, d.Name)
    }

    // SPI
    state, err := host.Init()
    if err != nil {
		log.Fatal(err)
	}

    log.Printf("Loaded periph.io drivers:\n")
    for _, impl := range state.Loaded {
        log.Printf("  - %s\n", impl)
    }
    log.Printf("Skipped periph.io drivers:\n")
    for _, fails := range state.Skipped {
        log.Printf("  - %s\n", fails)
    }
    log.Printf("Failed periph.io drivers:\n")
    for _, fails := range state.Failed {
        log.Printf("  - %s\n", fails)
    }

    spidevs := spireg.All()
    log.Printf("Available SPI devices:\n")
    for _, spidevref := range spidevs {
        log.Printf("%s:\t%d, aka: %s\n", spidevref.Name, spidevref.Number, strings.Join(spidevref.Aliases, ", "))
    }

    log.Printf("Opening the default SPI device")
    p, err := spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

    log.Printf("Connecting to the MFRC522")
    rfid, err := mfrc522.NewSPI(p, rpi.P1_22, nil) // (..., <reset-pin>, <irq-pin>)
	if err != nil {
		log.Fatal(err)
	}
    defer rfid.Halt()

    rfid.SetAntennaGain(5)

    log.Printf("Trying to read UID for 10s...")
    uid, err := rfid.ReadUID(10*time.Second)
    if err != nil {
		log.Fatal(err)
    }
    log.Printf("Read gave: %x", hex.EncodeToString(uid))

    // EVDEV, continued
    events := make(chan InputEvent)

    for _, d := range devicePaths {
        go readEvents(d.Path, d.Name, events)
    }

    for event := range events {
        ts := fmt.Sprintf("Event@%s: time %d.%06d", event.deviceIdentifier, event.evdevEvent.Time.Sec, event.evdevEvent.Time.Usec)

		switch event.evdevEvent.Type {
		case evdev.EV_SYN:
			switch event.evdevEvent.Code {
			case evdev.SYN_MT_REPORT:
				log.Printf("%s, ++++++++++++++ %s ++++++++++++\n", ts, event.evdevEvent.CodeName())
			case evdev.SYN_DROPPED:
				log.Printf("%s, >>>>>>>>>>>>>> %s <<<<<<<<<<<<\n", ts, event.evdevEvent.CodeName())
			default:
				log.Printf("%s, -------------- %s ------------\n", ts, event.evdevEvent.CodeName())
			}
		default:
			log.Printf("%s, %s\n", ts, event.evdevEvent.String())
		}
    }

    return nil
}

func main() {
    if err := doInputProc(); err != nil {
        log.Fatal(err)
    }
}
