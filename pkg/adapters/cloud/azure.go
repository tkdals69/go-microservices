package cloud

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type AzureAdapter struct {
	blobServiceClient *azblob.ServiceClient
}

func NewAzureAdapter() (*AzureAdapter, error) {
	azureAccountName := os.Getenv("AZURE_ACCOUNT_NAME")
	azureAccountKey := os.Getenv("AZURE_ACCOUNT_KEY")

	if azureAccountName == "" || azureAccountKey == "" {
		return nil, fmt.Errorf("Azure account name and key must be set in environment variables")
	}

	credential, err := azblob.NewSharedKeyCredential(azureAccountName, azureAccountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %v", err)
	}

	blobServiceClient, err := azblob.NewServiceClientWithSharedKey(azureAccountName, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob service client: %v", err)
	}

	return &AzureAdapter{
		blobServiceClient: blobServiceClient,
	}, nil
}

func (a *AzureAdapter) UploadBlob(ctx context.Context, containerName, blobName string, data []byte) error {
	containerClient := a.blobServiceClient.NewContainerClient(containerName)
	blobClient := containerClient.NewBlobClient(blobName)

	_, err := blobClient.Upload(ctx, azblob.NewStreamFromBytes(data), nil)
	if err != nil {
		return fmt.Errorf("failed to upload blob: %v", err)
	}

	return nil
}

func (a *AzureAdapter) DownloadBlob(ctx context.Context, containerName, blobName string) ([]byte, error) {
	containerClient := a.blobServiceClient.NewContainerClient(containerName)
	blobClient := containerClient.NewBlobClient(blobName)

	resp, err := blobClient.Download(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read blob body: %v", err)
	}

	return body, nil
}