# goscrape

A simple job scrapper. Deprecated in favor of [jobet](https://github.com/maxmwang/jobet).  

## How To Use

The scrapper daemon and server, `goscraped`, is a API scrapping daemon and gRPC server. The gRPC client, `goscrape` enables the user to probe companies on demand, adding to the saved companies DB if requested. 

## Technologies:

- [SQLite](https://www.sqlite.org/): Lightweight SQL database. 
- [gRPC](https://grpc.io/): Communication between CLI client and daemon. 
