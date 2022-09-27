package proxmox

import (
	"crypto/tls"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Login(t *testing.T) {
	client, err := NewClient(os.Getenv("PM_API_URL"), nil, &tls.Config{InsecureSkipVerify: true}, "", 300)
	assert.Nil(t, err)

	err = client.Login(os.Getenv("PM_USER"), os.Getenv("PM_PASS"), "")
	assert.Nil(t, err)
}
