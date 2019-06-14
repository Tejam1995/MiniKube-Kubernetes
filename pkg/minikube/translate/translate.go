/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package translate

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry-attic/jibber_jabber"
	"github.com/golang/glog"
	"golang.org/x/text/language"
)

var (
	// preferredLanguage is the default language messages will be output in
	preferredLanguage = language.AmericanEnglish
	// our default language
	defaultLanguage = language.AmericanEnglish

	// Translations is a translation map from strings that can be output to console
	// to its translation in the user's system locale.
	Translations map[string]interface{}
)

// Translate translates the given string to the supplied language.
func Translate(s string) string {
	if preferredLanguage == defaultLanguage {
		return s
	}

	if len(Translations) == 0 {
		return s
	}

	if translation, ok := Translations[s]; ok {
		t := translation.(string)
		if len(t) > 0 && t != " " {
			return t
		}
	}

	return s
}

// DetermineLocale finds the system locale and sets the preferred language
// for output appropriately.
func DetermineLocale() {
	locale, err := jibber_jabber.DetectIETF()
	if err != nil {
		glog.Warningf("Getting system locale failed: %s", err)
		locale = ""
	}
	err = SetPreferredLanguage(locale)
	if err != nil {
		glog.Warningf("Setting locale failed: %s", err)
		preferredLanguage = defaultLanguage
	}

	// Load translations for preferred language into memory.
	if preferredLanguage != defaultLanguage {
		translationFile := "pkg/minikube/translate/translations/" + preferredLanguage.String() + ".json"
		t, err := ioutil.ReadFile(translationFile)
		if err != nil {
			glog.Infof("Failed to load transalation file for %s: %s", preferredLanguage.String(), err)
			return
		}

		err = json.Unmarshal(t, &Translations)
		if err != nil {
			glog.Infof("Failed to populate translation map: %s", err)
		}
	}

}

// setPreferredLanguageTag configures which language future messages should use.
func setPreferredLanguageTag(l language.Tag) {
	glog.Infof("Setting Language to %s ...", l)
	preferredLanguage = l
}

// SetPreferredLanguage configures which language future messages should use, based on a LANG string.
func SetPreferredLanguage(s string) error {
	// "C" is commonly used to denote a neutral POSIX locale. See http://pubs.opengroup.org/onlinepubs/009695399/basedefs/xbd_chap07.html#tag_07_02
	if s == "" || s == "C" {
		setPreferredLanguageTag(defaultLanguage)
		return nil
	}
	// Handles "de_DE" or "de_DE.utf8"
	// We don't process encodings, since Rob Pike invented utf8 and we're mostly stuck with it.
	parts := strings.Split(s, ".")
	l, err := language.Parse(parts[0])
	if err != nil {
		return err
	}
	setPreferredLanguageTag(l)
	return nil
}

// GetPreferredLanguage returns the preferred language tag.
func GetPreferredLanguage() language.Tag {
	return preferredLanguage
}
