<div align="center">
<h1>Begonia</h1>
<p>
A gateway service that reverse proxies HTTP requests to gRPC services.
</p>
</div>

# About

Begonia is an HTTP to gRPC reverse proxy server that registers service routes to the gateway based on descriptor_set_out generated by protoc, thus implementing reverse proxy functionality. The HTTP service adheres to RESTful standards to handle HTTP requests and forwards RESTful requests to gRPC services
