package bingads

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
)

type BufferCloser struct {
	buf *bytes.Buffer
}

func (b BufferCloser) Close() error {
	return nil
}

func (b BufferCloser) Read(p []byte) (int, error) {
	return b.buf.Read(p)
}

type StringClient string

var envheader = `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Header><h:TrackingId xmlns:h="https://bingads.microsoft.com/CampaignManagement/v11">fc253ead-0334-4e2f-a4b9-35aa15203dd3</h:TrackingId></s:Header><s:Body>`
var envend = `</s:Body></s:Envelope>`

func (s StringClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Body: BufferCloser{bytes.NewBufferString(string(s))},
	}, nil
}

func getIds(xs []NegativeKeywordList) []int64 {
	r := make([]int64, len(xs))

	for i := 0; i < len(xs); i++ {
		r[i] = xs[i].Id
	}
	return r
}

func TestSandboxGetSharedEntitiesForSet(t *testing.T) {
	svc := getTestClient()

	existing, err := svc.GetSharedEntitiesByAccountId("NegativeKeywordList")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(existing)

	items, err := svc.GetListItemsBySharedList(&NegativeKeywordList{Id: existing[0].Id})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(items)

	/*
		ids := make([]int64, len(existing))
		for i, x := range existing {
			ids[i] = x.Id
		}

		ents, err := svc.GetSharedEntityAssociationsBySharedEntityIds(ids)
		if err != nil {
			t.Fatal(existing)
		}

		fmt.Println(ents)
	*/
}

func TestSandboxGetSharedEntities(t *testing.T) {
	svc := getTestClient()

	existing, err := svc.GetSharedEntitiesByAccountId("NegativeKeywordList")
	if err != nil {
		t.Fatal(existing)
	}

	a, err := svc.GetSharedEntityAssociationsByEntityIds([]int64{804004280})

	if err != nil {
		t.Fatal(err)
	}

	if len(a.Associations) > 0 {
		partials, err := svc.DeleteSharedEntityAssociations(a.Associations)
		if err != nil {
			t.Fatal(err)
		}

		if len(partials) > 0 {
			t.Fatal(partials)
		}
	}

	partials, err := svc.DeleteSharedEntities(existing)
	if err != nil {
		t.Fatal(err)
	}

	if len(partials) > 0 {
		t.Fatal(partials)
	}

	items := []NegativeKeyword{{
		//Id:        63001000817,
		MatchType: "Phrase",
		Text:      "asdf-1",
	}}

	added, err := svc.AddSharedEntity(&NegativeKeywordList{
		Name: "asdf negative keyword list",
	}, items)
	if err != nil {
		t.Fatal(err)
	}

	items2 := []NegativeKeyword{{
		//Id:        63001000817,
		MatchType: "Phrase",
		Text:      "asdf-2",
	}}

	sa, err := svc.AddListItemsToSharedList(&NegativeKeywordList{Id: added.SharedEntityId}, items2)
	if err != nil {
		t.Fatal(err)
	}

	if len(sa.PartialErrors) > 0 {
		t.Fatal(sa.PartialErrors)
	}

	err = svc.SetSharedEntityAssociations([]SharedEntityAssociation{
		{
			EntityId:         804004280,
			EntityType:       "Campaign",
			SharedEntityId:   added.SharedEntityId,
			SharedEntityType: "NegativeKeywordList",
		},
	})

	if err != nil {
		t.Error(err)
	}
}

