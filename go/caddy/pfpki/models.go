package pfpki

import (
	"bytes"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	// Import MySQL lib
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Digest Values:
// 0 UnknownSignatureAlgorithm
// 1 MD2WithRSA
// 2 MD5WithRSA
// 3 SHA1WithRSA
// 4 SHA256WithRSA
// 5 SHA384WithRSA
// 6 SHA512WithRSA
// 7 DSAWithSHA1
// 8 DSAWithSHA256
// 9 ECDSAWithSHA1
// 10 ECDSAWithSHA256
// 11 ECDSAWithSHA384
// 12 ECDSAWithSHA512
// 13 SHA256WithRSAPSS
// 14 SHA384WithRSAPSS
// 15 SHA512WithRSAPSS
// 16 PureEd25519

// KeyUsage Values:
// 1 KeyUsageDigitalSignature
// 2 KeyUsageContentCommitment
// 4 KeyUsageKeyEncipherment
// 8 KeyUsageDataEncipherment
// 16 KeyUsageKeyAgreement
// 32 KeyUsageCertSign
// 64 KeyUsageCRLSign
// 128 KeyUsageEncipherOnly
// 256 KeyUsageDecipherOnly

// ExtendedKeyUsage Values:
// 0 ExtKeyUsageAny
// 1 ExtKeyUsageServerAuth
// 2 ExtKeyUsageClientAuth
// 3 ExtKeyUsageCodeSigning
// 4 ExtKeyUsageEmailProtection
// 5 ExtKeyUsageIPSECEndSystem
// 6 ExtKeyUsageIPSECTunnel
// 7 ExtKeyUsageIPSECUser
// 8 ExtKeyUsageTimeStamping
// 9 ExtKeyUsageOCSPSigning
// 10 ExtKeyUsageMicrosoftServerGatedCrypto
// 11 ExtKeyUsageNetscapeServerGatedCrypto
// 12 ExtKeyUsageMicrosoftCommercialCodeSigning
// 13 ExtKeyUsageMicrosoftKernelCodeSigning

// CA struct
type CA struct {
	gorm.Model
	Cn                  string                  `json:"cn" gorm:"UNIQUE"`
	Mail                string                  `json:"mail"`
	Organisation        string                  `json:"organisation"`
	Country             string                  `json:"country"`
	State               string                  `json:"state"`
	Locality            string                  `json:"locality"`
	StreetAddress       string                  `json:"streetaddress"`
	PostalCode          string                  `json:"postalcode"`
	KeyType             Type                    `json:"keytype"`
	KeySize             int                     `json:"keysize"`
	Digest              x509.SignatureAlgorithm `json:"digest"`
	KeyUsage            string                  `json:"keyusage,omitempty"`
	ExtendedKeyUsage    string                  `json:"extendedkeyusage,omitempty"`
	Days                int                     `json:"days"`
	CaKey               string                  `json:"cakey,omitempty" gorm:"type:longtext"`
	CaCert              string                  `json:"cacert,omitempty" gorm:"type:longtext"`
	IssuerKeyHashmd5    string                  `json:"issuerkeyhashmd5,omitempty" gorm:"UNIQUE_INDEX"`
	IssuerKeyHashsha1   string                  `json:"issuerkeyhashsha1,omitempty" gorm:"UNIQUE_INDEX"`
	IssuerKeyHashsha256 string                  `json:"issuerkeyhashsha256,omitempty" gorm:"UNIQUE_INDEX"`
	IssuerKeyHashsha512 string                  `json:"issuerkeyhashsha512,omitempty" gorm:"UNIQUE_INDEX"`
}

// Profile struct
type Profile struct {
	gorm.Model
	Name             string `json:"name" gorm:"UNIQUE"`
	Ca               CA     `json:"ca"`
	CaID             uint
	CaName           string                  `json:"caname"`
	Validity         int                     `json:"validity"`
	KeyType          Type                    `json:"keytype"`
	KeySize          int                     `json:"keysize"`
	Digest           x509.SignatureAlgorithm `json:"digest"`
	KeyUsage         string                  `json:"keyusage,omitempty"`
	ExtendedKeyUsage string                  `json:"extendedkeyusage,omitempty"`
	P12SmtpServer    string                  `json:"p12smtpserver"`
	P12MailPassword  int                     `json:"p12mailpassword"`
	P12MailSubject   string                  `json:"p12mailsubject"`
	P12MailFrom      string                  `json:"p12mailfrom"`
	P12MailHeader    string                  `json:"p12mailheader"`
	P12MailFooter    string                  `json:"p12mailfooter"`
}

// Cert struct
type Cert struct {
	gorm.Model
	Cn                   string  `json:"cn"  gorm:"UNIQUE"`
	Mail                 string  `json:"mail"`
	StreetAddress        string  `json:"street,omitempty"`
	Organisation         string  `json:"organisation,omitempty"`
	Country              string  `json:"country,omitempty"`
	State                string  `json:"state,omitempty"`
	Locality             string  `json:"locality,omitempty"`
	PostalCode           string  `json:"postalcode,omitempty"`
	PrivateKey           string  `json:"privatekey,omitempty" gorm:"type:longtext"`
	PubKey               string  `json:"publickey,omitempty" gorm:"type:longtext"`
	ProfileName          string  `json:"profilename,omitempty"`
	Profile              Profile `json:"profile,omitempty"`
	ProfileID            uint
	ValidUntil           time.Time
	Date                 time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	Revoked              string    `json:"revoked,omitempty"`
	CRLReason            string    `json:"crlreason,omitempty"`
	UserIssuerHashmd5    string    `json:"userissuerhashmd5,omitempty" gorm:"UNIQUE_INDEX"`
	UserIssuerHashsha1   string    `json:"userissuerhashsha1,omitempty" gorm:"UNIQUE_INDEX"`
	UserIssuerHashsha256 string    `json:"userissuerhashsha256,omitempty" gorm:"UNIQUE_INDEX"`
	UserIssuerHashsha512 string    `json:"userissuerhashsha512,omitempty" gorm:"UNIQUE_INDEX"`
}

// curl -H "Content-Type: application/json" -d '{"cn":"YZaymCA","mail":"zaym@inverse.ca","organisation": "inverse","country": "CA","state": "QC", "locality": "Montreal", "streetaddress": "7000 avenue du parc", "postalcode": "H3N 1X1", "keytype": 1, "keysize": 2048, "Digest": 6, "days": 3650, "extendedkeyusage": "1|2", "keyusage": "1|32"}' http://127.0.0.1:12345/api/v1/pki/newca
func (c CA) new(pfpki *Handler) error {

	ca := &x509.Certificate{
		// Manage Serial Number
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization:  []string{c.Organisation},
			Country:       []string{c.Country},
			Province:      []string{c.State},
			Locality:      []string{c.Locality},
			StreetAddress: []string{c.StreetAddress},
			PostalCode:    []string{c.PostalCode},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, c.Days),
		IsCA:                  true,
		SignatureAlgorithm:    c.Digest,
		ExtKeyUsage:           extkeyusage(strings.Split(c.ExtendedKeyUsage, "|")),
		KeyUsage:              x509.KeyUsage(keyusage(strings.Split(c.KeyUsage, "|"))),
		BasicConstraintsValid: true,
		EmailAddresses:        []string{c.Mail},
	}

	keyOut, pub, key, err := GenerateKey(c.KeyType, c.KeySize)

	if err != nil {
		return err
	}
	var caBytes []byte

	switch c.KeyType {
	case KEY_RSA:
		caBytes, err = x509.CreateCertificate(rand.Reader, ca, ca, pub, key.(*rsa.PrivateKey))
	case KEY_ECDSA:
		caBytes, err = x509.CreateCertificate(rand.Reader, ca, ca, pub, key.(*ecdsa.PrivateKey))
	case KEY_DSA:
		caBytes, err = x509.CreateCertificate(rand.Reader, ca, ca, pub, key.(*dsa.PrivateKey))
	}
	if err != nil {
		return err
	}

	cert := new(bytes.Buffer)

	// Public key
	pem.Encode(cert, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes})

	pfpki.DB.AutoMigrate(&CA{})

	if err := pfpki.DB.Create(&CA{Cn: c.Cn, Mail: c.Mail, Organisation: c.Organisation, Country: c.Country, State: c.State, Locality: c.Locality, StreetAddress: c.StreetAddress, PostalCode: c.PostalCode, KeyType: c.KeyType, KeySize: c.KeySize, Digest: c.Digest, KeyUsage: c.KeyUsage, ExtendedKeyUsage: c.ExtendedKeyUsage, Days: c.Days, CaKey: keyOut.String(), CaCert: cert.String(), IssuerKeyHashmd5: c.IssuerKeyHashmd5, IssuerKeyHashsha1: c.IssuerKeyHashsha1, IssuerKeyHashsha256: c.IssuerKeyHashsha256, IssuerKeyHashsha512: c.IssuerKeyHashsha512}).Error; err != nil {
		return err
	}
	return nil
}

