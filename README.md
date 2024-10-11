![enter image description here](https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Fwww.pngmart.com%2Ffiles%2F16%2FHome-Depot-Logo-PNG-Image.png&f=1&nofb=1&ipt=def5757aced3b327c03b99e5ef3280e8f0c80bc271e4c8682ef7f1c5bd1f976f&ipo=images)

# Homedepot Category Scraper

## About This Actor

This Actor is a powerful, user-fiendly tool made to scrape products from specified Homedepot Categories. This tool will save you time and provide you with reliable data on products from Homedepot.

Made with Golang 1.22.1

## Tutorial

Basic Usage

```json
{
    "categoryIds": ["12345689"],
    "offset": 0
}
```

| parameter | type | argument | description |
| --------- | ----- | ------------------------- | ---------------------------- |
| categoryIds | array | _["1223", "12312312", ...]_ | An array of category ids |
| offset | int | _default=0_ | Start from a specific offset |

### Output Sample

```json
[
    {
        "brand": "BRAND",
        "image": "https://images.thdstatic.com/productImages/UUID/svn/image.jpg",
        "name": "NAME",
        "price": 50,
        "sku": "SKU",
        "url": "https://www.homedepot.com/p/NAME/SKU"
    }
]

```
