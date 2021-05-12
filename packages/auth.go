package backend

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
)

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

var GetToken map[string]interface{}

func Middleware() (*jwtmiddleware.JWTMiddleware, map[string]interface{}) {
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			// get token for database
			p := &GetToken
			*p = token.Claims.(jwt.MapClaims)

			// Verify 'aud' claim
			// 'aud' = audience
			aud := "https://infinite-ocean-99647.herokuapp.com/"

			// debugging ------
			// convert audience in the JWT token to []interface{} if multiple audiences
			convAud, ok := token.Claims.(jwt.MapClaims)["aud"].([]interface{})
			if !ok {
				// convert audience in the JWT token to string if only 1 audience
				strAud, ok := token.Claims.(jwt.MapClaims)["aud"].(string)
				// return error if can't convert to string
				if !ok {
					return token, errors.New("Invalid audience.")
				}
				// return error if audience doesn't match
				if strAud != aud {
					return token, errors.New("Invalid audience.")
				}
			} else {
				for _, v := range convAud {
					if v == aud {
						break
					} else {
						return token, errors.New("Invalid audience.")
					}
				}
			}

			// ------

			// this was the original way of verifying audience, wasn't working because it couldn't convert slice correctly

			// checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(aud, false)
			// if !checkAud {
			// 	return token, errors.New("Invalid audience.")
			// }

			// -----------------------

			// Verify 'iss' claim
			// 'iss' = issuer
			iss := "https://importhuman.us.auth0.com/"
			checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(iss, false)
			if !checkIss {
				return token, errors.New("Invalid issuer.")
			}

			cert, err := getPemCert(token)
			if err != nil {
				panic(err.Error())
			}

			result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
			return result, nil
		},
		SigningMethod: jwt.SigningMethodRS256,
	})
	return jwtMiddleware, GetToken
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	resp, err := http.Get("https://importhuman.us.auth0.com/.well-known/jwks.json")

	if err != nil {
		return cert, err
	}
	defer resp.Body.Close()

	var jwks = Jwks{}
	err = json.NewDecoder(resp.Body).Decode(&jwks)

	if err != nil {
		return cert, err
	}

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("Unable to find appropriate key.")
		return cert, err
	}

	return cert, nil
}
