# decWebRTC

Decentralized version of WebRTC.

## Running

To run the project, first compile from the root folder (it should be located in `$GOPATH/src/github.com/pauarge/decWebRTC`):

```
go build
```

It's possible that some additional packages need to be installed using `go get`.

Then, simply run the program by 

```
./decWebRTC -name=somename
```

Other command line options are:

```
  -disableGui
    	Disable GUI
  -gossipPort int
    	Port in which gossips are listened (default 5000)
  -guiPort int
    	Port in which the GUI is offered (default 8080)
  -name string
    	Name of the node
  -peers string
    	List of peers
  -rtimer int
    	How many seconds the peer waits between two route rumor messages (default 10)
```

Peers can be added using the CLI option or through the GUI.

To navigate to the GUI, with the program running, open a compatible browser (Firefox works the best) and navigate to [https://127.0.0.1:8080](https://127.0.0.1:8080). It's possible that a security exception for the SSL certificate must be added.

To run a TURN server, please install, setup and run [Coturn](https://github.com/coturn/coturn) [with its default configuration](https://github.com/coturn/coturn/wiki/CoturnConfig).