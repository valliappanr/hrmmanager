# Overview
Heart rate monitor for Polar device could be retrieved using GATT BLE. Using a bluetooth dongle. A bluetooth Dongle is connected to USB port and through that the recording can be retrieved.
The problem trying to run through kubernetes is that it need to be connected to a specific node and could be accessed by all other nodes using USB IP. USB IP wasn't working with the polar device
on arm64 raspberry pi. The workaround solution is to run on specific node and run all other apps on high available manner. And on the specific node run a controller app to make sure that the app is
always up and running. This application is a controller to make sure that driver application is up and running, if it is not running, it will restart the application.
