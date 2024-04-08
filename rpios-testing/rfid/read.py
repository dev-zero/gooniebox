#!/usr/bin/env python3

from time import sleep
import sys
from mfrc522 import MFRC522, BasicMFRC522
import RPi.GPIO as GPIO

mfrc = MFRC522(device=0)
reader = BasicMFRC522()

try:
    while True:
        print("Hold a tag near the reader")
        id, text = reader.read_no_block(11)
        print("ID: %s\nText: %s" % (id,text))
        sleep(5)
except KeyboardInterrupt:
    GPIO.cleanup()
