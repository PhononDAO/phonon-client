package util

import (
	"crypto/rand"
	"fmt"

	"github.com/manifoldco/promptui"
)

func RandomKey(length int) []byte {
	key := make([]byte, length)
	rand.Read(key)
	return key
}

//Create an interactive prompt for user's pin
func PinPrompt() (string, error) {
	//Prompt user for pin
	prompt := promptui.Prompt{
		Label: "Pin",
		Mask:  '*',
	}
	fmt.Println("Please enter 6 digit pin:")
	result, err := prompt.Run()
	if err != nil {
		fmt.Println("prompt failed: err: ", err)
		return "", err
	}
	return result, nil
}
