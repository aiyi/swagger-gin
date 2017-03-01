
## [DEPRECATED]
Generate REST API boilerplate code from Swagger 2.0 API spec for [gin] web framework. Thanks to the [go-swagger] project, it's awesome and has been heavily used as code base of swagger-gin. 

## Usage
Install the package
```sh
go get github.com/aiyi/swagger-gin
```

<b> To generate source files (models, restapi and example operations) </b>

Use default spec file "./swagger.json" and default target directory "./":
```sh
swagger-gin
```
Use specific api spec:
```sh
swagger-gin -spec=petstore.json
```
Use specific target folder:
```sh
swagger-gin -spec=petstore.json -target=petstore
```

[Gin]: http://gin-gonic.github.io/gin/
[go-swagger]: https://github.com/go-swagger/go-swagger
