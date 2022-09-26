package tlv

import (
	"encoding/hex"
	"fmt"
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

// Example with tag being 0x11
// A data length of 64 bytes will be 0x40 - so it would look like 0x1140
// A data length of 130 bytes will be 0x81 0x83 - so it would look like 0x118183
// A data length of 280 bytes will be 0x82 0x01 0x18 - so it would look like 0x11820118
func TestParseBERTLVPacket(t *testing.T) {
	cases := []string{
		// 80510000# - 1 TLV (no nesting)
		"90819030900202000080410468F84EC5BF0B5F4EB94ED720EF99FC1EFA4FECE6EA9B9138C3F4C07E169C0940DBF9128BD9ABAB58E8B8AEAA9F6313AEB9D17B05E37201FD68201F4AEB675903304502205AA91139AA469435F5F5EE3D18E333276C97F39792D02300D1DB89D89853EC2002210090AD60F5767B1A62FD91054B48DCBAE5CEA7B469349993A4C2A38698D92A3BE9912096DC0119296C3C2236E5146A72BCBD284CE68535727A43C2B784B378CB243C21",

		// 80530000# - 1 TLV (no nesting)
		"93473045022100C7DC911449F799FAD2F2522D2CDFF88B69A4BFFABC4D90A734454BAA63CE28AE0220128B78E3C1EA07F1B9CB84D38DA8BF6EC64F107EE73ED78275EEF9A6691BB7E1",

		// 80500000#
		"908191309102020000804104D2B04AE13EE4D991A94065009DC7C0C48620663FB8B5C8F3890FF53F00CA7D5148BFAD218922E076E30AB8DC1509C7CF1296772EEAC37F8C7A7180920CD908503046022100FD12CCA7BFA3139FF383B655B6018811042D2654DF5C2F56A00B9951A67824A602210080F8C336A5DF1205FA1B732512126B4C0370BB41E175F897363D5659BEB54CA0",

		// 80520000# - this has multiple TLVS?
		"912015FB87F3577AC358E209EED7316C6EDB96981E04B4ED7E5DF597677B0B90BC2B921058EAE71460719CF332C9206863BD50F093473045022100C67639FA671CCCFC91B8A0CBFFFF5553EFBDC143933170C7129C48A42B45E3BE02204B12629A498DF0974E6B9DA4F915DC0C64D4C07E9C167E2FFA354B2888B9F79C",
	}
	for i, c := range cases {
		fmt.Println("CASE", i)
		r, err := ParseTLVPacket(hexutils.HexToBytes(c), TagAppVersion)
		if err != nil {
			fmt.Println("ERR", i, err)
		}
		for k, v := range r {
			fmt.Println("## KEY", k, "VAL LEN", len(v[0]))
			fmt.Println("##", k, "VAL STR", hex.EncodeToString(v[0]))
		}
		fmt.Printf("########\n\n")
	}
}

// TODO:
// - test edgecases: 0 len TLV, overflowing TLV (>65535, or some other value)
// - test invalid TLV:
//    - TLV too short, VALUE is < LENGTH
//    - unexpected number of found tags
//    - expected tag not found
//    - more tags than expected
// EXTRAS:
// Add tests for other functions which were not covered earlier
