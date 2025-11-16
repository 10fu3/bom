package bom

import (
    "crypto/rand"
    "fmt"
    "github.com/google/uuid"
    "github.com/lucsky/cuid"
    "github.com/oklog/ulid/v2"
    "sync"
    "time"
)

// Supported identity strategies.
const (
    IdentityUUIDv4 = "uuidv4"
    IdentityUUIDv7 = "uuidv7"
    IdentityULID   = "ulid"
    IdentityCUID   = "cuid"
)

// IdentityGenerator describes an ID generator injected into create helpers.
type IdentityGenerator interface {
    Generate(strategy string) (string, error)
}

// DefaultIdentityGenerator implements IdentityGenerator using std libs.
type DefaultIdentityGenerator struct {
    mu          sync.Mutex
    cuidCounter uint32
}

// Generate creates an identifier string based on the requested strategy.
func (g *DefaultIdentityGenerator) Generate(strategy string) (string, error) {
    switch strategy {
    case IdentityUUIDv4:
        return uuid.New().String(), nil
    case IdentityUUIDv7:
        id, err := uuid.NewV7()
        if err != nil {
            return "", err
        }
        return id.String(), nil
    case IdentityULID:
        return generateULID()
    case IdentityCUID:
        return g.generateCUID()
    default:
        return "", fmt.Errorf("unsupported identity strategy %q", strategy)
    }
}

func generateULID() (string, error) {
    t := time.Now()
    entropy := ulid.Monotonic(rand.Reader, 0)
    id, err := ulid.New(ulid.Timestamp(t), entropy)

    if err != nil {
        return "", nil
    }

    return id.String(), nil
}

func (g *DefaultIdentityGenerator) generateCUID() (string, error) {
    c, err := cuid.NewCrypto(rand.Reader)
    if err != nil {
        return "", err
    }
    return c, nil
}
