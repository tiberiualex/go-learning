package main

func main() {
	cards := newDeck()
	cards.saveToFile("my_cards.txt")
	cards.shuffle()
	cards.print()
}
