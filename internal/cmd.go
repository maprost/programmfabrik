package internal

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os/exec"
)

type XmlTable struct {
	Name string `xml:"name,attr"`
	//G0   string           `xml:"g0,attr"`
	//G1   string           `xml:"g1,attr"`
	//G2   string           `xml:"g2,attr"`
	//Desc []XmlDescription `xml:"desc"`
	Tag []XmlTag `xml:"tag"`
}

type XmlTag struct {
	Name     string           `xml:"name,attr"`
	Type     string           `xml:"type,attr"`
	Writable string           `xml:"writable,attr"`
	Desc     []XmlDescription `xml:"desc"`
}

type XmlDescription struct {
	Lang  string `xml:"lang,attr"`
	Value string `xml:",chardata"`
}

func callExiftool(c chan JsonTable, done chan struct{}, filter string) error {
	cmd := exec.Command("exiftool", "-listx")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Split(bufio.ScanWords)
	var tableStr string
	addTableStr := false
	for scanner.Scan() {
		m := scanner.Text()
		if m == "<table" {
			tableStr = m + " "
			addTableStr = true
			fmt.Printf(".")

		} else if m == "</table>" {
			tableStr += m
			table, err := convertTable(tableStr)
			if err != nil {
				return err
			}
			if filter == "" {
				c <- table
			} else {
				for _, tag := range table.Tags {
					if tag.Path == filter {
						fmt.Printf("!")
						c <- table
						break
					}
				}
			}

			tableStr = ""
			addTableStr = false
		} else if addTableStr {
			tableStr += m + " "
		}
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	fmt.Printf(" done\n")
	return nil
}

func convertTable(txt string) (JsonTable, error) {
	var xmlTable XmlTable
	var jsonTable JsonTable

	err := xml.Unmarshal([]byte(txt), &xmlTable)
	if err != nil {
		return jsonTable, err
	}

	for _, tag := range xmlTable.Tag {
		jsonDesc := make(map[string]string)
		for _, desc := range tag.Desc {
			jsonDesc[desc.Lang] = desc.Value
		}

		jsonTable.Tags = append(jsonTable.Tags, JsonTag{
			Type:        tag.Type,
			Writeable:   tag.Writable == "true",
			Path:        xmlTable.Name + ":" + tag.Name,
			Group:       xmlTable.Name,
			Description: jsonDesc,
		})
	}

	return jsonTable, nil
}
