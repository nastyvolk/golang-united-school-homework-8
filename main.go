package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

var operations = []string{"add", "list", "findById", "remove"}

func Perform(args Arguments, writer io.Writer) error {
	if args["operation"] == "" {
		return errors.New("-operation flag has to be specified")
	}
	if !isOperationAllowed(args["operation"]) {
		return errors.New("Operation abcd not allowed!")
	}
	if args["fileName"] == "" {
		return errors.New("-fileName flag has to be specified")
	}
	f, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	users, err := readUsersFromFile(f)
	if err != nil {
		return err
	}
	switch args["operation"] {
	case "add":
		if args["item"] == "" {
			return errors.New("-item flag has to be specified")
		}

		u := &User{}
		err := json.Unmarshal([]byte(args["item"]), u)
		if err != nil {
			return err
		}

		userFromFile, _ := findUserAndPositionById(users, u.Id)

		if userFromFile != nil {
			writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", userFromFile.Id)))
		}

		users = append(users, *u)
		byteUsers, err := json.Marshal(users)
		if err != nil {
			return err
		}
		f.Write(byteUsers)
	case "list":
		b, err := json.Marshal(users)
		if err != nil {
			return err
		}
		writer.Write(b)
	case "findById":
		if args["id"] == "" {
			return errors.New("-id flag has to be specified")
		}
		userFromFile, _ := findUserAndPositionById(users, args["id"])
		if userFromFile != nil {
			byteUser, err := json.Marshal(userFromFile)
			if err != nil {
				return err
			}
			writer.Write(byteUser)
		}
	case "remove":
		if args["id"] == "" {
			return errors.New("-id flag has to be specified")
		}
		userFromFile, i := findUserAndPositionById(users, args["id"])
		if userFromFile != nil {
			users = append(users[:i], users[i+1:]...)
		} else {
			writer.Write([]byte(fmt.Sprintf("Item with id %s not found", args["id"])))
		}
		f.Truncate(0)
		f.Seek(0, 0)
		u, err := json.Marshal(users)
		if err != nil {
			return err
		}
		f.Write(u)
	}
	return nil
}

func isOperationAllowed(operation string) bool {
	for _, v := range operations {
		if operation == v {
			return true
		}
	}
	return false
}

func findUserAndPositionById(users []User, id string) (*User, int) {
	for i, u := range users {
		if u.Id == id {
			return &u, i
		}
	}
	return nil, 0
}

func readUsersFromFile(f *os.File) ([]User, error) {
	var users []User
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(b) != 0 {
		err = json.Unmarshal(b, &users)
		if err != nil {
			return nil, err
		}
	}
	return users, nil
}

func parseArgs() map[string]string {
	id := flag.String("id", "", "id")
	o := flag.String("operation", "", "operation")
	i := flag.String("item", "", "item")
	f := flag.String("fileName", "", "fileName")
	flag.Parse()
	args := Arguments{
		"id":        *id,
		"operation": *o,
		"item":      *i,
		"fileName":  *f,
	}
	return args
}

func main() {
	_, err := os.Stat("users.json")
	if errors.Is(err, os.ErrNotExist) {
		f, err := os.Create("users.json")
		if err != nil {
			panic(err)
		}
		defer f.Close()

	}
	err = Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}
