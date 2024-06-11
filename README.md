# content-unroller

[![CircleCI](https://circleci.com/gh/Financial-Times/content-unroller.svg?style=svg)](https://circleci.com/gh/Financial-Times/content-unroller)
[![Coverage Status](https://coveralls.io/repos/github/Financial-Times/content-unroller/badge.svg?branch=master)](https://coveralls.io/github/Financial-Times/content-unroller?branch=master)

## Introduction

Content unroller is an internally used API for expanding images and dynamic content of an article. It receives an article and returns the same article with some additional data:

* Each image UUID replaced by its actual data. The types of images that are expanded are:
  * main image
  * body embedded images
  * alternative images
  * lead images
* Each dynamic content UUID replaced by its actual data. It will be extracted from `bodyXML`, based on its type (`DynamicContent`)

## Installation

`go get github.com/Financial-Times/content-unroller`

### Run tests locally and compile

```shell
  go test ./... -race
  go install
```

## Usage

### Run service locally with default settings

```shell
  ./content-unroller --help

```

## Endpoints

### Application specific endpoints

| Endpoint           | Description                                                                                              |
|--------------------|----------------------------------------------------------------------------------------------------------|
| `/content`         | Calls **Content-Public-Read** service to expand main images, alternative images and body embedded images + dynamic content |
| `/internalcontent` | Calls **Content-Public-Read** service to expand lead images and body embedded dynamic content            |

### Admin specific endpoints

* /__ping
* /__build-info
* /__health
* /__gtg

## Example 1 (main image)

POST: `/content`
Body:

```json
{
  "id": "http://www.ft.com/thing/22c0d426-1466-11e7-b0c1-37e417ee6c76",
  "type": "http://www.ft.com/ontology/content/Article",
  "bodyXML": "<body></body>",
  "title": "Brexit begins as Theresa May triggers Article 50",
  "alternativeTitles": {
    "promotionalTitle": "Brexit begins as Theresa May triggers Article 50"
  },
  "standfirst": "Prime minister sets out Britain’s negotiating stance in statement to MPs",
  "alternativeStandfirsts": {},
  "byline": "George Parker and Kate Allen in London and Arthur Beesley in Brussels",
  "firstPublishedDate": "2017-03-29T11:07:52.000Z",
  "publishedDate": "2017-03-30T06:54:02.000Z",
  "identifiers": [
    {
      "authority": "http://api.ft.com/system/FTCOM-METHODE",
      "identifierValue": "22c0d426-1466-11e7-b0c1-37e417ee6c76"
    }
  ],
  "requestUrl": "http://test.api.ft.com/content/22c0d426-1466-11e7-b0c1-37e417ee6c76",
  "brands": [
    "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
  ],
  "mainImage": {
    "id": "http://test.api.ft.com/content/639cd952-149f-11e7-2ea7-a07ecd9ac73f"
  },
  "alternativeImages": {},
  "comments": {
    "enabled": true
  },
  "standout": {
    "editorsChoice": false,
    "exclusive": false,
    "scoop": false
  },
  "publishReference": "tid_ra4srof3qc",
  "lastModified": "2017-03-31T15:42:35.266Z",
  "canBeSyndicated": "yes",
  "accessLevel": "subscribed",
  "canBeDistributed": "yes"
}
```

Response:

```json
{
  "accessLevel": "subscribed",
  "alternativeImages": {},
  "alternativeStandfirsts": {},
  "alternativeTitles": {
    "promotionalTitle": "Brexit begins as Theresa May triggers Article 50"
  },
  "bodyXML": "<body></body>",
  "brands": [
    "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
  ],
  "byline": "George Parker and Kate Allen in London and Arthur Beesley in Brussels",
  "canBeDistributed": "yes",
  "canBeSyndicated": "yes",
  "comments": {
    "enabled": true
  },
  "firstPublishedDate": "2017-03-29T11:07:52.000Z",
  "id": "http://www.ft.com/thing/22c0d426-1466-11e7-b0c1-37e417ee6c76",
  "identifiers": [
    {
      "authority": "http://api.ft.com/system/FTCOM-METHODE",
      "identifierValue": "22c0d426-1466-11e7-b0c1-37e417ee6c76"
    }
  ],
  "lastModified": "2017-03-31T15:42:35.266Z",
  "mainImage": {
    "alternativeImages": {},
    "alternativeStandfirsts": {},
    "alternativeTitles": {},
    "brands": [
      "http://www.ft.com/thing/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
    ],
    "canBeDistributed": "yes",
    "description": "Donald Tusk, president of the European Union (EU), holds the letter invoking Article 50 of the Lisbon Treaty from U.K. Prime Minister Theresa May as leaves following a news conference at the European Council in Brussels, Belgium, on Wednesday, March 29, 2017. The U.K. will&nbsp;start the clock&nbsp;on two years of negotiations to withdraw from the European Union on Wednesday, when Britain's ambassador hands EU President Donald Tusk&nbsp;a hand-signed&nbsp;letter&nbsp;from Prime Minister&nbsp;Theresa May&nbsp;invoking Article 50 of the Lisbon Treaty, the legal exit mechanism. Photographer: Jasper Juinen/Bloomberg",
    "firstPublishedDate": "2017-03-29T19:39:00.000Z",
    "id": "http://www.ft.com/thing/639cd952-149f-11e7-2ea7-a07ecd9ac73f",
    "identifiers": [
      {
        "authority": "http://api.ft.com/system/FTCOM-METHODE",
        "identifierValue": "639cd952-149f-11e7-2ea7-a07ecd9ac73f"
      }
    ],
    "lastModified": "2021-10-24T01:58:39.324Z",
    "members": [
      {
        "alternativeImages": {},
        "alternativeStandfirsts": {},
        "alternativeTitles": {},
        "binaryUrl": "http://com.ft.coco-imagepublish.pre-prod.s3.amazonaws.com/639cd952-149f-11e7-b0c1-37e417ee6c76",
        "brands": [
          "http://www.ft.com/thing/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
        ],
        "canBeDistributed": "yes",
        "canBeSyndicated": "yes",
        "copyright": {
          "notice": "© Bloomberg"
        },
        "description": "Donald Tusk, president of the European Union (EU), holds the letter invoking Article 50 of the Lisbon Treaty from U.K. Prime Minister Theresa May as leaves following a news conference at the European Council in Brussels, Belgium, on Wednesday, March 29, 2017. The U.K. will&nbsp;start the clock&nbsp;on two years of negotiations to withdraw from the European Union on Wednesday, when Britain's ambassador hands EU President Donald Tusk&nbsp;a hand-signed&nbsp;letter&nbsp;from Prime Minister&nbsp;Theresa May&nbsp;invoking Article 50 of the Lisbon Treaty, the legal exit mechanism. Photographer: Jasper Juinen/Bloomberg",
        "firstPublishedDate": "2017-03-29T19:39:00.000Z",
        "id": "http://www.ft.com/thing/639cd952-149f-11e7-b0c1-37e417ee6c76",
        "identifiers": [
          {
            "authority": "http://api.ft.com/system/FTCOM-METHODE",
            "identifierValue": "639cd952-149f-11e7-b0c1-37e417ee6c76"
          }
        ],
        "lastModified": "2021-10-24T01:58:39.323Z",
        "masterSource": {
          "authority": "http://api.ft.com/system/FT-FOTOWARE",
          "identifierValue": "260092155B9D4D5A 9283E0C9376C49F5"
        },
        "pixelHeight": 1152,
        "pixelWidth": 2048,
        "publishReference": "tid_methode_carousel_0000596917_gentx",
        "publishedDate": "2017-03-29T19:39:00.000Z",
        "requestUrl": "https://api-t.ft.com/content/639cd952-149f-11e7-b0c1-37e417ee6c76",
        "rightsGroup": "Bloomberg",
        "title": "",
        "type": "http://www.ft.com/ontology/content/Image"
      }
    ],
    "publishReference": "tid_methode_carousel_0000596918_gentx",
    "publishedDate": "2017-03-29T19:39:00.000Z",
    "requestUrl": "https://api-t.ft.com/content/639cd952-149f-11e7-2ea7-a07ecd9ac73f",
    "title": "",
    "type": "http://www.ft.com/ontology/content/ImageSet"
  },
  "publishReference": "tid_ra4srof3qc",
  "publishedDate": "2017-03-30T06:54:02.000Z",
  "requestUrl": "http://test.api.ft.com/content/22c0d426-1466-11e7-b0c1-37e417ee6c76",
  "standfirst": "Prime minister sets out Britain’s negotiating stance in statement to MPs",
  "standout": {
    "editorsChoice": false,
    "exclusive": false,
    "scoop": false
  },
  "title": "Brexit begins as Theresa May triggers Article 50",
  "type": "http://www.ft.com/ontology/content/Article"
}
```

## Example 2 (alternative images)

POST: `/content`
Body:

```json
{
  "id": "http://www.ft.com/thing/bba8a342-28f4-11e7-bc4b-5528796fe35c",
  "type": "http://www.ft.com/ontology/content/Content",
  "title": "At $250m, this is America’s most expensive house",
  "alternativeTitles": {
    "promotionalTitle": "At $250m, this is America’s most expensive house"
  },
  "alternativeStandfirsts": {
    "promotionalStandfirst": "924 Bel Air Road is the frame for the ultimate selfie in which LA forms the backdrop"
  },
  "publishedDate": "2017-05-03T07:20:13.000Z",
  "webUrl": "https://propertylistings.ft.com/propertynews/los-angeles/5024-at-250m-this-is-americas-most-expensive-house.html",
  "identifiers": [
    {
      "authority": "http://api.ft.com/system/FTCOM-METHODE",
      "identifierValue": "bba8a342-28f4-11e7-bc4b-5528796fe35c"
    }
  ],
  "requestUrl": "http://test.api.ft.com/content/bba8a342-28f4-11e7-bc4b-5528796fe35c",
  "brands": [
    "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
  ],
  "alternativeImages": {
    "promotionalImage": "http://test.api.ft.com/content/59b138be-2f36-11e7-9555-23ef563ecf9a"
  },
  "publishReference": "tid_aj4xzjeh7h",
  "lastModified": "2017-05-03T07:24:32.591Z",
  "canBeDistributed": "verify",
  "canBeSyndicated": "verify"
}
```

Response:

```json
{
  "alternativeImages": {
    "promotionalImage": {
      "alternativeImages": {},
      "alternativeStandfirsts": {},
      "alternativeTitles": {},
      "binaryUrl": "http://com.ft.coco-imagepublish.pre-prod.s3.amazonaws.com/59b138be-2f36-11e7-9555-23ef563ecf9a",
      "canBeDistributed": "verify",
      "description": "",
      "firstPublishedDate": "2017-05-02T12:54:00.000Z",
      "id": "http://www.ft.com/thing/59b138be-2f36-11e7-9555-23ef563ecf9a",
      "identifiers": [
        {
          "authority": "http://api.ft.com/system/FTCOM-METHODE",
          "identifierValue": "59b138be-2f36-11e7-9555-23ef563ecf9a"
        }
      ],
      "lastModified": "2017-05-02T12:54:35.658Z",
      "pixelHeight": 1152,
      "pixelWidth": 2048,
      "publishReference": "tid_retjsscm7e",
      "publishedDate": "2017-05-02T12:54:00.000Z",
      "requestUrl": "http://test.api.ft.com/content/59b138be-2f36-11e7-9555-23ef563ecf9a",
      "title": "",
      "type": "http://www.ft.com/ontology/content/MediaResource"
    }
  },
  "alternativeStandfirsts": {
    "promotionalStandfirst": "924 Bel Air Road is the frame for the ultimate selfie in which LA forms the backdrop"
  },
  "alternativeTitles": {
    "promotionalTitle": "At $250m, this is America’s most expensive house"
  },
  "brands": [
    "http://api.ft.com/things/dbb0bdae-1f0c-11e4-b0cb-b2227cce2b54"
  ],
  "canBeDistributed": "verify",
  "canBeSyndicated": "verify",
  "id": "http://www.ft.com/thing/bba8a342-28f4-11e7-bc4b-5528796fe35c",
  "identifiers": [
    {
      "authority": "http://api.ft.com/system/FTCOM-METHODE",
      "identifierValue": "bba8a342-28f4-11e7-bc4b-5528796fe35c"
    }
  ],
  "lastModified": "2017-05-03T07:24:32.591Z",
  "publishReference": "tid_aj4xzjeh7h",
  "publishedDate": "2017-05-03T07:20:13.000Z",
  "requestUrl": "http://test.api.ft.com/content/bba8a342-28f4-11e7-bc4b-5528796fe35c",
  "title": "At $250m, this is America’s most expensive house",
  "type": "http://www.ft.com/ontology/content/Content",
  "webUrl": "https://propertylistings.ft.com/propertynews/los-angeles/5024-at-250m-this-is-americas-most-expensive-house.html"
}
```

## Example 3 (lead images)

POST: `/internalcontent`
Body:

```json
{
  "design": null,
  "tableOfContents": null,
  "topper": null,
  "leadImages": [
    {
      "id": "89f194c8-13bc-11e7-80f4-13e067d5072c",
      "type": "square"
    },
    {
      "id": "3e96c818-13bc-11e7-b0c1-37e417ee6c76",
      "type": "standard"
    },
    {
      "id": "8d7b4e22-13bc-11e7-80f4-13e067d5072c",
      "type": "wide"
    }
  ],
  "id": "http://www.ft.com/thing/5010e2e4-09bd-11e7-97d1-5e720a26771b",
  "uuid": "5010e2e4-09bd-11e7-97d1-5e720a26771b",
  "lastModified": "2017-03-31T08:23:37.061Z",
  "publishReference": "tid_8pqiiuxbvz"
}
```

Response:

```json
{
  "design": null,
  "lastModified": "2017-03-31T08:23:37.061Z",
  "leadImages": [
    {
      "id": "89f194c8-13bc-11e7-80f4-13e067d5072c",
      "image": {
        "alternativeImages": {},
        "alternativeStandfirsts": {},
        "alternativeTitles": {},
        "binaryUrl": "http://com.ft.coco-imagepublish.pre-prod.s3.amazonaws.com/89f194c8-13bc-11e7-80f4-13e067d5072c",
        "canBeDistributed": "verify",
        "description": "",
        "firstPublishedDate": "2017-03-28T13:45:00.000Z",
        "id": "http://www.ft.com/thing/89f194c8-13bc-11e7-80f4-13e067d5072c",
        "identifiers": [
          {
            "authority": "http://api.ft.com/system/FTCOM-METHODE",
            "identifierValue": "89f194c8-13bc-11e7-80f4-13e067d5072c"
          }
        ],
        "lastModified": "2017-03-28T13:45:53.438Z",
        "pixelHeight": 2612,
        "pixelWidth": 2612,
        "publishReference": "tid_lej6rdjegj",
        "publishedDate": "2017-03-28T13:45:00.000Z",
        "requestUrl": "http://test.api.ft.com/content/89f194c8-13bc-11e7-80f4-13e067d5072c",
        "title": "",
        "type": "http://www.ft.com/ontology/content/MediaResource"
      },
      "type": "square"
    },
    {
      "id": "3e96c818-13bc-11e7-b0c1-37e417ee6c76",
      "image": {
        "alternativeImages": {},
        "alternativeStandfirsts": {},
        "alternativeTitles": {},
        "binaryUrl": "http://com.ft.coco-imagepublish.pre-prod.s3.amazonaws.com/3e96c818-13bc-11e7-b0c1-37e417ee6c76",
        "canBeDistributed": "verify",
        "copyright": {
          "notice": "© EPA"
        },
        "description": "",
        "firstPublishedDate": "2017-03-28T13:42:00.000Z",
        "id": "http://www.ft.com/thing/3e96c818-13bc-11e7-b0c1-37e417ee6c76",
        "identifiers": [
          {
            "authority": "http://api.ft.com/system/FTCOM-METHODE",
            "identifierValue": "3e96c818-13bc-11e7-b0c1-37e417ee6c76"
          }
        ],
        "lastModified": "2017-03-28T13:43:00.375Z",
        "pixelHeight": 1152,
        "pixelWidth": 2048,
        "publishReference": "tid_tv7jfgi6jn",
        "publishedDate": "2017-03-28T13:42:00.000Z",
        "requestUrl": "http://test.api.ft.com/content/3e96c818-13bc-11e7-b0c1-37e417ee6c76",
        "title": "Leader of the PVV party Gert Wilders reacts to the election result",
        "type": "http://www.ft.com/ontology/content/MediaResource"
      },
      "type": "standard"
    },
    {
      "id": "8d7b4e22-13bc-11e7-80f4-13e067d5072c",
      "image": {
        "alternativeImages": {},
        "alternativeStandfirsts": {},
        "alternativeTitles": {},
        "binaryUrl": "http://com.ft.coco-imagepublish.pre-prod.s3.amazonaws.com/8d7b4e22-13bc-11e7-80f4-13e067d5072c",
        "canBeDistributed": "verify",
        "description": "",
        "firstPublishedDate": "2017-03-28T13:45:00.000Z",
        "id": "http://www.ft.com/thing/8d7b4e22-13bc-11e7-80f4-13e067d5072c",
        "identifiers": [
          {
            "authority": "http://api.ft.com/system/FTCOM-METHODE",
            "identifierValue": "8d7b4e22-13bc-11e7-80f4-13e067d5072c"
          }
        ],
        "lastModified": "2017-03-28T13:45:53.525Z",
        "pixelHeight": 1548,
        "pixelWidth": 4645,
        "publishReference": "tid_prlsj2avbn",
        "publishedDate": "2017-03-28T13:45:00.000Z",
        "requestUrl": "http://test.api.ft.com/content/8d7b4e22-13bc-11e7-80f4-13e067d5072c",
        "title": "",
        "type": "http://www.ft.com/ontology/content/MediaResource"
      },
      "type": "wide"
    }
  ],
  "publishReference": "tid_8pqiiuxbvz",
  "tableOfContents": null,
  "topper": null,
  "uuid": "5010e2e4-09bd-11e7-97d1-5e720a26771b"
}
```
