package derived

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prysmaticlabs/prysm/shared/bls"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/validator/keymanager"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
)

// ExtractKeystores retrieves the secret keys for specified public keys
// in the function input, encrypts them using the specified password,
// and returns their respective EIP-2335 keystores.
func (dr *Keymanager) ExtractKeystores(
	_ context.Context, publicKeys []bls.PublicKey, password string,
) ([]*keymanager.Keystore, error) {
	encryptor := keystorev4.New()
	keystores := make([]*keymanager.Keystore, len(publicKeys))
	lock.RLock()
	defer lock.RUnlock()
	for i, pk := range publicKeys {
		pubKeyBytes := pk.Marshal()
		secretKey, ok := secretKeysCache[bytesutil.ToBytes48(pubKeyBytes)]
		if !ok {
			return nil, fmt.Errorf(
				"secret key for public key %#x not found in cache",
				pubKeyBytes,
			)
		}
		cryptoFields, err := encryptor.Encrypt(secretKey.Marshal(), password)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"could not encrypt secret key for public key %#x",
				pubKeyBytes,
			)
		}
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		keystores[i] = &keymanager.Keystore{
			Crypto:  cryptoFields,
			ID:      id.String(),
			Pubkey:  fmt.Sprintf("%x", pubKeyBytes),
			Version: encryptor.Version(),
			Name:    encryptor.Name(),
		}
	}
	return keystores, nil
}