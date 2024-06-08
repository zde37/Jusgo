package database

import (
	"context" 
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/require"
)

func TestConnectToMongoDB(t *testing.T) {
	err := godotenv.Load("../../.env")
	require.NoError(t,err) 

	_, _, err = ConnectToMongoDB(os.Getenv("DB_SOURCE"), context.Background())
	require.NoError(t, err)
}
