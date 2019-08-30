# Redis Product Catalogue Service

## Overview
A simple REST API providing a way to manage and browse products

## Context
This project is done as a simple exercise for data modelling with Redis and Go

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
//TODO Add diagram

## Documentation

## Tests