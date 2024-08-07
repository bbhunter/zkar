package main

import (
	"encoding/base64"
	"fmt"
	"github.com/phith0n/zkar/serz"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	var app = cli.App{
		Name:  "zkar",
		Usage: "A Java serz tool",
		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "(WIP) generate Java serialization attack payloads",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "output",
						Usage:    "output file path",
						Aliases:  []string{"o"},
						Required: false,
						Value:    "",
					},
					&cli.BoolFlag{
						Name:     "list",
						Usage:    "list all available gadgets",
						Aliases:  []string{"l"},
						Required: false,
						Value:    false,
					},
				},
				Action: func(context *cli.Context) error {
					return fmt.Errorf("payloads generation feature is working in progress")
				},
			},
			{
				Name:  "dump",
				Usage: "parse the Java serz streams and dump the struct",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "serz data filepath",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "base64",
						Aliases:  []string{"B"},
						Usage:    "serz data as Base64 format string",
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "golang",
						Usage:    "dump the Go language based struct instead of human readable information",
						Required: false,
						Value:    false,
					},
					&cli.BoolFlag{
						Name: "jdk8u20",
						Usage: "This payload is a JDK8u20 payload generated by " +
							"<https://github.com/pwntester/JRE8u20_RCE_Gadget>",
						Required: false,
						Value:    false,
					},
				},
				Action: func(context *cli.Context) error {
					var filename = context.String("file")
					var b64data = context.String("base64")
					var data []byte
					var err error
					if (filename == "" && b64data == "") || (filename != "" && b64data != "") {
						return fmt.Errorf("one \"file\" or \"base64\" flag must be specified, and not both")
					}

					if filename != "" {
						data, err = os.ReadFile(filename)
					} else {
						data, err = base64.StdEncoding.DecodeString(b64data)
					}

					if err != nil {
						return err
					}

					var obj *serz.Serialization
					if context.Bool("jdk8u20") {
						obj, err = serz.FromJDK8u20Bytes(data)
					} else {
						obj, err = serz.FromBytes(data)
					}
					if err != nil {
						log.Fatalln(err)
						return nil
					}

					if context.Bool("golang") {
						serz.DumpToGoStruct(obj)
					} else {
						fmt.Println(obj.ToString())
					}

					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[error] %v\n", err.Error())
	}
}
