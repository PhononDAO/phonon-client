package tlv

import (
	"testing"

	"github.com/GridPlus/keycard-go/hexutils"
	"github.com/GridPlus/phonon-client/util"
	log "github.com/sirupsen/logrus"
)

//Duplicate tags needed for standalone TLV test here
const (
	TagSelectAppInfo           = 0xA4
	TagCardUID                 = 0x8F
	TagCardSecureChannelPubKey = 0x80
	TagAppVersion              = 0x02
)

func TestParseTLV(t *testing.T) {
	testSelectAPDU := hexutils.HexToBytes("A4 5F 8F 10 48 C5 33 14 41 93 28 CF DA 80 9D CB BB F2 AC A9 80 41 04 FC D8 88 73 52 BA 5D C1 F4 4E AA F9 BD A1 63 FB C0 66 1E B6 5C 78 9A FE 5F D8 14 D6 C1 66 EB CB B9 C7 C6 0E 65 66 AF 4E 3C 85 75 95 01 42 6F 53 85 E0 42 A3 37 47 B3 D9 33 2E DA 74 86 B0 5E 1F 02 02 00 01 03 01 00 8D 01 07 90 00")

	log.Debug("raw bytes: ")
	log.Debugf("% X", testSelectAPDU)
	collection, err := ParseTLVPacket(testSelectAPDU, TagSelectAppInfo)
	if err != nil {
		log.Error(err)
		return
	}
	appInfo, err := collection.FindTag(TagSelectAppInfo)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = ParseTLVPacket(appInfo)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = collection.FindTag(TagCardUID)
	if err != nil {
		log.Error(err)
		return
	}
	pubKey, err := collection.FindTag(TagCardSecureChannelPubKey)
	if err != nil {
		log.Error(err)
		return
	}

	ecdsaPubKey, err := util.ParseECCPubKey(pubKey)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("extracted pubkey: % X % X", ecdsaPubKey.X.Bytes(), ecdsaPubKey.Y.Bytes())

	appVersion, err := collection.FindTag(TagAppVersion)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("appVersion: % X", appVersion)
}