// func (c *CA) save() {
// 	// Public key
// 	certOut, err := os.Create("ca.crt")
// 	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: ca_b})
// 	certOut.Close()

// 	// Private key
// 	keyOut, err := os.OpenFile("ca.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
// 	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
// 	keyOut.Close()

// }

// curl -H "Content-Type: application/json" -d '{"name":"ZaymProfile","caname":"boby","validity": 365,"keytype": 1,"keysize": 2048, "digest": 6, "keyusage": "", "extendedkeyusage": "", "p12smtpserver": "10.0.0.6", "p12mailpassword": 1, "p12mailsubject": "New certificate", "P12MailFrom": "zaym@inverse.ca", "days": 365}' http://127.0.0.1:12345/api/v1/pki/newprofile
func (p Profile) new(pfpki *Handler) error {

	switch p.KeyType {
	case KEY_RSA:
		if p.KeySize < 2048 {
			return errors.New("invalid private key size, should be at least 2048")
		}
	case KEY_ECDSA:
		if !(p.KeySize == 256 || p.KeySize == 384 || p.KeySize == 521) {
			return errors.New("invalid private key size, should be 256 or 384 or 521")
		}
	case KEY_DSA:
		if !(p.KeySize == 1024 || p.KeySize == 2048 || p.KeySize == 3072) {
			return errors.New("invalid private key size, should be 1024 or 2048 or 3072")
		}
	default:
		return errors.New("KeyType unsupported")

	}
	// Create the table on the fly.
	pfpki.DB.AutoMigrate(&Profile{})
	var ca CA
	if CaDB := pfpki.DB.Where("Cn = ?", p.CaName).Find(&ca); CaDB.Error != nil {
		return CaDB.Error
	}

	if err := pfpki.DB.Create(&Profile{Name: p.Name, Ca: ca, CaName: p.CaName, Validity: p.Validity, KeyType: p.KeyType, KeySize: p.KeySize, Digest: p.Digest, KeyUsage: p.KeyUsage, ExtendedKeyUsage: p.ExtendedKeyUsage, P12SmtpServer: p.P12SmtpServer, P12MailPassword: p.P12MailPassword, P12MailSubject: p.P12MailSubject, P12MailFrom: p.P12MailFrom, P12MailHeader: p.P12MailHeader, P12MailFooter: p.P12MailFooter}).Error; err != nil {
		return err
	}
	return nil
}

