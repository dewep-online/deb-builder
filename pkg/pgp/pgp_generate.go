package pgp

import (
	"bytes"
	"io"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

const (
	PublicFilename  = "public.pgp"
	PrivateFilename = "private.pgp"
)

var (
	keyHeaders = map[string]string{
		"Version":      "Golang OpenPGP",
		"Generated by": "github.com/dewep-online/deb-builder",
	}
)

func (v *PGP) Generate(out string, name, comment, email string) error {
	buf := &bytes.Buffer{}

	key, err := openpgp.NewEntity(name, comment, email, nil)
	if err != nil {
		return errors.Wrap(err, "generate entity")
	}

	if err := v.setup(key); err != nil {
		return errors.Wrap(err, "setup entity")
	}

	if err := v.genPrivateKey(key, buf); err != nil {
		return errors.Wrap(err, "generate private key")
	}

	if err := os.WriteFile(out+"/"+PrivateFilename, buf.Bytes(), 0644); err != nil {
		return errors.Wrap(err, "write private key")
	}

	buf.Reset()

	if err := v.genPublicKey(key, buf); err != nil {
		return errors.Wrap(err, "generate public key")
	}

	if err := os.WriteFile(out+"/"+PublicFilename, buf.Bytes(), 0600); err != nil {
		return errors.Wrap(err, "write public key")
	}

	return nil
}

func (v *PGP) genPrivateKey(key *openpgp.Entity, w io.Writer) error {
	enc, err := armor.Encode(w, openpgp.PrivateKeyType, keyHeaders)
	if err != nil {
		return errors.Wrap(err, "create OpenPGP Armor")
	}

	defer enc.Close() //nolint: errcheck

	if err := key.SerializePrivate(enc, nil); err != nil {
		return errors.Wrap(err, "serialize private key")
	}

	return nil
}

func (v *PGP) genPublicKey(key *openpgp.Entity, w io.Writer) error {
	enc, err := armor.Encode(w, openpgp.PublicKeyType, keyHeaders)
	if err != nil {
		return errors.Wrap(err, "create OpenPGP Armor")
	}

	defer enc.Close() //nolint: errcheck

	if err := key.Serialize(enc); err != nil {
		return errors.Wrap(err, "serialize public key")
	}

	return nil
}

func (v *PGP) setup(key *openpgp.Entity) error {
	// Sign all the identities
	for _, id := range key.Identities {
		id.SelfSignature.PreferredCompression = []uint8{1, 2, 3, 0}
		id.SelfSignature.PreferredHash = []uint8{2, 8, 10, 1, 3, 9, 11}
		id.SelfSignature.PreferredSymmetric = []uint8{9, 8, 7, 3, 2}

		if err := id.SelfSignature.SignUserId(id.UserId.Id, key.PrimaryKey, key.PrivateKey, nil); err != nil {
			return err
		}
	}

	return nil
}
