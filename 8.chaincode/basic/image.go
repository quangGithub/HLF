package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/flogging"
)

type SmartContract struct {
	contractapi.Contract
}

var logger = flogging.MustGetLogger("image")

type Image struct {
	DocType           string    `json:"docType"`
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	ImageURL          string    `json:"imageURL"`
	Description       string    `json:"description"`
	Owner             string    `json:"owner"`
	StatusImage       string    `json:"statusImage"`
	AddedBy           string    `json:"addedBy"`
	AddedAt           time.Time `json:"addedAt"`
	Behavior          string    `json:"Behavior"`
	StatusTransaction string    `json:"statusTransaction"`
}
type Transaction struct {
	DocType    string    `json:"docType"`
	ID         string    `json:"id"`
	Buyer      string    `json:"buyer"`
	Seller     string    `json:"seller"`
	ImageID    string    `json:"image_id"`
	Status     string    `json:"status"`
	Created_at time.Time `json:"created_at"`
	Created_by string    `json:"created_by"`
	Updated_at time.Time `json:"updated_at"`
	Updated_by string    `json:"updated_by"`
}

func (s *SmartContract) CreateImage(ctx contractapi.TransactionContextInterface, imageID string, imageData string, username string) (string, error) {
	if len(imageData) == 0 {
		return "", fmt.Errorf("Please pass the correct image data")
	}
	var image Image
	err := json.Unmarshal([]byte(imageData), &image)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling image. %s", err.Error())
	}
	image.ID = imageID
	image.DocType = "image"
	image.StatusImage = "Not Listing"
	image.AddedBy = username
	image.AddedAt = time.Unix(time.Now().Unix(), 0)
	image.Owner = username
	image.Behavior = "Create image"
	image.StatusTransaction = ""
	imageAsBytes, err := json.Marshal(image)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling image. %s", err.Error())
	}
	ctx.GetStub().SetEvent("CreateAsset", imageAsBytes)
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(image.ID, imageAsBytes)
}

//create Transaction
func (s *SmartContract) CreateTransaction(ctx contractapi.TransactionContextInterface, transactionID string, transactionData string, username string) (string, error) {
	if len(transactionData) == 0 {
		return "", fmt.Errorf("Please pass the correct transaction data")
	}
	var transaction Transaction
	err := json.Unmarshal([]byte(transactionData), &transaction)
	if err != nil {
		return "", fmt.Errorf("Failed while unmarshling transaction. %s", err.Error())
	}
	transaction.DocType = "transaction"
	transaction.ID = transactionID
	transaction.Created_at = time.Unix(time.Now().Unix(), 0)
	transaction.Created_by = username
	transaction.Status = "WaitForSeller"
	transaction.Buyer = username
	var imageID = transaction.ImageID
	image, err := s.GetImageById(ctx, imageID, "")
	if err != nil {
		return "", fmt.Errorf("Failed while get image by id. %s", err.Error())
	}
	transaction.Seller = image.Owner
	if username == image.Owner {
		return "", fmt.Errorf("Failed while username equal to owner image. %s", err.Error())
	}
	if image.StatusImage == "Listing" {
		transactionAsBytes, err := json.Marshal(transaction)
		if err != nil {
			return "", fmt.Errorf("Failed while marshling transaction. %s", err.Error())
		}
		ctx.GetStub().SetEvent("CreateTransaction", transactionAsBytes)
		behavior := "Create transaction"
		statusTransaction := "Wait for seller"
		s.UpdateImageBehaviour(ctx, transaction.ImageID, username, behavior, statusTransaction)
		return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(transaction.ID, transactionAsBytes)
	} else {
		return "", fmt.Errorf("Image not listing on marketplace!")
	}
}