// curl -H "Content-Type: application/json" -d '{"cn":"ZaymCert","mail":"zaim@inverse.ca","street": "7000 parc avenue","organisation": "inverse", "country": "zaymland", "state": "me", "locality": "zaymtown", "postalcode": "H3N 1X1", "profilename": "ZaymProfile"}' http://127.0.0.1:12345/api/v1/pki/newcert
func (c Cert) new(pfpki *Handler) error {

	pfpki.DB.AutoMigrate(&Cert{})

	// Find the profile
	var prof Profile
	if profDB := pfpki.DB.Where("Name = ?", c.ProfileName).Find(&prof); profDB.Error != nil {
		return profDB.Error
	}

	// Find the CA
	var ca CA
	if CaDB := pfpki.DB.Where("Cn = ?", prof.CaName).Find(&ca); CaDB.Error != nil {
		return CaDB.Error
	}
	// Load the certificates from the database
	catls, err := tls.X509KeyPair([]byte(ca.CaCert), []byte(ca.CaKey))
	if err != nil {
		return err
	}
	cacert, err := x509.ParseCertificate(catls.Certificate[0])
	if err != nil {
		return err
	}

	// Prepare certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization:  []string{c.Organisation},
			Country:       []string{c.Country},
			Province:      []string{c.State},
			Locality:      []string{c.Locality},
			StreetAddress: []string{c.StreetAddress},
			PostalCode:    []string{c.PostalCode},
		},
		NotBefore:      time.Now(),
		NotAfter:       time.Now().AddDate(0, 0, prof.Validity),
		SubjectKeyId:   []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:    extkeyusage(strings.Split(prof.ExtendedKeyUsage, "|")),
		KeyUsage:       x509.KeyUsage(keyusage(strings.Split(prof.KeyUsage, "|"))),
		EmailAddresses: []string{c.Mail},
	}

	keyOut, pub, _, err := GenerateKey(prof.KeyType, prof.KeySize)

	if err != nil {
		return err
	}
	// Sign the certificate
	certByte, err := x509.CreateCertificate(rand.Reader, cert, cacert, pub, catls.PrivateKey)

	certBuff := new(bytes.Buffer)

	// Public key
	pem.Encode(certBuff, &pem.Block{Type: "CERTIFICATE", Bytes: certByte})

	if err := pfpki.DB.Create(&Cert{Cn: c.Cn, Mail: c.Mail, StreetAddress: c.StreetAddress, Organisation: c.Organisation, Country: c.Country, Profile: prof, PrivateKey: keyOut.String(), PubKey: certBuff.String(), ValidUntil: cert.NotAfter}).Error; err != nil {
		return err
	}
	return nil
}