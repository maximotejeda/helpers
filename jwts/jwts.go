// PAckage intended to make easier to work with Json Web Tokens

package jwts

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT struct for the JWT file
type JWT struct {
	privateKey    *rsa.PrivateKey
	publicKey     *rsa.PublicKey
	privatePemStr string
	PublicPemStr  string
}

var (
	actualDir, _  = os.Getwd()
	keys          = os.Getenv("KEYSDIR")
	priv          = os.Getenv("PRIVATEKEYNAME")
	pub           = os.Getenv("PUBLICKEYNAME")
	PrivateKeyDir = path.Join(actualDir, keys, priv)
	publicKeyDir  = path.Join(actualDir, keys, pub)
	tokenTTL      = os.Getenv("TOKENTTLS")
)

func NewJWT() *JWT {
	priv, pub := MyGenerateKeys()
	return &JWT{
		privateKey: priv,
		publicKey:  pub,
	}
}

// New Create a new instance of JWT
func (j *JWT) New() {
	if j.privateKey == nil {
		j.Renew()
	}
}

// Renew rotate the keys creating new keys in case of existence
func (j *JWT) Renew() {
	priv, pub := MyGenerateKeys()
	j.privateKey = priv
	j.publicKey = pub
	// need to try to convert to string
	j.privatePemStr = exportRSAPrivateKeyAsPemStr(priv)
	j.PublicPemStr, _ = exportRSAPublicKeyAsPemStr(pub)
	j.writeToDisk()
	//fmt.Printf("\nPrivatekeyPem: %s\n\n PublicKeyPem: %s", j.privatePemStr, j.publicPemStr)
}

// writeToDisk every time new keys are issued write them to disk overwriting the actuals
func (j *JWT) writeToDisk() {
	err := os.Mkdir("keys", 0770)
	if err != nil {
		fmt.Printf("\n%v\n", err)
	}

	err = os.WriteFile(publicKeyDir, []byte(j.PublicPemStr), 0644)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	err = os.WriteFile(PrivateKeyDir, []byte(j.privatePemStr), 0644)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

// ReadFromDisk Read keys from pem saved to disk
func (j *JWT) ReadFromDisk() {
	privateBytes, err := os.ReadFile(PrivateKeyDir)
	if err != nil {
		panic(err)
	}
	publicBytes, err := os.ReadFile(publicKeyDir)
	if err != nil {
		panic(err)
	}
	// assig string to J
	// mostly i intend to use keys as a way of replicate the service throug multiples pods
	// share the same keys with multiples pods if needed to replicate
	j.privatePemStr = string(privateBytes)
	j.PublicPemStr = string(publicBytes)
	j.privateKey, err = parsePrivateKeyFromPemStr(j.privatePemStr)
	if err != nil {
		panic(err)
	}
	j.publicKey, err = ParsePublicKeyFromPemStr(j.PublicPemStr)
	if err != nil {
		panic(err)
	}
}

func (j *JWT) Create(content interface{}) (token string, err error) {
	if j == nil || j.privateKey == nil {
		log.Print("nil struct pointer")
		return "", fmt.Errorf("nil pointer struct %v", j)
	}
	tokenTimeToLiveInt, err := strconv.Atoi(tokenTTL)
	if err != nil {
		tokenTimeToLiveInt = 900
	}
	tokenTimeToLive := time.Second * time.Duration(tokenTimeToLiveInt)
	now := time.Now().UTC()

	claims := make(jwt.MapClaims)

	claims["dat"] = content                         // Data we expect to have on JWT
	claims["exp"] = now.Add(tokenTimeToLive).Unix() // Expiration time after which the token ios invalid
	claims["iat"] = now.Unix()                      // The time at wwhen the token was issued
	claims["nbf"] = now.Unix()                      // The time before which the token must be disregarded.

	token, err = jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(j.privateKey)
	if err != nil {
		log.Printf("creating token %s", err.Error())
	}

	return token, nil
}

// Validate token in the algorithm used
// working as expected
// return claims["dat"]
func (j *JWT) Validate(token string) (map[string]interface{}, error) {
	if j == nil || j.privateKey == nil {
		log.Print("nil struct pointer")
		return nil, fmt.Errorf("nil pointer struct %v", j)
	}

	tok, err := jwt.Parse(token, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
		}
		return j.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("Validate parsing: %w", err)
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return nil, fmt.Errorf("validate: invalid")
	}
	dat, ok := claims["dat"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}, nil
	}
	return dat, nil
}

// RefreshToken: check header token validates it and check if is expired
// if expiration date is less than x issue a new
// Todo Set time for token and expiration timeout
func (j *JWT) RefreshToken(tokenStr string) (string, error) {
	_, err2 := j.Validate(tokenStr)
	var expNum int64
	switch {

	case err2 != nil && errors.Is(err2, jwt.ErrTokenExpired):
		token, err1 := jwt.Parse(tokenStr, nil)
		if token == nil {
			return "", err1
		}
		claims, _ := token.Claims.(jwt.MapClaims)
		// When the token expired?
		exp := claims["exp"]
		switch t := exp.(type) {
		case int:
			expNum = int64(t)
		case int64:
			expNum = t
		case float64:
			expNum = int64(t)
		default:
			fmt.Printf("The value is %v", t)
		}

		now := time.Now().Unix()
		tokenTTLInt, _ := strconv.Atoi(tokenTTL)
		if tokenTTLInt <= 0 {
			tokenTTLInt = 60
		}
		if (now-expNum) <= int64(tokenTTLInt) && (now-expNum) > 0 {
			//TODO modify the durarions
			newToken, err := j.Create(claims["dat"])
			if err != nil {
				return "", err
			} else {
				return newToken, err
			}
		} else {
			return "", fmt.Errorf("need new token session expired long")
		}
	case err2 != nil && !strings.Contains(err2.Error(), "Token is expired"):
		return "", err2
	}
	fmt.Println("Token still valid needs to wait for expiration")
	return tokenStr, nil
}

// here starts keys disk management
// function to generate key pairs from random 4096 bits
func MyGenerateKeys() (priv *rsa.PrivateKey, pub *rsa.PublicKey) {

	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		fmt.Printf("%v", err)
	}
	//fmt.Print(priv)
	pub = &priv.PublicKey

	priv.Validate()
	return priv, pub

}

// Parse private keys from pem
func parsePrivateKeyFromPemStr(privPEM string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block contianing key")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

// Encode the key to be able to save to disk and reuse or expor to configmap
func exportRSAPrivateKeyAsPemStr(privKey *rsa.PrivateKey) string {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKey_pem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privKeyBytes,
		},
	)
	return string(privKey_pem)
}

// Parse private keys from pem
func ParsePublicKeyFromPemStr(pubPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block from public key ")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break
	}
	return nil, errors.New("key type is not rsa")
}

// Encode the key to be able to save to disk and reuse or expor to configmap
func exportRSAPublicKeyAsPemStr(pubKey *rsa.PublicKey) (string, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return "", err
	}
	pubKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubKeyBytes,
		},
	)
	return string(pubKeyPem), nil
}