func TestUnmarshalResponse(t *testing.T) {

	s := StringClient(envheader + `<ApplyProductPartitionActionsResponse xmlns="https://bingads.microsoft.com/CampaignManagement/v11"><AdGroupCriterionIds xmlns:a="http://schemas.datacontract.org/2004/07/System" xmlns:i="http://www.w3.org/2001/XMLSchema-instance"><a:long>1</a:long><a:long>2</a:long><a:long>3</a:long></AdGroupCriterionIds><PartialErrors xmlns:i="http://www.w3.org/2001/XMLSchema-instance"/></ApplyProductPartitionActionsResponse>` + envend)
	svc := &CampaignService{
		Endpoint: "https://campaign.api.sandbox.bingads.microsoft.com/Api/Advertiser/CampaignManagement/v11/CampaignManagementService.svc",
		Session:  &Session{HTTPClient: s},
	}

	res, err := svc.ApplyProductPartitionActions(nil)

	if err != nil {
		t.Error(err)
	}
	if len(res.AdGroupCriterionIds) != 3 {
		t.Errorf("expected 3 ids, got %d", len(res.AdGroupCriterionIds))
	}

	s = StringClient(envheader + `<ApplyProductPartitionActionsResponse xmlns="https://bingads.microsoft.com/CampaignManagement/v11"><AdGroupCriterionIds xmlns:a="http://schemas.datacontract.org/2004/07/System" xmlns:i="http://www.w3.org/2001/XMLSchema-instance"><a:long i:nil="true"/></AdGroupCriterionIds><PartialErrors xmlns:i="http://www.w3.org/2001/XMLSchema-instance"><BatchError><Code>4129</Code><Details i:nil="true"/><ErrorCode>CampaignServiceDuplicateProductConditions</ErrorCode><FieldPath>ProductConditions</FieldPath><ForwardCompatibilityMap i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/System.Collections.Generic"/><Index>0</Index><Message>Children of a product partition node cannot contain duplicate product conditions.</Message><Type>BatchError</Type></BatchError></PartialErrors></ApplyProductPartitionActionsResponse>` + envend)
	svc = &CampaignService{
		Endpoint: "https://campaign.api.sandbox.bingads.microsoft.com/Api/Advertiser/CampaignManagement/v11/CampaignManagementService.svc",
		Session:  &Session{HTTPClient: s},
	}

	res, err = svc.ApplyProductPartitionActions(nil)
	if err != nil {
		t.Error(err)
	}

	if len(res.AdGroupCriterionIds) != 0 {
		t.Errorf("expected no ids")
	}
}

func TestSandboxApplyProductPartitionActions(t *testing.T) {
	svc := getTestClient()

	existing, err := svc.GetAdGroupCriterionsByIds(1167681348701053)
	if err != nil {
		t.Fatal(err)
	}

	if len(existing) == 0 {
		t.Fatalf("expected all products root group to exist")
	}

	fmt.Println(existing)

	parentid := fmt.Sprintf("%d", existing[0].Id)

	tomodifyid, err := func() (int64, error) {
		for _, e := range existing {
			if e.Criterion.Condition.Attribute == "str" {
				return e.Id, nil
			}
		}
		return 0, fmt.Errorf("expected existing str partition")
	}()

	if err != nil {
		t.Fatal(err)
	}

	todelete, _ := func() (int64, error) {
		for _, e := range existing {
			if e.Criterion.Condition.Attribute == "int" {
				return e.Id, nil
			}
		}
		return 0, fmt.Errorf("expected existing str partition")
	}()

	cleanup := func(id int64) {
		toremove := BiddableAdGroupCriterion{
			AdGroupId: 1167681348701053,
			Id:        id,
		}

		actions := []AdGroupCriterionAction{{"Delete", toremove}}
		res, err := svc.ApplyProductPartitionActions(actions)

		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(res)
		if len(res.AdGroupCriterionIds) != 1 {
			t.Fatalf("expected 1 delete, got %d", len(res.AdGroupCriterionIds))
		}
	}

	if todelete != 0 {
		cleanup(todelete)
	}

	a := BiddableAdGroupCriterion{
		AdGroupId: 1167681348701053,
		Criterion: ProductPartition{
			Condition:         ProductCondition{"int", "ProductType1"},
			ParentCriterionId: parentid,
			PartitionType:     "Unit",
		},
		Status: "Active",
		CriterionBid: CriterionBid{
			Amount: 0.35,
		},
	}

	b := BiddableAdGroupCriterion{
		AdGroupId: 1159984767306214,
		Criterion: ProductPartition{
			Condition:         ProductCondition{"str", "ProductType1"},
			ParentCriterionId: parentid,
			PartitionType:     "Unit",
		},
		Status: "Active",
		CriterionBid: CriterionBid{
			Amount: 0.35,
		},
	}

	c := BiddableAdGroupCriterion{
		AdGroupId: 1167681348701053,
		Id:        tomodifyid,
		Criterion: ProductPartition{
			Condition:         ProductCondition{"str", "ProductType1"},
			ParentCriterionId: parentid,
			PartitionType:     "Unit",
		},
		Status: "Active",
		CriterionBid: CriterionBid{
			Amount: 0.35,
		},
	}

	expectedPartialFailures := map[int]bool{
		0: true,
		2: true,
	}

	actions := []AdGroupCriterionAction{
		{"Add", b},
		{"Add", a},
		{"Add", b},
		{"Update", c},
	}

	applied, err := svc.ApplyProductPartitionActions(actions)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(applied)

	expectedSuccessful := len(actions) - len(expectedPartialFailures)

	if len(applied.AdGroupCriterionIds) != expectedSuccessful {
		t.Errorf("expected %d created, got %d", expectedSuccessful, len(applied.AdGroupCriterionIds))
	}

	for _, failure := range applied.PartialErrors {
		if !expectedPartialFailures[failure.Index] {
			t.Errorf("unexpected partial failure at index %d", failure.Index)
		}
	}

	cleanup(applied.AdGroupCriterionIds[0])
}

