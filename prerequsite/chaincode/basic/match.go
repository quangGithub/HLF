package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	// "runtime/internal/math"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/flogging"
)

type SmartContract struct {
	contractapi.Contract
}

var logger = flogging.MustGetLogger("match")

type Profile struct {
	DocType         string    `json:"docType"`
	ProfileID       string    `json:"profileId"`
	Username        string    `json:"username"`
	IDCardImageURL  string    `json:"idcardImageURL"`
	IDCardNo1       string    `json:"idcardNo1"`
	IDCardNo2       string    `json:"idcardNo2"`
	Fullname        string    `json:"fullname"`
	Gender          string    `json:"gender"`
	Birthdate       string    `json:"birthdate"`
	Address         string    `json:"address"`
	Hometown        string    `json:"hometown"`
	FingerprintCode string    `json:"fingerprintCode"`
	CardCode        string    `json:"cardCode"`
	Email           string    `json:"email"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Match struct {
	DocType          string    `json:"docType"`
	MatchID          string    `json:"matchId"`
	OwnerProfileId   string    `json:"ownerProfileId"`
	MatcherProfileId string    `json:"matcherProfileId"`
	MatchingLocation string    `json:"matchingLocation"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Create Profile
func (s *SmartContract) CreateProfile(ctx contractapi.TransactionContextInterface, profileID string, profileData string) (string, error) {
	if len(profileData) == 0 {
		return "", fmt.Errorf("Please pass the correct profile data")
	}
	profileAsBytesCheck, _ := ctx.GetStub().GetState(profileID)
	if profileAsBytesCheck != nil {
		return "", fmt.Errorf("Profile ID exist!")
	}
	var profile Profile
	err := json.Unmarshal([]byte(profileData), &profile)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling profile. %s", err.Error())
	}
	profile.DocType = "profile"
	profile.ProfileID = profileID
	profile.CreatedAt = time.Unix(time.Now().Unix(), 0)
	profile.UpdatedAt = time.Unix(time.Now().Unix(), 0)
	profileAsBytes, err := json.Marshal(profile)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling profile. %s", err.Error())
	}
	ctx.GetStub().SetEvent("CreateAsset", profileAsBytes)
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(profile.ProfileID, profileAsBytes)
}

// Create Match
func (s *SmartContract) CreateMatch(ctx contractapi.TransactionContextInterface, matchID string, matchData string) (string, error) {
	if len(matchData) == 0 {
		return "", fmt.Errorf("Please pass the correct match data")
	}
	matchAsBytesCheck, _ := ctx.GetStub().GetState(matchID)
	if matchAsBytesCheck != nil {
		return "", fmt.Errorf("Match ID exist")
	}
	var match Match
	err := json.Unmarshal([]byte(matchData), &match)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling match. %s", err.Error())
	}
	_, err = s.GetProfileById(ctx, match.MatcherProfileId)
	if err != nil {
		return "", fmt.Errorf("Fail to get matcher information.", err.Error())
	}
	match.DocType = "match"
	match.MatchID = matchID
	match.Status = "Created"
	match.CreatedAt = time.Unix(time.Now().Unix(), 0)
	matchAsBytes, err := json.Marshal(match)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling match. %s", err.Error())
	}
	ctx.GetStub().SetEvent("CreateAsset", matchAsBytes)
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(match.MatchID, matchAsBytes)
}

// Update Match
func (s *SmartContract) UpdateMatch(ctx contractapi.TransactionContextInterface, matchID string, userProfileId string, status string) (string, error) {
	if len(matchID) == 0 {
		return "", fmt.Errorf("Please pass the correct match id")
	}
	matchAsBytes, err := ctx.GetStub().GetState(matchID)
	if err != nil {
		return "", fmt.Errorf("Failed to get match data. %s", err.Error())
	}
	if matchAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", matchID)
	}
	match := new(Match)
	err = json.Unmarshal(matchAsBytes, match)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal match. %s", err.Error())
	}
	if userProfileId == match.MatcherProfileId {
		if status == "Agree" {
			match.Status = "Agree"
			match.UpdatedAt = time.Unix(time.Now().Unix(), 0)
		} else if status == "Cancel" {
			match.Status = "Cancel"
			match.UpdatedAt = time.Unix(time.Now().Unix(), 0)
		} else {
			return "", fmt.Errorf("Status invalid!")
		}
	} else {
		return "", fmt.Errorf("User's not matcher!")
	}
	matchAsBytes, err = json.Marshal(match)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling match. %s", err.Error())
	}
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(match.MatchID, matchAsBytes)
}

// Get history for Profile, Match
func (s *SmartContract) GetHistoryForAsset(ctx contractapi.TransactionContextInterface, assetID string) (string, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(assetID)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return "", fmt.Errorf(err.Error())
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Value\":")
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}
		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")
		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	return string(buffer.Bytes()), nil
}

// Get profile by Id
func (s *SmartContract) GetProfileById(ctx contractapi.TransactionContextInterface, profileID string) (*Profile, error) {
	if len(profileID) == 0 {
		return nil, fmt.Errorf("Please provide correct profile Id")
	}
	profileAsBytes, err := ctx.GetStub().GetState(profileID)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}
	profile := new(Profile)
	_ = json.Unmarshal(profileAsBytes, profile)
	return profile, nil
}

// Get match by Id
func (s *SmartContract) GetMatchById(ctx contractapi.TransactionContextInterface, matchID string) (*Match, error) {
	if len(matchID) == 0 {
		return nil, fmt.Errorf("Please provide correct profile Id")
	}
	matchAsBytes, err := ctx.GetStub().GetState(matchID)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}
	match := new(Match)
	_ = json.Unmarshal(matchAsBytes, match)
	return match, nil
}

// Get Match by owner and matcher
func (s *SmartContract) GetMatchsByProfile(ctx contractapi.TransactionContextInterface, profileID string) ([]*Match, error) {
	queryString := fmt.Sprintf(` { "selector":{ "docType":"match", "$or": [ { "ownerProfileId":"%s" }, { "matcherProfileId":"%s" } ] } }`, profileID, profileID)
	return getQueryResultForQueryStringMatch(ctx, queryString)
}
func getQueryResultForQueryStringMatch(ctx contractapi.TransactionContextInterface, queryString string) ([]*Match, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	return constructQueryResponseFromIteratorMatch(resultsIterator)
}
func constructQueryResponseFromIteratorMatch(resultsIterator shim.StateQueryIteratorInterface) ([]*Match, error) {
	var matchList []*Match
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var match Match
		err = json.Unmarshal(queryResult.Value, &match)
		if err != nil {
			return nil, err
		}
		matchList = append(matchList, &match)
	}
	return matchList, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create match chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincodes: %s", err.Error())
	}
}
