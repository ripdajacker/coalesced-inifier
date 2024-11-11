package write

import (
	"bytes"
	"coalesced-inifier/gowenc"
	"fmt"
	"gopkg.in/ini.v1"
	"path/filepath"
	"strings"
)

func AppendIni(w *bytes.Buffer, realBase string, entryPath string, prefix string) error {
	opts := ini.LoadOptions{AllowShadows: true, AllowDuplicateShadowValues: true, UnescapeValueDoubleQuotes: true, AllowNestedValues: true, IgnoreInlineComment: false, ChildSectionDelimiter: "UNLIKELY DELIMITER !!! -/2"}
	iniFile, err := ini.LoadSources(opts, entryPath)

	if err != nil {
		panic(err)
	}

	rel, err := filepath.Rel(realBase, entryPath)

	coalescedName := fmt.Sprintf("%s%s", prefix, rel)
	coalescedName = strings.Replace(coalescedName, "/", "\\", -1)

	fmt.Printf("Packing '%s' as '%s'...", entryPath, coalescedName)

	err = gowenc.WriteUtf16InvLen(w, coalescedName)

	if err != nil {
		return err
	}

	sections := iniFile.Sections()

	sectionCount := 0
	for i := 0; i < len(sections); i++ {
		section := sections[i]

		if section.Name() != "DEFAULT" {
			sectionCount++
		}
	}

	err = gowenc.WriteUint32BE(w, sectionCount)
	if err != nil {
		return err
	}

	for i := 0; i < len(sections); i++ {
		section := sections[i]

		if section.Name() != "DEFAULT" {
			// println("  Section", section.Name(), len(section.Keys()))
			err = gowenc.WriteUtf16InvLen(w, section.Name())
			if err != nil {
				return err
			}

			keys := section.Keys()

			keyCount := 0
			for _, key := range keys {
				shadows := key.ValueWithShadows
				if len(shadows()) == 0 {
					keyCount++
				} else {
					keyCount += len(shadows())
				}
			}

			gowenc.WriteUint32BE(w, keyCount)
			for j := 0; j < len(keys); j++ {
				key := keys[j]
				keyName := key.Name()
				if strings.HasPrefix(keyName, "IGNORED_SC_") {
					keyName = strings.Replace(keyName, "IGNORED_SC_", ";", 1)
				} else if strings.HasPrefix(keyName, "IGNORED_SH_") {
					keyName = strings.Replace(keyName, "IGNORED_SH_", "#", 1)
				}

				shadows := key.ValueWithShadows()
				if len(shadows) == 0 {
					gowenc.WriteUtf16InvLen(w, keyName)
					gowenc.WriteUint32BE(w, 0)
				} else {
					for _, value := range shadows {
						gowenc.WriteUtf16InvLen(w, keyName)
						gowenc.WriteUtf16InvLen(w, value)
					}
				}
			}
		}
	}

	return nil
}
