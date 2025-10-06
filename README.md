# Fujigo - A CLI tool to interact with old Fujifilm cameras over serial
# ================================================================

Old Fujifilm cameras, such as DX-5 are pretty cool. But they use SmartMedia cards which can be quite difficult to read on modern computers. The only alternative to get images off these cameras when SmartMedia card is being capricious is to use the serial interface of the camera. The official software Fujifilm provides is Windows-only and does not run on modern computers (i mean post Windows 98). This tool aims to provide an alternative way to access images stored on these cameras.

This tool is based on the protocol analysis you can find here: https://christian1.tripod.com/FujiMX.html. 

On linux, you can already use the great fujiplay tool to get images off these cameras: https://www.imo.universite-paris-saclay.fr/~thierry.bousch/fujiplay.html. 

This tool is still in early development. Currently, it can only list available serial ports and get information about the connected camera model. In the future, extracting images and other functionalities may be added. The goal is to develop a cross-platform tool for the few people still using these cameras!
