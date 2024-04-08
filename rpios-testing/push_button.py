#!/usr/bin/env python3

from gpiozero import Button
button = Button(13, pull_up=False)
print("waiting for button to be pressed...")
button.wait_for_press()
print("Button was pressed")
