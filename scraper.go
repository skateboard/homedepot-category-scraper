package main

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	http "github.com/bogdanfinn/fhttp"
	"github.com/data-harvesters/goapify"
	goapifytls "github.com/data-harvesters/goapify-tls"
	"github.com/skateboard/ajson"
)

var (
	mainStoreID = "0915" // biggest store ID

	productTimeouts     = make(map[string]time.Time)
	productTimeoutsSync sync.Mutex
)

type scraper struct {
	actor *goapify.Actor
	input *input

	client *goapifytls.TlsClient
}

func newScraper(input *input, actor *goapify.Actor) (*scraper, error) {
	tlsClient, err := goapifytls.NewTlsClient(actor, goapifytls.DefaultOptions())
	if err != nil {
		return nil, err
	}

	return &scraper{
		actor:  actor,
		input:  input,
		client: tlsClient,
	}, nil
}

func (s *scraper) Run() {
	fmt.Println("beginning scrapping...")

	q := NewQueue()
	finished := make(chan bool)

	go func() {
		for {
			i := q.Pop()
			if i == nil {
				break
			}
			go s.startScrapeCategory(i.(string))

			time.Sleep(500 * time.Millisecond)
		}

		finished <- true
	}()

	for _, categoryID := range s.input.CategoryIDs {
		go func(categoryID string) {
			q.Push(categoryID)
		}(categoryID)
	}
	<-finished
	fmt.Println("succesfully scraped all products from categories")
}