func TestUnmarshalCampaignScope(t *testing.T) {
	client := StringClient(envheader + `<GetCampaignCriterionsByIdsResponse xmlns="https://bingads.microsoft.com/CampaignManagement/v11"><CampaignCriterions xmlns:i="http://www.w3.org/2001/XMLSchema-instance"><CampaignCriterion i:type="BiddableCampaignCriterion"><CampaignId>283025743</CampaignId><Criterion i:type="ProductScope"><Type>ProductScope</Type><Conditions><ProductCondition><Attribute>top_brand</Attribute><Operand>CustomLabel0</Operand></ProductCondition></Conditions></Criterion><ForwardCompatibilityMap i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/System.Collections.Generic"/><Id>840001008</Id><Status i:nil="true"/><Type>BiddableCampaignCriterion</Type><CriterionBid i:nil="true"/></CampaignCriterion></CampaignCriterions><PartialErrors i:nil="true" xmlns:i="http://www.w3.org/2001/XMLSchema-instance"/></GetCampaignCriterionsByIdsResponse>` + envend)
	s := &CampaignService{
		Session: &Session{HTTPClient: client},
	}

	res, err := s.GetCampaignCriterionsByIds(123)

	if err != nil {
		t.Fatal(err)
	}

	//	{283025743 {ProductScope [{top_brand CustomLabel0}]  } 840001008  BiddableCampaignCriterion}]
	expected := []CampaignCriterion{{
		CampaignId: 283025743,
		Type:       "BiddableCampaignCriterion",
		Id:         840001008,
		Criterion: ProductScope{
			Type:       "ProductScope",
			Conditions: []ProductCondition{{"top_brand", "CustomLabel0"}},
		},
	}}

	if !reflect.DeepEqual(res, expected) {
		t.Errorf("expected %v, got %v", expected, res)
	}
}

func getTestClient() *CampaignService {
	client := &Session{
		AccountId:      os.Getenv("BING_ACCOUNT_ID"),
		CustomerId:     os.Getenv("BING_CUSTOMER_ID"),
		Username:       os.Getenv("BING_USERNAME"),
		Password:       os.Getenv("BING_PASSWORD"),
		DeveloperToken: os.Getenv("BING_DEV_TOKEN"),
		HTTPClient:     &http.Client{},
	}

	return &CampaignService{
		Endpoint: "https://campaign.api.sandbox.bingads.microsoft.com/Api/Advertiser/CampaignManagement/v11/CampaignManagementService.svc",
		Session:  client,
	}
}

func TestAddAdGroupSandbox(t *testing.T) {

	svc := getTestClient()

	toadd := []AdGroup{
		{
			Name:           "new adgroup1",
			Language:       "English",
			Network:        "OwnedAndOperatedOnly",
			AdDistribution: "Search",
			BiddingScheme:  BiddingScheme("ManualCpcBiddingScheme"),
		},
	}

	res, err := svc.AddAdGroups(804004280, toadd)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(res)
}

//TODO: DeleteCampaignCriterions
func TestSandboxAddCampaignCriterions(t *testing.T) {
	svc := getTestClient()
	crits, err := svc.GetCampaignCriterionsByIds(804004280)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("criterions:", crits)

	cs := []CampaignCriterion{
		CampaignCriterion{
			Type:       "BiddableCampaignCriterion",
			CampaignId: 804004280,
			Criterion: ProductScope{
				Type:       "ProductScope",
				Conditions: []ProductCondition{{"valve", "Brand"}},
				//	PartitionType: "Subdivision",
			},
			CriterionBid: CriterionBid{"FixedBid", 0.03},
			Status:       "Active",
		},
	}

	res, err := svc.AddCampaignCriterions("ProductScope", cs)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(res)
}

func TestAddCampaigns(t *testing.T) {

	svc := getTestClient()

	toadd := []Campaign{{
		BiddingScheme: ManualCpc,
		BudgetType:    "DailyBudgetStandard",
		DailyBudget:   25,
		Description:   "a new campaign",
		Name:          "newcapaign",
		Status:        "Active",
		TimeZone:      "EasternTimeUSCanada",
		CampaignType:  Shopping,
		Settings: []CampaignSettings{{
			Type:             "ShoppingSetting",
			Priority:         0,
			SalesCountryCode: "US",
			StoreId:          1397151,
		}},
	}}

	ids, err := svc.AddCampaigns(toadd)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(ids)
}

