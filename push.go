package main

/**
* @todo : implement a native push service
*
* it should be used according to parameter provided in cli command
* expected options:
* --push-service <url> : complete url to channel to send data, if specified this will override all other options
* --push-port <int> ; port to use
* --push-host <hostname, default=localhost>
*
* host + port combined will build complete url to the channel
* only applied if push-service is not specified,
* in that case we will serve natively this push service as a built-in websocket server
*
**/