func (s *scraper) startScrapeCategory(categoryID string) {
	startIndex := int64(s.input.Offset)
	totalProducts := int64(0)

	fmt.Printf("%s: startIndex: %d\n", categoryID, startIndex)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("%s: finished\n", categoryID)
			return
		default:
			client := s.client.ProxiedClient()
			jar := s.client.GetCookieJar()
			client.SetCookieJar(jar)

			payload := strings.NewReader(fmt.Sprintf(`{
				"operationName": "searchModel",
				"variables": {
					"navParam": "%v",
					"pageSize": 48,
					"startIndex": %v,
					"storeId": "%v",
					"additionalSearchParams": {
						"plp": "true"
					}
				},
					"query": "query searchModel($keyword: String, $navParam: String, $storefilter: StoreFilter = ALL, $storeId: String, $itemIds: [String], $channel: Channel = DESKTOP, $additionalSearchParams: AdditionalParams, $loyaltyMembershipInput: LoyaltyMembershipInput, $startIndex: Int, $pageSize: Int, $orderBy: ProductSort, $filter: ProductFilter, $zipCode: String, $skipInstallServices: Boolean = true, $skipKPF: Boolean = false, $skipSpecificationGroup: Boolean = false, $skipSubscribeAndSave: Boolean = false) {\n  searchModel(keyword: $keyword, navParam: $navParam, storefilter: $storefilter, storeId: $storeId, itemIds: $itemIds, channel: $channel, additionalSearchParams: $additionalSearchParams, loyaltyMembershipInput: $loyaltyMembershipInput) {\n    metadata {\n      hasPLPBanner\n      categoryID\n      analytics {\n        semanticTokens\n        dynamicLCA\n        __typename\n      }\n      canonicalUrl\n      searchRedirect\n      clearAllRefinementsURL\n      contentType\n      isStoreDisplay\n      productCount {\n        inStore\n        __typename\n      }\n      stores {\n        storeId\n        storeName\n        address {\n          postalCode\n          __typename\n        }\n        nearByStores {\n          storeId\n          storeName\n          distance\n          address {\n            postalCode\n            __typename\n          }\n          __typename\n        }\n        __typename\n      }\n      __typename\n    }\n    id\n    searchReport {\n      totalProducts\n      didYouMean\n      correctedKeyword\n      keyword\n      pageSize\n      searchUrl\n      sortBy\n      sortOrder\n      startIndex\n      __typename\n    }\n    relatedResults {\n      universalSearch {\n        title\n        __typename\n      }\n      relatedServices {\n        label\n        __typename\n      }\n      visualNavs {\n        label\n        imageId\n        webUrl\n        categoryId\n        imageURL\n        __typename\n      }\n      visualNavContainsEvents\n      relatedKeywords {\n        keyword\n        __typename\n      }\n      __typename\n    }\n    products(startIndex: $startIndex, pageSize: $pageSize, orderBy: $orderBy, filter: $filter) {\n      itemId\n      dataSources\n      identifiers {\n        canonicalUrl\n        brandName\n        itemId\n        productLabel\n        modelNumber\n        productType\n        storeSkuNumber\n        parentId\n        isSuperSku\n        __typename\n      }\n      media {\n        images {\n          url\n          type\n          subType\n          sizes\n          __typename\n        }\n        __typename\n      }\n      pricing(storeId: $storeId) {\n        value\n        alternatePriceDisplay\n        alternate {\n          bulk {\n            pricePerUnit\n            thresholdQuantity\n            value\n            __typename\n          }\n          unit {\n            caseUnitOfMeasure\n            unitsOriginalPrice\n            unitsPerCase\n            value\n            __typename\n          }\n          __typename\n        }\n        original\n        mapAboveOriginalPrice\n        message\n        preferredPriceFlag\n        promotion {\n          type\n          description {\n            shortDesc\n            longDesc\n            __typename\n          }\n          dollarOff\n          percentageOff\n          savingsCenter\n          savingsCenterPromos\n          specialBuySavings\n          specialBuyDollarOff\n          specialBuyPercentageOff\n          dates {\n            start\n            end\n            __typename\n          }\n          __typename\n        }\n        specialBuy\n        unitOfMeasure\n        __typename\n      }\n      reviews {\n        ratingsReviews {\n          averageRating\n          totalReviews\n          __typename\n        }\n        __typename\n      }\n      availabilityType {\n        discontinued\n        type\n        __typename\n      }\n      badges(storeId: $storeId) {\n        name\n        __typename\n      }\n      details {\n        collection {\n          collectionId\n          name\n          url\n          __typename\n        }\n        highlights\n        __typename\n      }\n      favoriteDetail {\n        count\n        __typename\n      }\n      fulfillment(storeId: $storeId, zipCode: $zipCode) {\n        backordered\n        backorderedShipDate\n        bossExcludedShipStates\n        excludedShipStates\n        seasonStatusEligible\n        fulfillmentOptions {\n          type\n          fulfillable\n          services {\n            type\n            hasFreeShipping\n            freeDeliveryThreshold\n            locations {\n              curbsidePickupFlag\n              isBuyInStoreCheckNearBy\n              distance\n              inventory {\n                isOutOfStock\n                isInStock\n                isLimitedQuantity\n                isUnavailable\n                quantity\n                maxAllowedBopisQty\n                minAllowedBopisQty\n                __typename\n              }\n              isAnchor\n              locationId\n              storeName\n              state\n              type\n              __typename\n            }\n            __typename\n          }\n          __typename\n        }\n        __typename\n      }\n      info {\n        hasSubscription\n        isBuryProduct\n        isSponsored\n        isGenericProduct\n        isLiveGoodsProduct\n        sponsoredBeacon {\n          onClickBeacon\n          onViewBeacon\n          __typename\n        }\n        sponsoredMetadata {\n          campaignId\n          placementId\n          slotId\n          __typename\n        }\n        globalCustomConfigurator {\n          customExperience\n          __typename\n        }\n        returnable\n        hidePrice\n        productSubType {\n          name\n          link\n          __typename\n        }\n        categoryHierarchy\n        samplesAvailable\n        customerSignal {\n          previouslyPurchased\n          __typename\n        }\n        productDepartmentId\n        productDepartment\n        augmentedReality\n        ecoRebate\n        quantityLimit\n        sskMin\n        sskMax\n        unitOfMeasureCoverage\n        wasMaxPriceRange\n        wasMinPriceRange\n        swatches {\n          isSelected\n          itemId\n          label\n          swatchImgUrl\n          url\n          value\n          __typename\n        }\n        totalNumberOfOptions\n        paintBrand\n        dotComColorEligible\n        __typename\n      }\n      installServices(storeId: $storeId, zipCode: $zipCode) @skip(if: $skipInstallServices) {\n        scheduleAMeasure\n        gccCarpetDesignAndOrderEligible\n        __typename\n      }\n      keyProductFeatures @skip(if: $skipKPF) {\n        keyProductFeaturesItems {\n          features {\n            name\n            refinementId\n            refinementUrl\n            value\n            __typename\n          }\n          __typename\n        }\n        __typename\n      }\n      specificationGroup @skip(if: $skipSpecificationGroup) {\n        specifications {\n          specName\n          specValue\n          __typename\n        }\n        specTitle\n        __typename\n      }\n      subscription @skip(if: $skipSubscribeAndSave) {\n        defaultfrequency\n        discountPercentage\n        subscriptionEnabled\n        __typename\n      }\n      sizeAndFitDetail {\n        attributeGroups {\n          attributes {\n            attributeName\n            dimensions\n            __typename\n          }\n          dimensionLabel\n          productType\n          __typename\n        }\n        __typename\n      }\n      __typename\n    }\n    taxonomy {\n      brandLinkUrl\n      breadCrumbs {\n        browseUrl\n        creativeIconUrl\n        deselectUrl\n        dimensionId\n        dimensionName\n        label\n        refinementKey\n        url\n        __typename\n      }\n      __typename\n    }\n    templates\n    partialTemplates\n    dimensions {\n      label\n      refinements {\n        refinementKey\n        label\n        recordCount\n        selected\n        imgUrl\n        url\n        nestedRefinements {\n          label\n          url\n          recordCount\n          refinementKey\n          __typename\n        }\n        __typename\n      }\n      collapse\n      dimensionId\n      isVisualNav\n      isVisualDimension\n      isNumericFilter\n      nestedRefinementsLimit\n      visualNavSequence\n      __typename\n    }\n    orangeGraph {\n      universalSearchArray {\n        pods {\n          title\n          description\n          imageUrl\n          link\n          __typename\n        }\n        info {\n          title\n          __typename\n        }\n        __typename\n      }\n      productTypes\n      __typename\n    }\n    appliedDimensions {\n      label\n      refinements {\n        label\n        refinementKey\n        url\n        __typename\n      }\n      isNumericFilter\n      __typename\n    }\n    __typename\n  }\n}\n"
			}`, categoryID, startIndex, mainStoreID))

			_, err := client.Get("https://www.homedepot.com/") // get the home page
			if err != nil {
				fmt.Printf("%s: failed to get homepage: %v\n", categoryID, err)
				continue
			}

			req, err := http.NewRequest("POST", "https://www.homedepot.com/federation-gateway/graphql?opname=searchModel",
				payload)
			if err != nil {
				fmt.Printf("%s: failed to get products: %v\n", categoryID, err)
				continue
			}

			req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/110.0")
			req.Header.Add("Accept", "*/*")
			req.Header.Add("Accept-Language", "en-US,en;q=0.5")
			req.Header.Add("Referer", "https://www.homedepot.com/b/Bath-Bathroom-Exhaust-Fans-Bath-Fans/N-5yc1vZc4kq")
			req.Header.Add("content-type", "application/json")
			req.Header.Add("X-Experience-Name", "b2b")
			req.Header.Add("apollographql-client-name", "b2b")
			req.Header.Add("apollographql-client-version", "0.0.0")
			req.Header.Add("X-current-url", "/b/Bath-Bathroom-Exhaust-Fans-Bath-Fans/N-5yc1vZc4kq")
			req.Header.Add("x-hd-dc", "origin")
			req.Header.Add("x-customer-type", "B2B")
			req.Header.Add("x-customer-role", "ADMIN")
			req.Header.Add("x-segment-id", "Contractors")
			req.Header.Add("Origin", "https://www.homedepot.com")
			req.Header.Add("Connection", "keep-alive")
			req.Header.Add("Sec-Fetch-Dest", "empty")
			req.Header.Add("Sec-Fetch-Mode", "cors")
			req.Header.Add("Sec-Fetch-Site", "same-origin")
			req.Header.Add("Pragma", "no-cache")
			req.Header.Add("Cache-Control", "no-cache")
			req.Header.Add("TE", "trailers")

			res, err := client.Do(req)
			if err != nil {
				fmt.Printf("%s: failed sending request: %v\n", categoryID, err)
				continue
			}

			b, err := io.ReadAll(res.Body)
			if err != nil {
				fmt.Printf("%s: failed reading body: %v\n", categoryID, err)
				continue
			}
			res.Body.Close()

			if res.StatusCode != 200 {
				fmt.Printf("%s: failed getting products: %d %s\n", categoryID, res.StatusCode, string(b))
				continue
			}

			j := ajson.Parse(string(b))
			searchModel := j.Get("data").Get("searchModel")

			if totalProducts == 0 {
				searchReport := searchModel.Get("searchReport")
				tProducts := searchReport.Get("totalProducts").Int()

				fmt.Printf("%s: total products: %d\n", categoryID, tProducts)

				totalProducts = tProducts
			}
			products := searchModel.Get("products")

			var prods []map[string]any
			for _, product := range products.Array() {
				itemId := product.Get("itemId").String()
				identifiers := product.Get("identifiers")
				image := strings.Replace(product.Get("media").Get("images").Array()[0].Get("url").String(), "<SIZE>", "300", -1)

				productTimeoutsSync.Lock()
				if time.Now().After(productTimeouts[itemId].Add(24 * time.Hour)) {
					prods = append(prods, map[string]any{
						"sku":   itemId,
						"name":  identifiers.Get("productLabel").String(),
						"brand": identifiers.Get("brandName").String(),
						"image": image,
						"url":   fmt.Sprintf("https://www.homedepot.com%s", identifiers.Get("canonicalUrl").String()),
						"price": math.Round(product.Get("pricing").Get("original").Float()),
					})
					productTimeouts[itemId] = time.Now()
				}
				productTimeoutsSync.Unlock()
			}

			if startIndex >= totalProducts { // we are done with this category, reset the start index
				startIndex = 0            // reset the start index
				time.Sleep(1 * time.Hour) // sleep for an hour before starting again to avoid rate limiting
			} else {
				startIndex += 48 // 48 is the max number of products per page
			}

			err = s.actor.Output(prods)
			if err != nil {
				fmt.Printf("%s: failed sending output: %v\n", categoryID, err)
				continue
			}

			time.Sleep(500 * time.Millisecond)
		}
	}

}
