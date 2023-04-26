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
	ID              string    `json:"id"`
	Username        string    `json:"username"`
	CCCDImageURL    string    `json:"cccdImageURL"`
	CCCDNo          string    `json:"cccdNo"`
	CMNDNo          string    `json:"cmndNo"`
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
	ID               string    `json:"id"`
	OwnerProfileId   string    `json:"ownerProfileId"`
	MatcherProfileId string    `json:"matcherProfileId"`
	QRImageURL       string    `json:"qrImageURL"`
	MatchingLocation string    `json:"matchingLocation"`
	MatchingTime     time.Time `json:"matchingTime"`
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
	profile.ID = profileID
	profile.CreatedAt = time.Unix(time.Now().Unix(), 0)
	profile.UpdatedAt = time.Unix(time.Now().Unix(), 0)
	profileAsBytes, err := json.Marshal(profile)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling profile. %s", err.Error())
	}
	ctx.GetStub().SetEvent("CreateAsset", profileAsBytes)
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(profile.ID, profileAsBytes)
}

// Update Profile
func (s *SmartContract) UpdateProfile(ctx contractapi.TransactionContextInterface, profileID string, profileData string) (string, error) {
	if len(profileID) == 0 {
		return "", fmt.Errorf("Please pass the correct profile id")
	}
	profileAsBytes, err := ctx.GetStub().GetState(profileID)
	if err != nil {
		return "", fmt.Errorf("Failed to get profile data. %s", err.Error())
	}
	if profileAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", profileID)
	}
	profile := new(Profile)
	err = json.Unmarshal(profileAsBytes, profile)
	if err != nil {
		return "", fmt.Errorf("Failed to unmarshal profile. %s", err.Error())
	}
	profile.UpdatedAt = time.Unix(time.Now().Unix(), 0)
	profileAsBytes, err = json.Marshal(profile)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling profile. %s", err.Error())
	}
	return ctx.GetStub().GetTxID(),
		ctx.GetStub().PutState(profile.ID, profileAsBytes)
}

// Create Match
func (s *SmartContract) CreateMatch(ctx contractapi.TransactionContextInterface, matchID string, matchData string, ownerProfileId string) (string, error) {
	if len(matchData) == 0 {
		return "", fmt.Errorf("Please pass the correct match data")
	}
	matchAsBytesCheck, _ := ctx.GetStub().GetState(matchID)
	if matchAsBytesCheck != nil {
		return "", fmt.Errorf("Mtach ID exist")
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
	match.ID = matchID
	match.OwnerProfileId = ownerProfileId
	match.Status = "Created"
	match.CreatedAt = time.Unix(time.Now().Unix(), 0)
	matchAsBytes, err := json.Marshal(match)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling match. %s", err.Error())
	}
	ctx.GetStub().SetEvent("CreateAsset", matchAsBytes)
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(match.ID, matchAsBytes)
}

// Accept Match
func (s *SmartContract) AcceptMatch(ctx contractapi.TransactionContextInterface, matchID string, userProfileId string) (string, error) {
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
		match.Status = "Confirmed"
		match.MatchingTime = time.Unix(time.Now().Unix(), 0)
		match.UpdatedAt = time.Unix(time.Now().Unix(), 0)
	} else {
		return "", fmt.Errorf("Failed update match, input not appropriate. %s", err.Error())
	}
	matchAsBytes, err = json.Marshal(match)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling match. %s", err.Error())
	}
	return ctx.GetStub().GetTxID(),
		ctx.GetStub().PutState(match.ID, matchAsBytes)
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

// Delete Profile by id
func (s *SmartContract) DeleteProfileById(ctx contractapi.TransactionContextInterface, profileID string) (string, error) {
	if len(profileID) == 0 {
		return "", fmt.Errorf("Please provide correct profile Id")
	}
	return ctx.GetStub().GetTxID(), ctx.GetStub().DelState(profileID)
}

// Get all profiles
func (s *SmartContract) GetAllProfiles(ctx contractapi.TransactionContextInterface) ([]*Profile, error) {
	queryString := `{"selector":{"docType":"profile"}}`
	return getQueryResultForQueryStringProfile(ctx, queryString)
}
func getQueryResultForQueryStringProfile(ctx contractapi.TransactionContextInterface, queryString string) ([]*Profile, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	return constructQueryResponseFromIteratorProfile(resultsIterator)
}
func constructQueryResponseFromIteratorProfile(resultsIterator shim.StateQueryIteratorInterface) ([]*Profile, error) {
	var profileList []*Profile
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var profile Profile
		err = json.Unmarshal(queryResult.Value, &profile)
		if err != nil {
			return nil, err
		}
		profileList = append(profileList, &profile)
	}
	return profileList, nil
}

// Get Match by owner and matcher
func (s *SmartContract) GetMatchByOwner(ctx contractapi.TransactionContextInterface, ownerID string) ([]*Match, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"match", "ownerProfileId":"%s"}}`, ownerID)
	return getQueryResultForQueryStringMatch(ctx, queryString)
}
func (s *SmartContract) GetMatchByMatcher(ctx contractapi.TransactionContextInterface, ownerID string) ([]*Match, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"match", "matcherProfileId":"%s"}}`, ownerID)
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
