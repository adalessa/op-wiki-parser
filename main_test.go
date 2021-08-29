package main

import "testing"

func Test_it_parses(t *testing.T) {
	chapter, err := processChapter(238)

	if err != nil {
		t.Fatal("Got an error: ", err)
	}

	if chapter.Number != 238 {
		t.Errorf("Wrong chapter number got %d", chapter.Number)
	}

	err = pushToApi(*chapter)
	if err != nil {
		t.Fatal("Got an error: ", err)
	}

	t.Fail()
}
