package main

import (
	"fmt"
	"errors"
	"encoding/pem"
        "strings"
        "crypto"
	"crypto/tls"
        "crypto/x509"
        "crypto/x509/pkix"
	"crypto/rsa"
	"crypto/ecdsa"
)

var oid = map[string]string{
    "2.5.4.3":                    "CN",
    "2.5.4.4":                    "SN",
    "2.5.4.5":                    "serialNumber",
    "2.5.4.6":                    "C",
    "2.5.4.7":                    "L",
    "2.5.4.8":                    "ST",
    "2.5.4.9":                    "streetAddress",
    "2.5.4.10":                   "O",
    "2.5.4.11":                   "OU",
    "2.5.4.12":                   "title",
    "2.5.4.17":                   "postalCode",
    "2.5.4.42":                   "GN",
    "2.5.4.43":                   "initials",
    "2.5.4.44":                   "generationQualifier",
    "2.5.4.46":                   "dnQualifier",
    "2.5.4.65":                   "pseudonym",
    "0.9.2342.19200300.100.1.25": "DC",
    "1.2.840.113549.1.9.1":       "emailAddress",
    "0.9.2342.19200300.100.1.1":  "userid",
}

func getDNFromCert(namespace pkix.Name, sep string) (string, string, error) {
    subject := []string{}
    var commonName string
    for _, s := range namespace.ToRDNSequence() {
        for _, i := range s {
            if v, ok := i.Value.(string); ok {
                if name, ok := oid[i.Type.String()]; ok {
                    // <oid name>=<value>
                    subject = append(subject, fmt.Sprintf("%s=%s", name, v))
                    if name == "CN" {
                      commonName = v
                    }
                } else {
                    // <oid>=<value> if no <oid name> is found
                    subject = append(subject, fmt.Sprintf("%s=%s", i.Type.String(), v))
                }
            } else {
                // <oid>=<value in default format> if value is not string
                subject = append(subject, fmt.Sprintf("%s=%v", i.Type.String, v))
            }
        }
    }
    return sep + strings.Join(subject, sep), commonName, nil
}

func X509KeyPair(certPEMBlock, keyPEMBlock []byte) (cert tls.Certificate, cn string, err error) {
	fmt.Println(certPEMBlock)
  var certDERBlock *pem.Block
  for {
    certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
    if certDERBlock == nil {
	    fmt.Println("certDERBlock is nil")
      break
    }
    if certDERBlock.Type == "CERTIFICATE" {
	    fmt.Println("found one...")
      cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
    }
  }

  if len(cert.Certificate) == 0 {
    err = errors.New("crypto/tls: failed to parse certificate PEM data")
    return
  }

  var keyDERBlock *pem.Block
  for {
    keyDERBlock, keyPEMBlock = pem.Decode(keyPEMBlock)
    if keyDERBlock == nil {
      err = errors.New("crypto/tls: failed to parse key PEM data")
      return
    }

    // we don't support a cryptic certificate right now
		// if x509.IsEncryptedPEMBlock(keyDERBlock) {
  //      out, err2 := x509.DecryptPEMBlock(keyDERBlock, pw)
		//  	 if err2 != nil {
		//  		  err = err2
		//  		  return
		//  	 }
  //      keyDERBlock.Bytes = out
  //      break
  //   }
    if keyDERBlock.Type == "PRIVATE KEY" || strings.HasSuffix(keyDERBlock.Type, " PRIVATE KEY") {
      break
    }
  }

  cert.PrivateKey, err = parsePrivateKey(keyDERBlock.Bytes)
  if err != nil {
    return
  }
  // We don't need to parse the public key for TLS, but we so do anyway
  // to check that it looks sane and matches the private key.
  x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
  if err != nil {
    return
  }

  _, cn, err = getDNFromCert(x509Cert.Subject, "/")
  if err != nil {
    fmt.Println("err")
  }
  return
}

func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
  if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
    return key, nil
  }
  if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
    switch key := key.(type) {
    case *rsa.PrivateKey, *ecdsa.PrivateKey:
      return key, nil
    default:
      return nil, errors.New("crypto/tls: found unknown private key type in PKCS#8 wrapping")
    }
  }
  if key, err := x509.ParseECPrivateKey(der); err == nil {
    return key, nil
  }

  return nil, errors.New("crypto/tls: failed to parse private key")
}

func parseCert(certFile, keyFile string) (cert tls.Certificate, cn string, err error){
  certPEMBlock := []byte(certFile)
  keyPEMBlock := []byte(keyFile)
  return X509KeyPair(certPEMBlock, keyPEMBlock)
}