//addcampaignerror
//<AddCampaignsResponse xmlns="https://bingads.microsoft.com/CampaignManagement/v11"><CampaignIds xmlns:a="http://schemas.datacontract.org/2004/07/System" xmlns:i="http://www.w3.org/2001/XMLSchema-instance"><a:long i:nil="true"/></CampaignIds><PartialErrors xmlns:i="http://www.w3.org/2001/XMLSchema-instance"><BatchError><Code>1154</Code><Details i:nil="true"/><ErrorCode>CampaignServiceCampaignShoppingCampaignStoreIdInvalid</ErrorCode><FieldPath i:nil="true"/><ForwardCompatibilityMap i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/System.Collections.Generic"/><Index>0</Index><Message>The store ID of the shopping campaign is invalid.</Message><Type>BatchError</Type></BatchError></PartialErrors></AddCampaignsResponse>

func TestUnmarshalCampaigns(t *testing.T) {
	client := StringClient(envheader + `<GetCampaignsByAccountIdResponse xmlns="https://bingads.microsoft.com/CampaignManagement/v11"><Campaigns xmlns:i="http://www.w3.org/2001/XMLSchema-instance"><Campaign><BiddingScheme i:type="ManualCpcBiddingScheme"><Type>ManualCpc</Type></BiddingScheme><BudgetType>DailyBudgetStandard</BudgetType><DailyBudget>25</DailyBudget><Description>dota2</Description><ForwardCompatibilityMap xmlns:a="http://schemas.datacontract.org/2004/07/System.Collections.Generic"/><Id>804002264</Id><Name>dota2</Name><NativeBidAdjustment i:nil="true"/><Status>Active</Status><TimeZone>EasternTimeUSCanada</TimeZone><TrackingUrlTemplate i:nil="true"/><UrlCustomParameters i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/><CampaignType>Shopping</CampaignType><Settings><Setting i:type="ShoppingSetting"><Type>ShoppingSetting</Type><LocalInventoryAdsEnabled i:nil="true"/><Priority>0</Priority><SalesCountryCode>US</SalesCountryCode><StoreId>1387210</StoreId></Setting></Settings><BudgetId i:nil="true"/><Languages i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/></Campaign></Campaigns></GetCampaignsByAccountIdResponse>` + envend)

	svc := &CampaignService{
		Endpoint: "https://campaign.api.sandbox.bingads.microsoft.com/Api/Advertiser/CampaignManagement/v11/CampaignManagementService.svc",
		Session:  &Session{HTTPClient: client},
	}

	res, err := svc.GetCampaignsByAccountId(Shopping)

	if err != nil {
		t.Fatal(err)
	}

	expected := []Campaign{{
		BiddingScheme: BiddingScheme("ManualCpc"),
		BudgetType:    "DailyBudgetStandard",
		DailyBudget:   25,
		Description:   "dota2",
		Id:            804002264,
		Name:          "dota2",
		Status:        "Active",
		TimeZone:      "EasternTimeUSCanada",
		CampaignType:  Shopping,
		Settings: []CampaignSettings{{
			Type:             "ShoppingSetting",
			Priority:         0,
			SalesCountryCode: "US",
			StoreId:          1387210,
		}},
	}}

	if !reflect.DeepEqual(res, expected) {
		t.Errorf("expected %v, got %v", expected, res)
	}
}

func TestGetSandboxAdGroups(t *testing.T) {
	client := getTestClient()
	res, err := client.GetAdgroupsByCampaign(804002264)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}

func TestGetSandboxCriterion(t *testing.T) {
	client := getTestClient()
	res, err := client.GetAdGroupCriterionsByIds(1159984767305062)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}