// Update Transaction
func (s *SmartContract) UpdateTransaction(ctx contractapi.TransactionContextInterface, transactionID string, status string, username string) (string, error) {
	if len(transactionID) == 0 {
		return "", fmt.Errorf("Please pass the correct transaction data")
	}
	transactionAsBytes, err := ctx.GetStub().GetState(transactionID)
	if err != nil {
		return "", fmt.Errorf("Failed to get transaction data. %s", err.Error())
	}
	transaction := new(Transaction)
	_ = json.Unmarshal(transactionAsBytes, transaction)
	if err != nil {
		return "", fmt.Errorf("Failed to get image data. %s", err.Error())
	}
	if transactionAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", transactionID)
	}
	if status == "Cancel" && username == transaction.Buyer && transaction.Status == "WaitForSeller" {
		transaction.Status = "BuyerCancel"
		transaction.Updated_by = transaction.Buyer
		transaction.Updated_at = time.Unix(time.Now().Unix(), 0)
		behavior := "Buyer cancel"
		statusTransaction := "Buyer cancel"
		s.UpdateImageBehaviour(ctx, transaction.ImageID, username, behavior, statusTransaction)
	} else if status == "Cancel" && username == transaction.Seller && transaction.Status == "WaitForSeller" {
		transaction.Status = "SellerCancel"
		transaction.Updated_by = transaction.Seller
		transaction.Updated_at = time.Unix(time.Now().Unix(), 0)
		behavior := "SellerCancel"
		statusTransaction := "Seller cancel"
		s.UpdateImageBehaviour(ctx, transaction.ImageID, username, behavior, statusTransaction)
	} else if status == "SellerConfirmed" && username == transaction.Seller && transaction.Status == "WaitForSeller" {
		transaction.Status = "Transfer Successs"
		transaction.Updated_by = transaction.Seller
		transaction.Updated_at = time.Unix(time.Now().Local().Unix(), 0)
		behavior := "Transfer success"
		s.UpdateImageOwner(ctx, transaction.ImageID, transaction.Buyer, transaction.Seller, behavior)
	} else {
		return "", fmt.Errorf("Failed update transaction, input not appropriate. %s", err.Error())
	}
	transactionAsBytes, err = json.Marshal(transaction)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling transaction. %s", err.Error())
	}
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(transaction.ID, transactionAsBytes)
}
func (s *SmartContract) UpdateImageOwner(ctx contractapi.TransactionContextInterface, imageID string, newOwner string, oldOwner string, behavior string) (string, error) {
	if len(imageID) == 0 {
		return "", fmt.Errorf("Please pass the correct image id")
	}
	imageAsBytes, err := ctx.GetStub().GetState(imageID)
	if err != nil {
		return "", fmt.Errorf("Failed to get image data. %s", err.Error())
	}
	if imageAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", imageID)
	}
	image := new(Image)
	_ = json.Unmarshal(imageAsBytes, image)
	image.Owner = newOwner
	image.AddedBy = oldOwner
	image.AddedAt = time.Unix(time.Now().Unix(), 0)
	image.StatusImage = "Not Listing"
	image.Behavior = behavior
	image.StatusTransaction = "Seller transfer successful"
	imageAsBytes, err = json.Marshal(image)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling image. %s", err.Error())
	}
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(image.ID, imageAsBytes)
}
func (s *SmartContract) UpdateImageStatusListingMarketplace(ctx contractapi.TransactionContextInterface, imageID string, status string, owner string) (string, error) {
	if len(imageID) == 0 {
		return "", fmt.Errorf("Please pass the correct image id")
	}
	imageAsBytes, err := ctx.GetStub().GetState(imageID)
	if err != nil {
		return "", fmt.Errorf("Failed to get image data. %s", err.Error())
	}
	if imageAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", imageID)
	}
	getOwnerImage, err := s.GetImageById(ctx, imageID, owner)
	if err != nil {
		return "", fmt.Errorf("Failed while get image by id. %s", err.Error())
	}
	if owner != getOwnerImage.Owner {
		return "", fmt.Errorf("Failed while get image by owner. %s", err.Error())
	}
	image := new(Image)
	_ = json.Unmarshal(imageAsBytes, image)
	if status == "Listing" {
		image.StatusImage = "Listing"
		image.AddedBy = owner
		image.AddedAt = time.Unix(time.Now().Unix(), 0)
		image.Behavior = "Listing"
	} else if status == "Cancel" {
		image.StatusImage = "Not Listing"
		image.Behavior = "Delisting"
	} else {
		return "", fmt.Errorf("Failed update image status listing marketplace, input not appropriate. %s", err.Error())
	}
	imageAsBytes, err = json.Marshal(image)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling image. %s", err.Error())
	}
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(image.ID, imageAsBytes)
}
func (s *SmartContract) UpdateImageBehaviour(ctx contractapi.TransactionContextInterface, imageID string, username string, behavior string, statusTransaction string) (string, error) {
	if len(imageID) == 0 {
		return "", fmt.Errorf("Please pass the correct image id")
	}
	imageAsBytes, err := ctx.GetStub().GetState(imageID)
	if err != nil {
		return "", fmt.Errorf("Failed to get image data. %s", err.Error())
	}
	if imageAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", imageID)
	}
	image := new(Image)
	_ = json.Unmarshal(imageAsBytes, image)

	image.AddedAt = time.Unix(time.Now().Unix(), 0)
	image.AddedBy = username
	image.Behavior = behavior
	image.StatusTransaction = statusTransaction
	if statusTransaction == "Wait for seller" {
		image.StatusImage = "Buying"
	} else if statusTransaction == "Seller cancel" || statusTransaction == "Buyer cancel" {
		image.StatusImage = "Not Listing"
	}
	imageAsBytes, err = json.Marshal(image)
	if err != nil {
		return "", fmt.Errorf("Failed while marshling image. %s", err.Error())
	}
	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(image.ID, imageAsBytes)
}
func (s *SmartContract) GetHistoryForAsset(ctx contractapi.TransactionContextInterface, imageID string) (string, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(imageID)
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
func (s *SmartContract) GetImageById(ctx contractapi.TransactionContextInterface, imageID string, username string) (*Image, error) {
	if len(imageID) == 0 {
		return nil, fmt.Errorf("Please provide correct contract Id")
		// return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	imageAsBytes, err := ctx.GetStub().GetState(imageID)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}
	image := new(Image)
	_ = json.Unmarshal(imageAsBytes, image)
	if image.Owner != username && image.StatusImage != "Listing" {
		return nil, fmt.Errorf("%s does not exist", imageID)
	}
	return image, nil
}
func (s *SmartContract) DeleteImageById(ctx contractapi.TransactionContextInterface, imageID string, owner string) (string, error) {
	if len(imageID) == 0 {
		return "", fmt.Errorf("Please provide correct contract Id")
	}
	getOwnerImage, err := s.GetImageById(ctx, imageID, owner)
	if err != nil {
		return "", fmt.Errorf("Failed while get image by id. %s", err.Error())
	}
	if owner != getOwnerImage.Owner {
		return "", fmt.Errorf("Failed while get image by owner. %s", err.Error())
	}
	return ctx.GetStub().GetTxID(), ctx.GetStub().DelState(imageID)
}
func (s *SmartContract) GetAllImages(ctx contractapi.TransactionContextInterface) ([]*Image, error) {
	queryString := `{"selector":{"docType":"image"}}`
	return getQueryResultForQueryString(ctx, queryString)
}
func (s *SmartContract) GetAllImagesOnMarketplace(ctx contractapi.TransactionContextInterface) ([]*Image, error) {
	queryString := `{"selector":{"docType":"image", "statusImage":"Listing"}}`
	return getQueryResultForQueryString(ctx, queryString)
}
func (s *SmartContract) GetAllImagesOwner(ctx contractapi.TransactionContextInterface, owner string) ([]*Image, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"image", "owner":"%s"}}`, owner)
	return getQueryResultForQueryString(ctx, queryString)
}
func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*Image, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	return constructQueryResponseFromIterator(resultsIterator)
}
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Image, error) {
	var images []*Image
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var image Image
		err = json.Unmarshal(queryResult.Value, &image)
		if err != nil {
			return nil, err
		}
		images = append(images, &image)
	}
	return images, nil
}
func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create image chaincode: %s", err.Error())
		return
	}
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincodes: %s", err.Error())
	}
}
