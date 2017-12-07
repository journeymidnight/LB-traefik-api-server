package main

import (
	"fmt"
	"errors"
	"encoding/pem"
        "strings"
        "crypto"
	"crypto/tls"
        "crypto/x509"
	"crypto/rsa"
	"crypto/ecdsa"
)

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

  // Add check for SAN extension, and we currently not support it
  if x509Cert.DNSNames != nil || x509Cert.EmailAddresses != nil || 
      x509Cert.IPAddresses !=nil /*|| x509Cert.Extensions != nil*/ {
    err = errors.New("crypto/tls: found extra name(SAN)")
    return
  }

  if x509Cert != nil {
      cn = x509Cert.Subject.CommonName
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
