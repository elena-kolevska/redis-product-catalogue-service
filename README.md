# Redis Product Catalogue Service
[https://elena-kolevska.github.io/redis-product-catalogue-service/]()

## Overview
A simple REST API providing a way to manage and browse products

## Context
This project is done as a simple exercise for data modelling with Redis in Go

## Logical Data Model
Image  
- Id : Number
- Value : Binary
- Product: Product (1)  

Product  
- Id : Number
- Name : String
- Description: String
- Vendor : String
- Price : Number
- Currency : String
- MainCategory : Category (1)
- Images : Image (0..n)

Category
- Id : Number
- Name : String
- Products : Product (0..n)

## Physical Data Model
![data Model](data_model.png "Redis Data Model")

## Documentation
You can find the API documentation [here](https://elena-kolevska.github.io/redis-product-catalogue-service/).  

It's written using the OpenAPI 3 specification, with the `yaml` files living in the `docs` folder.  

To export the documentation to a self-contained, zero-dependencies html file, you can use the [ReDoc library](https://github.com/Redocly/redoc). Just run the following command:  

```redoc-cli bundle docs/index.yaml --output docs/index.html --title "Redis Product Catalogue Service Documentation" --options.theme.colors.primary.main=#D82C20```

## Configuration
When setting up the program rename the `conf_example.json` file to `conf.json` and populate it with your values. 

## Tests
Execute tests the standard way, by running
```bash
go test
``` 

## TODO
- Implement authentication