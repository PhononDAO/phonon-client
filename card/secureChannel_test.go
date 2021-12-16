package card

import (
	"github.com/GridPlus/phonon-client/cert"
	"github.com/GridPlus/phonon-client/util"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestSecureChannelRawEncryptDecrypt(t *testing.T) {
	testData := "Encrypt Me"

	sc := SecureChannel{}

	iv := util.RandomKey(16)
	encKey := util.RandomKey(32)
	macKey := util.RandomKey(32)

	sc.Init(iv, encKey, macKey)

	log.Debugf("iv: % X", sc.iv)
	ciphertext, err := sc.Encrypt([]byte(testData))
	if err != nil {
		t.Error("error encrypting data: ", err)
		return
	}
	log.Debugf("ciphertext: % X", ciphertext)
	log.Debugf("iv: % X", sc.iv)
	resultData, err := sc.DecryptDirect(ciphertext, iv)
	if err != nil {
		t.Error("error decrypting ciphertext: ", err)
		return
	}
	log.Debugf("resultData: % X", resultData)
	log.Debugf("iv: % X", sc.iv)

	resultString := string(resultData)
	if string(resultData) != testData {
		t.Errorf("decrypted result: %v did not equal testData: %v", resultString, testData)
		return
	}
}

func TestSecureChannelBackAndForth(t *testing.T) {
	// testData := "Encrypt Me"
	testData := "This is a message that is greater than 16 bytes long"

	senderSC := SecureChannel{}
	receiverSC := SecureChannel{}

	iv := util.RandomKey(16)
	encKey := util.RandomKey(32)
	macKey := util.RandomKey(32)

	senderSC.Init(iv, encKey, macKey)
	receiverSC.Init(iv, encKey, macKey)
	log.Debugf("sender iv: % X", senderSC.iv)

	ciphertext, err := senderSC.Encrypt([]byte(testData))
	if err != nil {
		t.Error("error encrypting data: ", err)
		return
	}
	log.Debugf("sender iv: % X", senderSC.iv)
	log.Debugf("receiver iv: % X", receiverSC.iv)
	resultData, err := receiverSC.Decrypt(ciphertext)
	if err != nil {
		t.Error("receiver error decrypting, ", err)
	}
	log.Debugf("receiver iv: % X", receiverSC.iv)

	testResponse := "Encrypt a second time"
	responseCiphertext, err := receiverSC.Encrypt([]byte(testResponse))
	if err != nil {
		t.Error("receiver error encrypting: ", err)
		return
	}

	responseResultData, err := senderSC.Decrypt(responseCiphertext)
	if err != nil {
		t.Error("sender error decrypting response: ", err)
		return
	}

	if string(resultData) != testData {
		t.Error("first encrypted message did not match")
		return
	}
	if string(responseResultData) != testResponse {
		t.Error("second encrypted message did not match")
		return
	}
}

func TestMockCardPairing(t *testing.T) {
	senderCard, err := NewMockCard(false, false)
	if err != nil {
		t.Error(err)
		return
	}
	senderCard.InstallCertificate(cert.SignWithDemoKey)
	receiverCard, err := NewMockCard(false, false)
	if err != nil {
		t.Error(err)
		return
	}
	receiverCard.InstallCertificate(cert.SignWithDemoKey)

	initCardPairingData, err := senderCard.InitCardPairing(receiverCard.IdentityCert)
	if err != nil {
		t.Error("error in initCardPairing: ", err)
		return
	}
	cardPairData, err := receiverCard.CardPair(initCardPairingData)
	if err != nil {
		t.Error("error in card pair: ", err)
		return
	}
	cardPairData2, err := senderCard.CardPair2(cardPairData)
	if err != nil {
		t.Error("error in card pair 2: ", err)
		return
	}
	err = receiverCard.FinalizeCardPair(cardPairData2)
	if err != nil {
		t.Error("error in finalize card pair: ", err)
		return
	}

	testData := "Encrypt this"
	ciphertext, err := senderCard.sc.Encrypt([]byte(testData))
	if err != nil {
		t.Error("error encrypting initial data: ", err)
		return
	}

	resultData, err := receiverCard.sc.Decrypt(ciphertext)
	if err != nil {
		t.Error("error decrypting data: ", err)
		return
	}
	if testData != string(resultData) {
		t.Error("testData did not equal resultData")
		return
	}
	replyData := "More Encrypted Secret Info"
	replyCiphertext, err := receiverCard.sc.Encrypt([]byte(replyData))
	if err != nil {
		t.Error("error encrypting reply data: ", err)
		return
	}
	replyResultData, err := senderCard.sc.Decrypt(replyCiphertext)
	if err != nil {
		t.Error("error decrypting reply data: ", err)
		return
	}
	if replyData != string(replyResultData) {
		t.Error("replyData did not equal replyResultData: ", err)
		return
	}
}