func TestUnmarshalAdgroupCriterions(t *testing.T) {
	s := StringClient(envheader + `<GetAdGroupCriterionsByIdsResponse xmlns="https://bingads.microsoft.com/CampaignManagement/v11"><AdGroupCriterions xmlns:i="http://www.w3.org/2001/XMLSchema-instance"><AdGroupCriterion i:type="BiddableAdGroupCriterion"><AdGroupId>1159984767305062</AdGroupId><Criterion i:type="ProductPartition"><Type>ProductPartition</Type><Condition><Attribute/><Operand>All</Operand></Condition><ParentCriterionId i:nil="true"/><PartitionType>Subdivision</PartitionType></Criterion><Id>4576098675327012</Id><Status>Active</Status><Type>BiddableAdGroupCriterion</Type><CriterionBid i:nil="true"/><DestinationUrl i:nil="true"/><EditorialStatus i:nil="true"/><FinalAppUrls i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/><FinalMobileUrls i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/><FinalUrls i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/><TrackingUrlTemplate i:nil="true"/><UrlCustomParameters i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/></AdGroupCriterion><AdGroupCriterion i:type="BiddableAdGroupCriterion"><AdGroupId>1159984767305062</AdGroupId><Criterion i:type="ProductPartition"><Type>ProductPartition</Type><Condition><Attribute>agi</Attribute><Operand>ProductType1</Operand></Condition><ParentCriterionId>4576098675327012</ParentCriterionId><PartitionType>Unit</PartitionType></Criterion><Id>4576098675327013</Id><Status>Active</Status><Type>BiddableAdGroupCriterion</Type><CriterionBid i:type="FixedBid"><Type>FixedBid</Type><Amount>1</Amount></CriterionBid><DestinationUrl i:nil="true"/><EditorialStatus i:nil="true"/><FinalAppUrls i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/><FinalMobileUrls i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/><FinalUrls i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/><TrackingUrlTemplate i:nil="true"/><UrlCustomParameters i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/></AdGroupCriterion><AdGroupCriterion i:type="BiddableAdGroupCriterion"><AdGroupId>1159984767305062</AdGroupId><Criterion i:type="ProductPartition"><Type>ProductPartition</Type><Condition><Attribute>int</Attribute><Operand>ProductType1</Operand></Condition><ParentCriterionId>4576098675327012</ParentCriterionId><PartitionType>Unit</PartitionType></Criterion><Id>4576098675327014</Id><Status>Active</Status><Type>BiddableAdGroupCriterion</Type><CriterionBid i:type="FixedBid"><Type>FixedBid</Type><Amount>1</Amount></CriterionBid><DestinationUrl i:nil="true"/><EditorialStatus i:nil="true"/><FinalAppUrls i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/><FinalMobileUrls i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/><FinalUrls i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/><TrackingUrlTemplate i:nil="true"/><UrlCustomParameters i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/></AdGroupCriterion><AdGroupCriterion i:type="BiddableAdGroupCriterion"><AdGroupId>1159984767305062</AdGroupId><Criterion i:type="ProductPartition"><Type>ProductPartition</Type><Condition><Attribute>str</Attribute><Operand>ProductType1</Operand></Condition><ParentCriterionId>4576098675327012</ParentCriterionId><PartitionType>Unit</PartitionType></Criterion><Id>4576098675327015</Id><Status>Active</Status><Type>BiddableAdGroupCriterion</Type><CriterionBid i:type="FixedBid"><Type>FixedBid</Type><Amount>1</Amount></CriterionBid><DestinationUrl i:nil="true"/><EditorialStatus i:nil="true"/><FinalAppUrls i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/><FinalMobileUrls i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/><FinalUrls i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/><TrackingUrlTemplate i:nil="true"/><UrlCustomParameters i:nil="true" xmlns:a="http:/  chemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/></AdGroupCriterion><AdGroupCriterion i:type="BiddableAdGroupCriterion"><AdGroupId>1159984767305062</AdGroupId><Criterion i:type="ProductPartition"><Type>ProductPartition</Type><Condition><Attribute/><Operand>ProductType1</Operand></Condition><ParentCriterionId>4576098675327012</ParentCriterionId><PartitionType>Unit</PartitionType></Criterion><Id>4576098675327016</Id><Status>Active</Status><Type>BiddableAdGroupCriterion</Type><CriterionBid i:type="FixedBid"><Type>FixedBid</Type><Amount>1</Amount></CriterionBid><DestinationUrl i:nil="true"/><EditorialStatus i:nil="true"/><FinalAppUrls i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/><FinalMobileUrls i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/><FinalUrls i:nil="true" xmlns:a="http://schemas.microsoft.com/2003/10/Serialization/Arrays"/><TrackingUrlTemplate i:nil="true"/><UrlCustomParameters i:nil="true" xmlns:a="http://schemas.datacontract.org/2004/07/Microsoft.AdCenter.Advertiser.CampaignManagement.Api.DataContracts.V11"/></AdGroupCriterion></AdGroupCriterions></GetAdGroupCriterionsByIdsResponse>` + envend)

	svc := &CampaignService{
		Endpoint: "https://campaign.api.sandbox.bingads.microsoft.com/Api/Advertiser/CampaignManagement/v11/CampaignManagementService.svc",
		Session:  &Session{HTTPClient: s},
	}

	res, err := svc.GetAdGroupCriterionsByIds(1)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)
}
