# Phonon-client

### Release Conventions
See https://github.com/marketplace/actions/autotag

All commits to master will automatically create a tag with a bumped patch version, and upload the generated binaries under a release corresponding to the created tag.

Commit with #minor or #major to bump minor and major version numbers rather than the patch number. Commits that reference an issue with the "enhancement" label will also automatically bump the minor version.



# Use Cases
Proposed use cases for the phonon network, which the client API should ultimately be able to support

### Make a Payment
For example, buy a cup of coffee using phonons. Simple, one-time transfer of specific amount of phonons from one card to another.

### Micropayments
Make a series of streaming micropayments. Triggered over time or by a specific event?

### Atomic Swaps
Exchange phonons between two parties at an agreed upon exchange rate

### Changemaker
Make change with a "changemaker" entity, who possesses various denominations of change, and allows users to swap phonons
for larger or smaller denominations at their convenience, charging a nominal fee.
Make the change you want to see in the world.


## API Draft

Description of  operations that a phonon client would need to support in order to build a functioning system that can support the above operations.

```
detect() cardID {
  //detect that a card has been connected to the reader and is available to the system to transact
}
```

```
startTransfer(cardA cardID, cardB cardID) error {
  //Create a secure channel between two cards so that phonons can be transferred securely
}
```

```
send(sourceCard cardID, targetCard cardID, amount float, currencyType int) error {
  //sends a specific amount, internally calculating a valid collection of phonons to complete the transfer
  //Simplest, low level send that other more complex exchange methods can be built on top of
  //returns an error if the transaction cannot be completed due to lack of funds or impossible amount to satisfy with existing phonons
}
```

```
deposit(targetCard cardID, amount float, currencyType int, denominations denominationStrategy) error {
  //Deposit an amount of phonons of a certain currency
  //denominations can hold a list or function specifying what denominations the phonons should be held in to make up the requested amount
  //error can be returned if the blockchain transaction is unsuccessful
  //on success the phonons will be deposited on the card
}
```

```
withdraw(targetCard cardID, amount float, currencyType int, address string, denominationPreference denominationStrategy) error {
  //Withdraw an amount of phonons of a certain currency from the card to the chain at the specified address
  //denominationPreference can be set to specify which phonons to liquidate if this is a partial withdrawal
  //(for example largest phonons first, smallest first, etc.)
  //errors can be returned if there is an issue completing the withdrawal
}
```

Other higher level payment primitives built on top of send...
```
swap() //atomically swapping different asset types
autopay() //regularly send a specified amount on an interval
makeChange() //Send in two directions based on change calculation
```

