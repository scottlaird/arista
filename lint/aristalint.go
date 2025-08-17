package main

// Note that this is a linter/schema checker for the Arista model
// definitions in this repository, *not* a generic Arista config
// linter.

import (
        "flag"
        "fmt"
        "os"
        "time"

        "gopkg.in/yaml.v3"
)

type (
        Definition struct {
                Name                     string    `yaml:"name"`
                Models                   []Model   `yaml:"models"`
                LastEOSRevisionSupported string    `yaml:"last-eos-revision-supported"`
                PDFDatasheetURL          string    `yaml:"pdf-datasheet-url"`
                EndOfSaleAnnounced       bool      `yaml:"end-of-sale-announced"`
                EndOfSaleDate            time.Time `yaml:"end-of-sale-date"`
                EndOfSaleURL             string    `yaml:"end-of-sale-url"`
                EndOfSupportDate         time.Time `yaml:"end-of-support-date"`
                EndOfSupportURL          string    `yaml:"end-of-support-url"`
                Notes                    []string  `yaml:"notes"`
        }

        Model struct {
                Name         string   `yaml:"name"`
                TypicalWatts float64  `yaml:"typical-watts"`
                MaxWatts     float64  `yaml:"max-watts"`
                RackUnits    int      `yaml:"rack-units"`
                CPUCores     int      `yaml:"cpu-cores"` // Should probably include the CPU type, but that's harder to find.
                CPURamGB     int      `yaml:"cpu-ram-gb"`
                CPUFlashGB   int      `yaml:"cpu-flash-gb"`
                Ports        []Port   `yaml:"ports"`
                SwitchChip   string   `yaml:"switch-chip"`
                Notes        []string `yaml:"notes"`
        }

        Port struct {
                Type  string `yaml:"type"`
                Count int    `yaml:"count"`
                Note  string `yaml:"note"`
        }
)

var (
        portTypes = listToMap([]string{
                "1000base-T",
                "10Gbase-T",
                "SFP+",
                "SFP28",
                "QSFP+",
                "QSFP28",
        })

        switchChips = listToMap([]string{
                "Trident2",
                "Trident2+",
        })
)

// Syntactic sugar for the lists of supported interface and switch chips, above.
func listToMap(list []string) map[string]bool {
        m := make(map[string]bool)
        for _, i := range list {
                m[i] = true
        }
        return m
}

func main() {
        flag.Parse()
        failed := false

        for _, filename := range flag.Args() {
                err := verifyFile(filename)
                if err != nil {
                        fmt.Fprintf(os.Stderr, "%s: FAIL: %v\n", filename, err)
                        failed = true
                } else {
                        fmt.Fprintf(os.Stderr, "%s: ok\n", filename)
                }
        }

        if failed == true {
                os.Exit(1)
        }
}

func verifyFile(filename string) error {
        f, err := os.Open(filename)
        if err != nil {
                return err
        }
        defer f.Close()

        decoder := yaml.NewDecoder(f)
        decoder.KnownFields(true)

        deviceDefinition := Definition{}
        err = decoder.Decode(&deviceDefinition)
        if err != nil {
                return err
        }

        for _, model := range deviceDefinition.Models {
                err = verifyModel(model)
                if err != nil {
                        return fmt.Errorf("device %q: %v", model.Name, err)
                }
        }

        return nil
}

func verifyModel(model Model) error {
        for _, port := range model.Ports {
                if portTypes[port.Type] != true {
                        return fmt.Errorf("port type %q unknown", port.Type)
                }
        }
        return nil
}
